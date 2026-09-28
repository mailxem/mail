package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mailxem/payments.go/docs"
	"github.com/mailxem/payments.go/internal/config"
	"github.com/mailxem/payments.go/internal/handlers"
	"github.com/mailxem/payments.go/internal/keys"
	custommiddleware "github.com/mailxem/payments.go/internal/middleware"
	"github.com/mailxem/payments.go/internal/models"
	"github.com/mailxem/payments.go/internal/repositories"
	"github.com/mailxem/payments.go/internal/services"
	"github.com/mailxem/payments.go/pkg/dodo"
	echoSwagger "github.com/swaggo/echo-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Diagonal Payments API
// @version 1.0
// @description API for managing subscription plans and payments with Dodo Payments integration
// @termsOfService https://diagonal.com/terms

// @contact.name Diagonal Support
// @contact.url https://diagonal.com/support
// @contact.email support@diagonal.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey AdminAuth
// @in header
// @name X-Admin-Secret
// @description Admin secret for accessing admin-only endpoints

func main() {

	// try to load .env file
	//
	// let's load the config from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	_, _ = keys.NewInfisicalSecrets(
		os.Getenv("ENABLE_INFISICAL") == "true",
	)

	// Load configuration
	cfg, err := config.Load()

	fmt.Println(
		"Configuration loaded:",
		"Port:", cfg.Port,
		"Environment:", cfg.Environment,
	)

	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Configure Swagger info based on environment
	docs.SwaggerInfo.Host = cfg.Port
	if cfg.Environment == "production" {
		docs.SwaggerInfo.Host = "paywall.synehq.com"
		docs.SwaggerInfo.Schemes = []string{"https"}
	} else {
		docs.SwaggerInfo.Host = "localhost:" + cfg.Port
		docs.SwaggerInfo.Schemes = []string{"http"}
	}

	// Connect to database
	db, err := setupDatabase(cfg.DatabaseURL, cfg.POSTGRES_MAX_OPEN_CONNS)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := runMigrations(db, cfg); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	} else {
		log.Println("Migrations completed successfully")
	}

	// Initialize dependencies
	dodoClient := dodo.NewClient(cfg)

	subscriptionRepo := repositories.NewSubscriptionRepository(db)
	paymentRepo := repositories.NewPaymentRepository(db)
	usageRepo := repositories.NewUsageRepository(db)
	webhookRepo := repositories.NewWebhookRepository(db)
	webhookLogRepo := repositories.NewWebhookLogRepository(db)
	planRepo := repositories.NewPlanRepository(db)
	planFeaturesRepo := repositories.NewPlanFeaturesRepository(db)
	teamRepo := repositories.NewTeamRepository(db)
	teamService := services.NewTeamService(teamRepo)
	planService := services.NewPlanService(planRepo, planFeaturesRepo)
	emailService := services.NewEmailService(cfg, teamService, planService)

	webhookService := services.NewWebhookService(webhookRepo, webhookLogRepo, subscriptionRepo, paymentRepo, emailService, cfg.DodoWebhookSecret)
	subscriptionService := services.NewSubscriptionService(cfg, subscriptionRepo, paymentRepo, usageRepo, dodoClient, planService, emailService, teamService)
	paymentService := services.NewPaymentService(cfg, subscriptionRepo, paymentRepo, emailService, dodoClient, webhookService, planService)

	subscriptionHandlers := handlers.NewSubscriptionHandlers(subscriptionService, paymentService, webhookService, planService, cfg)
	planHandler := handlers.NewPlanHandler(planService, cfg.AdminSecret)
	webhookHandler := handlers.NewWebhookHandler(webhookService)

	// Setup Echo server
	e := echo.New()

	// Register validator
	e.Validator = &CustomValidator{validator: validator.New()}

	// Set custom error handler
	e.HTTPErrorHandler = custommiddleware.CustomErrorHandler(cfg.Environment == "production")

	setupMiddleware(e, cfg)
	setupRoutes(e, subscriptionHandlers, planHandler, webhookHandler, cfg)

	// Start daily maintenance ticker
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := subscriptionService.RunDailyMaintenance(context.Background()); err != nil {
				log.Printf("daily maintenance error: %v", err)
			} else {
				log.Printf("daily maintenance completed")
			}
		}
	}()

	// Start server
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Payment service started on port %s", cfg.Port)
	log.Printf("Swagger documentation available at http://localhost:%s/swagger/index.html", cfg.Port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// CustomValidator wraps the go-playground validator
type CustomValidator struct {
	validator *validator.Validate
}

// Validate validates a struct
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func setupDatabase(databaseURL string, maxConn int) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(maxConn)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func runMigrations(db *gorm.DB, cfg *config.Config) error {
	err := db.AutoMigrate(
		&models.Subscription{},
		&models.Payment{},
		&models.UsageRecord{},
		&models.WebhookEvent{},
		&models.WebhookLog{},
		&models.Plan{},
		&models.PlanFeatures{},
		&models.BillingDetails{},
	)

	if err != nil {
		return err
	}

	// we need to execute the seed script
	seedScript := filepath.Join("scripts", "seed", fmt.Sprintf("%s.sql", cfg.Environment))
	sqlContent, err := os.ReadFile(seedScript)
	if err != nil {
		return err
	}

	_ = db.Exec(string(sqlContent))

	return nil
}

func setupMiddleware(e *echo.Echo, cfg *config.Config) {
	// Request ID (for tracing)
	e.Use(middleware.RequestID())

	// Logging
	e.Use(middleware.Logger())

	// Panic recovery
	e.Use(middleware.Recover())

	// Security headers
	e.Use(custommiddleware.SecurityHeaders())

	// HSTS (production only)
	e.Use(custommiddleware.HSTS(cfg.Environment == "production"))

	// Request body size limit (10MB max)
	e.Use(custommiddleware.RequestSizeLimit(10 * 1024 * 1024))

	// CORS - Restrictive configuration
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.AllowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-Team-ID",
			"X-API-Key",
			"X-Admin-Secret",
		},
		AllowCredentials: true,
		MaxAge:           3600, // 1 hour
	}))

	// Global rate limiting (generous for normal traffic)
	e.Use(custommiddleware.GenerousRateLimiter())

	// Timeout middleware
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))
}

