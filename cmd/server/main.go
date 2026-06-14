package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chim "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/uswuth/vytora-backend/internal/config"
	"github.com/uswuth/vytora-backend/internal/database"
	"github.com/uswuth/vytora-backend/internal/handlers"
	"github.com/uswuth/vytora-backend/internal/middleware" // auth + rbac
	"github.com/uswuth/vytora-backend/internal/repository"
	"github.com/uswuth/vytora-backend/internal/services"
)

func main() {
	cfg := config.Load()

	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer database.Close()

	// Repositories
	userRepo := repository.NewUserRepository(database.Pool)
	vendorRepo := repository.NewVendorRepository(database.Pool)
	riskAssessmentRepo := repository.NewRiskAssessmentRepository(database.Pool)
	compRepo := repository.NewComplianceRecordRepository(database.Pool)

	// Services
	jwtService := services.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryHours)
	seqService := services.NewSequenceService(database.Pool)

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo, jwtService)
	vendorHandler := handlers.NewVendorHandler(vendorRepo, seqService)
	riskAssessmentHandler := handlers.NewRiskAssessmentHandler(riskAssessmentRepo, vendorRepo, seqService)
	compHandler := handlers.NewComplianceRecordHandler(compRepo, vendorRepo, seqService)

	r := chi.NewRouter()

	// Global middleware
	r.Use(chim.Logger)
	r.Use(chim.Recoverer)
	r.Use(chim.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes
	r.Post("/api/v1/login", authHandler.Login)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Protected routes (JWT required)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtService))

		r.Get("/api/v1/me", func(w http.ResponseWriter, r *http.Request) {
			claims := r.Context().Value(middleware.UserContextKey).(*services.Claims)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"user_id":"%s","code":"%s","email":"%s","role":"%s"}`,
				claims.UserID, claims.Code, claims.Email, claims.Role)
		})

		// Vendor routes – with RBAC
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canCreateVendor"))
			r.Post("/api/v1/vendors", vendorHandler.Create)
		})

		// These can have separate RBAC checks
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canEditVendor")) // covers view + edit for now
			r.Get("/api/v1/vendors", vendorHandler.List)
			r.Get("/api/v1/vendors/{code}", vendorHandler.Get)
			r.Put("/api/v1/vendors/{code}", vendorHandler.Update)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canDeleteVendor"))
			r.Delete("/api/v1/vendors/{code}", vendorHandler.Delete)
		})

		// Risk Assessment routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canCreateRiskAssessment"))
			r.Post("/api/v1/risk-assessments", riskAssessmentHandler.Create)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canReviewRisk")) // also used for view list/detail
			r.Get("/api/v1/risk-assessments", riskAssessmentHandler.List)
			r.Get("/api/v1/risk-assessments/{code}", riskAssessmentHandler.Get)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canApproveRisk"))
			r.Put("/api/v1/risk-assessments/{code}/approve", riskAssessmentHandler.Approve)
		})

		// Compliance routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canReviewCompliance"))
			r.Post("/api/v1/compliance", compHandler.Create)
			r.Get("/api/v1/compliance", compHandler.List)
			r.Get("/api/v1/compliance/{code}", compHandler.Get)
			r.Put("/api/v1/compliance/{code}", compHandler.Update)
		})
		r.Get("/api/v1/compliance/expiring", compHandler.Expiring) // read access for anyone authenticated
	})

	fmt.Println("VRMP server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
