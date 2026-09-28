package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-advanced-admin/admin"
	admingorm "github.com/go-advanced-admin/orm-gorm"
	adminecho "github.com/go-advanced-admin/web-echo"
	"golang.org/x/time/rate"

	"kori/internal/api/middleware"
	"kori/internal/api/validator"
	"kori/internal/config"
	"kori/internal/handlers"
	"kori/internal/models"
	"kori/internal/routes"
	"kori/internal/utils"

	console "kori/internal/utils/logger"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

type Server struct {
	echo   *echo.Echo
	config *config.Config
	db     *gorm.DB
}

var log = console.New("API-Server")

// NewServer @title Kori API
// @version 1.0
// @description This is the API documentation for the Kori project.
// @host localhost:8080
// @BasePath /api/v1
func NewServer(cfg *config.Config, db *gorm.DB) *Server {
	e := echo.New()

	// Create custom validator
	e.Validator = validator.NewValidator()

	// Configure middleware
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORSWithConfig(corsConfig()))
	e.Use(echomiddleware.RequestID())
	e.Use(echomiddleware.Secure())
	e.Use(echomiddleware.TimeoutWithConfig(echomiddleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))
	e.Use(echomiddleware.GzipWithConfig(echomiddleware.GzipConfig{
		Level: 5,
	}))
	e.Use(echomiddleware.BodyLimit("10M"))

	// Custom error handler
	e.HTTPErrorHandler = customHTTPErrorHandler

	// Create server instance
	s := &Server{
		echo:   e,
		config: cfg,
		db:     db,
	}

	// Seed permissions
	if err := models.SeedPermissions(db); err != nil {
		log.Warn("Warning: Failed to seed permissions: %v", err)
	} else {
		log.Success("Successfully seeded permissions")
	}

	if err := models.CreateSuperAdminFromEnv(db, cfg); err != nil {
		log.Warn("Warning: Failed to create super admin: %v", err)
	} else {
		log.Success("Successfully created super admin")
	}

	// Initialize Redis/Valkey client for rate limiting
	redisClient, redisErr := utils.NewRedisClient(cfg)
	if redisErr != nil {
		log.Warn("Warning: Failed to initialize Redis/Valkey client for rate limiting: %v", redisErr)
		// Fall back to basic rate limiting without Redis/Valkey
		e.Use(echomiddleware.RateLimiter(echomiddleware.NewRateLimiterMemoryStore(rate.Limit(20))))
	} else {
		// Configure advanced rate limiting with Redis/Valkey
		rateLimitConfig := middleware.CreateDefaultRateLimitConfig(redisClient.Client)
		e.Use(middleware.RateLimiter(rateLimitConfig))
		log.Success("Successfully configured rate limiting with %s", redisClient.Type)
	}

	// Create a new GORM integrator
	gormIntegrator := admingorm.NewIntegrator(db)
	// Create a new Echo integrator
	echoIntegrator := adminecho.NewIntegrator(e.Group(""))

	// Define your permission checker function
	permissionChecker := func(
		request admin.PermissionRequest, ctx interface{},
	) (bool, error) {
		// Implement your permission logic here
		return false, nil
	}

	// Create a new admin panel
	adminPanel, err := admin.NewPanel(
		gormIntegrator, echoIntegrator, permissionChecker, nil,
	)
	if err != nil {
		err := log.Error("Failed to create admin panel", err)
		if err != nil {
			return nil
		}
	}

	// Register the admin panel
	_, err = adminPanel.RegisterApp(
		"Kori",
		"Kori Admin Panel",
		nil,
	)
	if err != nil {
		err := log.Error("Failed to create admin panel", err)
		if err != nil {
			return nil
		}
	}

	trackingHandler := handlers.NewTrackingHandler(s.db)

	routes.SetupAuthRoutes(s.echo, s.db, s.config)

	// Register routes
	s.registerRoutes()
	routes.SetupImportRoutes(s.echo, s.db, s.config)
	routes.SetupSMTPRoutes(s.echo, s.config, s.db)
	routes.SetupEMAILRoutes(s.echo, s.config, s.db)
	routes.SetupIMAPRoutes(s.echo, s.config, s.db)
	routes.RegisterTrackingRoutes(s.echo, trackingHandler, s.config, s.db)
	return s
}

func (s *Server) Start() error {
	return s.echo.Start(fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port))
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}

func (s *Server) GetEcho() *echo.Echo {
	return s.echo
}

// Health check endpoint
func (s *Server) healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"version": "1.0.0",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// Custom HTTP error handler
func customHTTPErrorHandler(err error, c echo.Context) {
	var (
		code    = http.StatusInternalServerError
		message interface{}
	)

	switch e := err.(type) {
	case *echo.HTTPError:
		code = e.Code
		message = e.Message
	case validator.ValidationErrors:
		code = http.StatusBadRequest
		message = formatValidationErrors(e)
	default:
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			message = he.Message
		} else {
			message = http.StatusText(code)
		}
	}

	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, map[string]interface{}{
				"error": message,
				"code":  code,
				"time":  time.Now().Format(time.RFC3339),
			})
		}
		if err != nil {
			c.Echo().Logger.Error(err)
		}
	}
}

// formatValidationErrors formats validation errors into a map
func formatValidationErrors(errors validator.ValidationErrors) map[string]string {
	errMap := make(map[string]string)
	for _, err := range errors {
		field := err.Field()
		tag := err.Tag()
		param := err.Param()

		switch tag {
		case "required":
			errMap[field] = fmt.Sprintf("%s is required", field)
		case "email":
			errMap[field] = fmt.Sprintf("%s must be a valid email", field)
		case "min":
			errMap[field] = fmt.Sprintf("%s must be at least %s", field, param)
		case "max":
			errMap[field] = fmt.Sprintf("%s must be at most %s", field, param)
		case "url":
			errMap[field] = fmt.Sprintf("%s must be a valid URL", field)
		case "uuid":
			errMap[field] = fmt.Sprintf("%s must be a valid UUID", field)
		case "oneof":
			errMap[field] = fmt.Sprintf("%s must be one of [%s]", field, param)
		case "hostname":
			errMap[field] = fmt.Sprintf("%s must be a valid hostname", field)
		case "fqdn":
			errMap[field] = fmt.Sprintf("%s must be a valid domain name", field)
		case "gt":
			errMap[field] = fmt.Sprintf("%s must be greater than %s", field, param)
		case "required_if":
			errMap[field] = fmt.Sprintf("%s is required when %s", field, param)
		case "json":
			errMap[field] = fmt.Sprintf("%s must be valid JSON", field)
		case "user_role":
			errMap[field] = fmt.Sprintf("%s must be either 'admin' or 'member'", field)
		case "campaign_status":
			errMap[field] = fmt.Sprintf("%s must be one of: DRAFT, SCHEDULED, RUNNING, COMPLETED, FAILED", field)
		default:
			errMap[field] = fmt.Sprintf("%s failed validation: %s", field, tag)
		}
	}
	return errMap
}
