package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chim "github.com/go-chi/chi/v5/middleware"
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
	"github.com/uswuth/vytora-backend/internal/handlers"
	"github.com/uswuth/vytora-backend/internal/middleware"
	"github.com/uswuth/vytora-backend/internal/services"
)

func main() {
	// Setup logger
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

	// Services
	jwtService := services.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryHours)
	seqService := services.NewSequenceService(database.Pool)

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo, jwtService)
	userManagementHandler := user.NewHandler(userRepo, seqService.NextCode)
	vendorHandler := vendor.NewHandler(vendorRepo, categoryRepo, seqService.NextCode)
	riskAssessmentHandler := risk_assessment.NewHandler(riskAssessmentRepo, vendorRepo, seqService.NextCode)
	compHandler := compliance_record.NewHandler(compRepo, vendorRepo, seqService.NextCode)
	contractHandler := contract.NewHandler(contractRepo, vendorRepo, seqService.NextCode)
	workflowHandler := vendor.NewWorkflowHandler(vendorRepo, auditRepo, seqService.NextCode)
	auditHandler := audit_trail.NewHandler(auditRepo)
	reportHandler := report.NewHandler(reportRepo)
	categoryHandler := category.NewHandler(categoryRepo, seqService.NextCode)

	r := chi.NewRouter()

	// Global middleware
	r.Use(chim.RequestID)
	r.Use(middleware.StructuredLogger(logger))
	r.Use(chim.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiter: 100 requests per minute
	limiter := middleware.NewRateLimiter(100, time.Minute)
	r.Use(limiter.Middleware)

	// Health check (also pings DB)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := database.Pool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","error":"database unreachable"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Public routes
	r.Post("/api/v1/login", authHandler.Login)

	// User management (admin only)
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequirePermission("canManageUsers"))
		r.Post("/api/v1/users", userManagementHandler.Create)
		r.Get("/api/v1/users", userManagementHandler.List)
		r.Get("/api/v1/users/{id}", userManagementHandler.Get)
		r.Put("/api/v1/users/{id}/role", userManagementHandler.UpdateRole)
		r.Put("/api/v1/users/{id}/deactivate", userManagementHandler.Deactivate)
		r.Put("/api/v1/users/{id}/activate", userManagementHandler.Activate)
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

		// Vendor routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canCreateVendor"))
			r.Post("/api/v1/vendors", vendorHandler.Create)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canEditVendor"))
			r.Get("/api/v1/vendors", vendorHandler.List)
			r.Get("/api/v1/vendors/{code}", vendorHandler.Get)
			r.Put("/api/v1/vendors/{code}", vendorHandler.Update)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canDeleteVendor"))
			r.Delete("/api/v1/vendors/{code}", vendorHandler.Delete)
		})

		// Workflow
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canSubmitVendorRequest"))
			r.Put("/api/v1/vendors/{code}/submit", workflowHandler.Submit)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canReviewRisk"))
			r.Put("/api/v1/vendors/{code}/review-risk", workflowHandler.ReviewRisk)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canReviewCompliance"))
			r.Put("/api/v1/vendors/{code}/review-compliance", workflowHandler.ReviewCompliance)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canEditVendor"))
			r.Put("/api/v1/vendors/{code}/approve", workflowHandler.Approve)
			r.Put("/api/v1/vendors/{code}/reject", workflowHandler.Reject)
		})

		// Risk assessments
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canCreateRiskAssessment"))
			r.Post("/api/v1/risk-assessments", riskAssessmentHandler.Create)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canReviewRisk"))
			r.Get("/api/v1/risk-assessments", riskAssessmentHandler.List)
			r.Get("/api/v1/risk-assessments/{code}", riskAssessmentHandler.Get)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canApproveRisk"))
			r.Put("/api/v1/risk-assessments/{code}/approve", riskAssessmentHandler.Approve)
		})

		// Compliance
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canReviewCompliance"))
			r.Post("/api/v1/compliance", compHandler.Create)
			r.Get("/api/v1/compliance", compHandler.List)
			r.Get("/api/v1/compliance/{code}", compHandler.Get)
			r.Put("/api/v1/compliance/{code}", compHandler.Update)
		})
		r.Get("/api/v1/compliance/expiring", compHandler.Expiring)

		// Contracts
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canEditVendor"))
			r.Post("/api/v1/contracts", contractHandler.Create)
			r.Get("/api/v1/contracts", contractHandler.List)
			r.Get("/api/v1/contracts/{code}", contractHandler.Get)
		})
		r.Get("/api/v1/contracts/expiring", contractHandler.Expiring)

		// Audit
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canViewAuditHistory"))
			r.Get("/api/v1/audit", auditHandler.List)
		})

		// Reports
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canAccessAllReports"))
			r.Get("/api/v1/reports/summary", reportHandler.Summary)
			r.Get("/api/v1/reports/monthly-onboarding", reportHandler.MonthlyOnboarding)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canViewAssignedVendors"))
			r.Get("/api/v1/reports/summary", reportHandler.Summary)
			r.Get("/api/v1/reports/monthly-onboarding", reportHandler.MonthlyOnboarding)
		})
		// Category routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canManageCategories"))
			r.Post("/api/v1/categories", categoryHandler.Create)
			r.Put("/api/v1/categories/{code}", categoryHandler.Update)
			r.Delete("/api/v1/categories/{code}", categoryHandler.Delete)
		})
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequirePermission("canViewCategories"))
			r.Get("/api/v1/categories", categoryHandler.List)
			r.Get("/api/v1/categories/{code}", categoryHandler.Get)
		})
	})

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info().Msg("VRMP server starting on http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("server start failed")
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("forced shutdown")
	}
	logger.Info().Msg("Server stopped")
}
