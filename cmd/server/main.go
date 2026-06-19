package main

import (
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	cfg := config.Load()

	if err := database.Connect(cfg.DatabaseURL); err != nil {
		logger.Fatal().Err(err).Msg("Database connection failed")
	}
	defer database.Close()

	// repo
	userRepo := user.NewRepository(database.Pool)
	vendorRepo := vendor.NewRepository(database.Pool)
	riskAssessmentRepo := risk_assessment.NewRepository(database.Pool)
	compRepo := compliance_record.NewRepository(database.Pool)
	contractRepo := contract.NewRepository(database.Pool)
	auditRepo := audit_trail.NewRepository(database.Pool)
	reportRepo := report.NewRepository(database.Pool)
	categoryRepo := category.NewRepository(database.Pool)

	jwtService := services.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryHours)
	secretPrefix := cfg.JWTSecret
	if len(secretPrefix) > 8 {
		secretPrefix = secretPrefix[:8]
	}
	logger.Info().Str("jwt_secret_prefix", secretPrefix).Msg("JWT secret loaded")
	seqService := services.NewSequenceService(database.Pool)

	// handler
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

	app := fiber.New()

	// Global middleware
	app.Use(middleware.RequestIDMiddleware)
	app.Use(middleware.StructuredLogger(logger))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(cfg.AllowedOrigins, ","),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Accept,Authorization,Content-Type",
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Rate limiter: 100 requests per minute
	limiter := middleware.NewRateLimiter(100, time.Minute)
	app.Use(limiter.Middleware)

	// Prometheus metrics middleware
	app.Use(middleware.MetricsMiddleware)

	// Health checks
	healthChecker := middleware.NewHealthChecker(logger, cfg.HealthAllowedIPs)
	healthChecker.RegisterRoutes(app)

	// Prometheus metrics endpoint
	app.Get("/metrics", middleware.MetricsHandler())

	// Public routes
	app.Post("/api/v1/login", authHandler.Login)

	// Protected auth routes
	authGroup := app.Group("/api/v1/auth")
	authGroup.Use(middleware.AuthMiddleware(jwtService))
	authGroup.Post("/extend", authHandler.ExtendSession)

	// User management (admin only)
	userGroup := app.Group("/api/v1/users")
	userGroup.Use(middleware.AuthMiddleware(jwtService))
	userGroup.Use(middleware.RequirePermission("canManageUsers"))
	userGroup.Post("", userManagementHandler.Create)
	userGroup.Get("", userManagementHandler.List)
	userGroup.Get("/:id", userManagementHandler.Get)
	userGroup.Put("/:id/role", userManagementHandler.UpdateRole)
	userGroup.Put("/:id/deactivate", userManagementHandler.Deactivate)
	userGroup.Put("/:id/activate", userManagementHandler.Activate)

	// Protected routes
	protected := app.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(jwtService))

	protected.Get("/me", func(c *fiber.Ctx) error {
		claims := c.Locals(middleware.UserContextKey).(*services.Claims)
		return c.JSON(fiber.Map{
			"user_id": claims.UserID,
			"code":    claims.Code,
			"email":   claims.Email,
			"role":    claims.Role,
		})
	})

	// Vendor routes
	protected.Post("/vendors", vendorHandler.Create, middleware.RequirePermission("canCreateVendor"))

	vendorEditGroup := protected.Group("/vendors", middleware.RequirePermission("canEditVendor"))
	vendorEditGroup.Get("", vendorHandler.List)
	vendorEditGroup.Get("/:code", vendorHandler.Get)
	vendorEditGroup.Put("/:code", vendorHandler.Update)

	vendorDeleteGroup := protected.Group("/vendors", middleware.RequirePermission("canDeleteVendor"))
	vendorDeleteGroup.Delete("/:code", vendorHandler.Delete)

	// Workflow
	protected.Put("/vendors/:code/submit", workflowHandler.Submit, middleware.RequirePermission("canSubmitVendorRequest"))
	protected.Put("/vendors/:code/review-risk", workflowHandler.ReviewRisk, middleware.RequirePermission("canReviewRisk"))
	protected.Put("/vendors/:code/review-compliance", workflowHandler.ReviewCompliance, middleware.RequirePermission("canReviewCompliance"))
	protected.Put("/vendors/:code/approve", workflowHandler.Approve, middleware.RequirePermission("canEditVendor"))
	protected.Put("/vendors/:code/reject", workflowHandler.Reject, middleware.RequirePermission("canEditVendor"))

	// Risk assessments
	protected.Post("/risk-assessments", riskAssessmentHandler.Create, middleware.RequirePermission("canCreateRiskAssessment"))

	riskReviewGroup := protected.Group("/risk-assessments", middleware.RequirePermission("canReviewRisk"))
	riskReviewGroup.Get("", riskAssessmentHandler.List)
	riskReviewGroup.Get("/:code", riskAssessmentHandler.Get)

	riskManageGroup := protected.Group("/risk-assessments", middleware.RequirePermission("canReviewRisk"))
	riskManageGroup.Put("/:code", riskAssessmentHandler.Update)
	riskManageGroup.Delete("/:code", riskAssessmentHandler.Delete)

	protected.Put("/risk-assessments/:code/approve", riskAssessmentHandler.Approve, middleware.RequirePermission("canApproveRisk"))

	// Compliance
	compGroup := protected.Group("/compliance", middleware.RequirePermission("canReviewCompliance"))
	compGroup.Post("", compHandler.Create)
	compGroup.Get("", compHandler.List)
	compGroup.Get("/:code", compHandler.Get)
	compGroup.Put("/:code", compHandler.Update)
	compGroup.Delete("/:code", compHandler.Delete)
	protected.Get("/compliance/expiring", compHandler.Expiring)

	// Contracts
	contractGroup := protected.Group("/contracts", middleware.RequirePermission("canEditVendor"))
	contractGroup.Post("", contractHandler.Create)
	contractGroup.Get("", contractHandler.List)
	contractGroup.Get("/:code", contractHandler.Get)
	contractGroup.Put("/:code", contractHandler.Update)
	contractGroup.Delete("/:code", contractHandler.Delete)
	protected.Get("/contracts/expiring", contractHandler.Expiring)

	// Audit
	auditGroup := protected.Group("/audit", middleware.RequirePermission("canViewAuditHistory"))
	auditGroup.Get("", auditHandler.List)

	// Reports
	protected.Get("/reports/summary", reportHandler.Summary, middleware.RequirePermission("canAccessAllReports"))
	protected.Get("/reports/monthly-onboarding", reportHandler.MonthlyOnboarding, middleware.RequirePermission("canAccessAllReports"))
	protected.Get("/reports/summary-2", reportHandler.Summary, middleware.RequirePermission("canViewAssignedVendors"))
	protected.Get("/reports/monthly-onboarding-2", reportHandler.MonthlyOnboarding, middleware.RequirePermission("canViewAssignedVendors"))

	// Category routes
	catManageGroup := protected.Group("/categories", middleware.RequirePermission("canManageCategories"))
	catManageGroup.Post("", categoryHandler.Create)
	catManageGroup.Put("/:code", categoryHandler.Update)
	catManageGroup.Delete("/:code", categoryHandler.Delete)

	catViewGroup := protected.Group("/categories", middleware.RequirePermission("canViewCategories"))
	catViewGroup.Get("", categoryHandler.List)
	catViewGroup.Get("/:code", categoryHandler.Get)

	logger.Info().Msg("VRMP server starting on 0.0.0.0:8080")
	if err := app.Listen("0.0.0.0:8080"); err != nil {
		logger.Fatal().Err(err).Msg("server start failed")
	}
}