func setupRoutes(e *echo.Echo, h *handlers.SubscriptionHandlers, planHandler *handlers.PlanHandler, webhookHandler *handlers.WebhookHandler, cfg *config.Config) {
	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// API routes
	api := e.Group("/api/v1")

	// Health check
	api.GET("/health", h.Health)

	// Admin routes - Strict rate limiting + authentication + audit logging
	admin := api.Group("/admin")
	admin.Use(custommiddleware.StrictRateLimiter())
	admin.Use(custommiddleware.AdminAuth(cfg.AdminSecret))
	admin.Use(custommiddleware.AuditMiddleware())
	admin.POST("/maintenance/run", h.RunMaintenance)

	// Plans
	api.GET("/plans", h.GetPlans)

	// Subscriptions
	subscriptions := api.Group("/subscriptions")
	subscriptions.GET("/team/:teamId", h.GetSubscription)
	subscriptions.PUT("/:id", h.UpdateSubscription)
	subscriptions.DELETE("/:id", h.CancelSubscription)
	subscriptions.GET("/:id/usage", h.GetUsageStats)
	subscriptions.GET("/:id/payments", h.GetPaymentHistory)
	subscriptions.GET("/:id/payments/:payment_id/invoice", h.GetInvoicePaymentPDF)
	subscriptions.GET("/:id/session", h.GetCustomerPortalSession)

	// Billing details
	billingDetails := api.Group("/billing")
	billingDetails.POST("/details", h.CreateBillingDetails)
	billingDetails.GET("/details/:teamId", h.GetBillingDetails)
	billingDetails.PUT("/details/:id", h.UpdateBillingDetails)

	// Usage tracking
	usage := api.Group("/usage")
	usage.GET("/check/:teamId/:feature", h.CheckUsageLimit)
	usage.POST("/record/:teamId", h.RecordUsage)

	// Payments - Moderate rate limiting
	payments := api.Group("/payments")
	payments.Use(custommiddleware.ModerateRateLimiter())
	payments.POST("/checkout", h.CreateCheckoutSession)

	// Webhooks - No rate limiting (external service), but has signature verification
	webhooks := api.Group("/webhooks")
	webhooks.POST("/dodo", webhookHandler.HandleDodoWebhook)

	// Register plan routes using the handler's RegisterRoutes method
	planHandler.RegisterRoutes(api)

	// Webhook management routes
	webhookGroup := api.Group("/webhooks")
	webhookGroup.GET("/events", webhookHandler.GetWebhookEvents)
	webhookGroup.GET("/events/:eventId/logs", webhookHandler.GetWebhookLogs)
	webhookGroup.POST("/retry-failed", webhookHandler.RetryFailedWebhooks)
	webhookGroup.POST("/cleanup", webhookHandler.CleanupOldWebhooks)
}
