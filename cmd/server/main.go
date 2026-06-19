package main

import (
	"context"
	"os"
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

	userRepo := user.NewRepository(database.Pool)
	vendorRepo := vendor.NewRepository(database.Pool)
	riskAssessmentRepo := risk_assessment.NewRepository(database.Pool)
	compRepo := compliance_record.NewRepository(database.Pool)
	contractRepo := contract.NewRepository(database.Pool)
	auditRepo := audit_trail.NewRepository(database.Pool)
	reportRepo := report.NewRepository(database.Pool)
	categoryRepo := category.NewRepository(database.Pool)

	jwtService := services.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryHours)
	seqService := services.NewSequenceService(database.Pool)

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
	app.Use(middleware.StructuredLogger(logger))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Accept,Authorization,Content-Type",
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiter: 100 requests per minute
	limiter := middleware.NewRateLimiter(100, time.Minute)
	app.Use(limiter.Middleware)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()
		if err := database.Pool.Ping(ctx); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unhealthy",
				"error":  "database unreachable",
			})
		}
		return c.JSON(fiber.Map{"status": "healthy"})
	})

	// Public routes
	app.Post("/api/v1/login", authHandler.Login)

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

	protected.Put("/risk-assessments/:code/approve", riskAssessmentHandler.Approve, middleware.RequirePermission("canApproveRisk"))

	// Compliance
	compGroup := protected.Group("/compliance", middleware.RequirePermission("canReviewCompliance"))
	compGroup.Post("", compHandler.Create)
	compGroup.Get("", compHandler.List)
	compGroup.Get("/:code", compHandler.Get)
	compGroup.Put("/:code", compHandler.Update)
	protected.Get("/compliance/expiring", compHandler.Expiring)

	// Contracts
	contractGroup := protected.Group("/contracts", middleware.RequirePermission("canEditVendor"))
	contractGroup.Post("", contractHandler.Create)
	contractGroup.Get("", contractHandler.List)
	contractGroup.Get("/:code", contractHandler.Get)
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

	logger.Info().Msg("VRMP server starting on http://localhost:8080")
	if err := app.Listen(":8080"); err != nil {
		logger.Fatal().Err(err).Msg("server start failed")
	}
}