package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/uswuth/vytora-backend/internal/config"
	"github.com/uswuth/vytora-backend/internal/database"
	"github.com/uswuth/vytora-backend/internal/entity/audit_trail"
	"github.com/uswuth/vytora-backend/internal/entity/category"
	"github.com/uswuth/vytora-backend/internal/entity/compliance_record"
	"github.com/uswuth/vytora-backend/internal/entity/contract"
	"github.com/uswuth/vytora-backend/internal/entity/report"
	"github.com/uswuth/vytora-backend/internal/entity/risk_assessment"
	"github.com/uswuth/vytora-backend/internal/entity/user"
	"github.com/uswuth/vytora-backend/internal/entity/vendor"
	"github.com/uswuth/vytora-backend/internal/entity/vendor_contact"

	"github.com/uswuth/vytora-backend/internal/graphql/directives"
	"github.com/uswuth/vytora-backend/internal/graphql/generated"
	"github.com/uswuth/vytora-backend/internal/graphql/resolver"
	graphqlmiddleware "github.com/uswuth/vytora-backend/internal/middleware/graphql"
	"github.com/uswuth/vytora-backend/internal/logger"
	"github.com/uswuth/vytora-backend/internal/services"
	startup "github.com/uswuth/vytora-backend/internal/startup"
)

func main() {
	startTime := time.Now()

	cfg, err := config.LoadWithDefaults()
	if err != nil {
		startup.ShowConfigError("config", err)
	}

	if err := database.Connect(cfg.DatabaseURL); err != nil {
		startup.ShowConfigError("db", err)
	}
	defer database.Close()

	var dbPoolStats string
	if database.Pool != nil {
		st := database.Pool.Stat()
		dbPoolStats = fmt.Sprintf("acquired=%d, idle=%d, total=%d", st.AcquiredConns(), st.IdleConns(), st.TotalConns())
	}

	// Repositories
	userRepo := user.NewRepository(database.Pool)
	vendorRepo := vendor.NewRepository(database.Pool)
	riskAssessmentRepo := risk_assessment.NewRepository(database.Pool)
	compRepo := compliance_record.NewRepository(database.Pool)
	contractRepo := contract.NewRepository(database.Pool)
	auditRepo := audit_trail.NewRepository(database.Pool)
	reportRepo := report.NewRepository(database.Pool)
	categoryRepo := category.NewRepository(database.Pool)
	contactRepo := vendor_contact.NewRepository(database.Pool)

	// Services
	jwtService := services.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryHours)
	seqService := services.NewSequenceService(database.Pool)
	auditLogger := audit_trail.NewLogger(auditRepo, seqService.NextCode)

	// Root resolver
	res := &resolver.Resolver{
		UserRepo:            userRepo,
		VendorRepo:          vendorRepo,
		RiskAssessmentRepo:  riskAssessmentRepo,
		ComplianceRepo:      compRepo,
		ContractRepo:        contractRepo,
		AuditRepo:           auditRepo,
		ReportRepo:          reportRepo,
		CategoryRepo:        categoryRepo,
		ContactRepo:         contactRepo,
		JWTService:          jwtService,
		SeqService:          seqService,
		AuditLogger:         auditLogger,
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: res,
		Directives: generated.DirectiveRoot{
			IsAuthenticated: directives.IsAuthenticated,
			HasPermission:   directives.HasPermission,
		},
	}))

	// HTTP router with production middleware stack
	r := chi.NewRouter()

	// Middleware: order matters — outer runs first
	r.Use(chimiddleware.RequestID)        // Unique request ID per request
	r.Use(logger.RequestLogger)           // Custom colored request logging
	r.Use(chimiddleware.Recoverer)        // Panic recovery — prevents crashes
	r.Use(chimiddleware.Timeout(30 * time.Second)) // Request timeout — prevents hanging
	r.Use(chimiddleware.Compress(5, "text/plain", "application/json", "application/graphql+json")) // Gzip compression
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(graphqlmiddleware.RateLimiter(cfg.RateLimitRequests, cfg.RateLimitInterval))
	r.Use(graphqlmiddleware.MetricsMiddleware)

	// Health & metrics
	r.Get("/healthz", graphqlmiddleware.HealthCheckHandler(cfg.HealthAllowedIPs))
	r.Get("/readyz", graphqlmiddleware.ReadinessHandler(cfg.HealthAllowedIPs, database.Pool))
	r.Get("/metrics", graphqlmiddleware.MetricsHandler())

	// GraphQL endpoint: enforce POST with a clearer 405 for other methods
	r.Handle("/graphql", methodNotAllowedHint(bodySizeLimit(cfg.MaxBodySize)(graphqlmiddleware.AuthMiddleware(jwtService)(srv))))
	r.Get("/", playground.Handler("GraphQL playground", "/graphql"))
	r.Get("/playground", playground.Handler("GraphQL playground", "/graphql"))

	// HTTP server
	addr := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Show beautiful startup display
	envMode := "development"

	dbLabel := cfg.ConnectionName
	if dbLabel == "" {
		dbLabel = cfg.DBName
	}
	startup.ShowStartup(&startup.StartupDisplay{
		AppName:     "graphql-api",
		Version:     "1.0.0",
		Env:         envMode,
		Config: map[string]string{
			"config": ".env loaded",
		},
		DB:          dbLabel + " connected (" + cfg.DBName + ")",
		DBPoolStats: dbPoolStats,
		Server:      addr,
		Endpoints: []startup.Endpoint{
			{Name: "graphql", URL: "http://localhost:" + cfg.Port + "/graphql"},
			{Name: "playground", URL: "http://localhost:" + cfg.Port + "/playground"},
		},
		StartTime: startTime,
	})

	// Graceful shutdown — wait for SIGINT/SIGTERM
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server start failed: %v\n", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	startup.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server forced to shutdown: %v\n", err)
		os.Exit(1)
	}
}

func methodNotAllowedHint(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, `{"errors":[{"message":"Method Not Allowed. Use the playground at / or /playground for GraphQL queries."}]}`)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// bodySizeLimit rejects requests with bodies larger than maxBytes.
// Uses http.MaxBytesReader which returns a 413 Payload Too Large if the body exceeds the limit.
func bodySizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}