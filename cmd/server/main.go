package main

import (
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"

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
	"github.com/uswuth/vytora-backend/internal/services"
)

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	cfg := config.Load()

	if err := database.Connect(cfg.DatabaseURL); err != nil {
		logger.Fatal().Err(err).Msg("Database connection failed")
	}
	defer database.Close()

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

	// HTTP router
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
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

	// GraphQL endpoint & playground with body size limit
	r.Handle("/graphql", bodySizeLimit(cfg.MaxBodySize)(graphqlmiddleware.AuthMiddleware(jwtService)(srv)))
	r.Get("/", playground.Handler("GraphQL playground", "/graphql"))

	logger.Info().Msg("GraphQL server starting on 0.0.0.0:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", r); err != nil {
		logger.Fatal().Err(err).Msg("server start failed")
	}
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