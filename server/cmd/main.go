package main

import (
	"bytes"
	"context"
	"gopkg.in/gomail.v2"
	"kori/docs/swagger"
	"kori/internal/handlers"
	"kori/internal/keys"
	"kori/internal/models/seeder/airley"
	"kori/internal/sending"
	"kori/internal/utils"
	"kori/internal/utils/crypto"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kori/internal/ai"
	"kori/internal/api"
	"kori/internal/automation"
	"kori/internal/automation/processors"
	"kori/internal/config"
	"kori/internal/db"
	"kori/internal/models"
	"kori/internal/routes"
	"kori/internal/services"
	"kori/internal/tasks"
	"kori/internal/utils/logger"

	"github.com/joho/godotenv"
)

// 🚀 Main function
// @Summary Main function
// @Description Main function
// @title Xem API
// @version 1.0
// @description API documentation for Xem API
// @host api.xem.email
// @BasePath /
// @schemes https

// @securityDefinitions.basic BasicAuth
// @in header
// @name Authorization

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY
func main() {

	logger := logger.New("kori")

	// check if .env file exists
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		logger.Info("No .env file found, skipping environment variable loading")
	} else {
		logger.Info("Loading environment variables from .env file")
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Failed to load environment variables: %v", err)
		}
	}

	_, err := keys.NewInfisicalSecrets(os.Getenv("INFISICAL_CLIENT_SECRET") != "")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	services.Initialize(cfg)

	// Initialize keys
	if err := crypto.InitializeKeys(
		cfg.Crypto.PrivateKey); err != nil {
		log.Fatalf("Failed to initialize keys: %v", err)
	}

	// Connect to database
	if err := db.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
	}()

	// Set up Discord webhook for monitoring
	if cfg.Monitor.DiscordWebhookURL != "" {
		db.SetDiscordWebhook(cfg.Monitor.DiscordWebhookURL)
		logger.Info("Discord webhook monitoring enabled")
	}

	// Start monitoring database connection pool
	db.MonitorConnectionPool(10 * time.Hour)

	db_instance := db.GetDB()

	managedConfig, err := sending.LoadConfig()
	if err != nil {
		log.Fatalf("Invalid managed sending configuration: %v", err)
	}
	if err := sending.Migrate(db_instance); err != nil {
		log.Fatalf("Managed sending migration failed: %v", err)
	}
	var provider sending.Provider
	if managedConfig.Enabled {
		provider, err = sending.NewSES(context.Background(), managedConfig)
		if err != nil {
			log.Fatalf("Managed sending initialization failed: %v", err)
		}
	}
	managed := sending.New(db_instance, managedConfig, provider)
	managedCtx, stopManaged := context.WithCancel(context.Background())
	defer stopManaged()
	managedDone := make(chan struct{})
	closeManagedSMTP := func() {}
	utils.RecipientPolicy = func(team string, recipients []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return managed.RecipientPolicy(ctx, team, recipients)
	}
	if managedConfig.Enabled {
		utils.ManagedDelivery = func(email *models.Email, message *gomail.Message) error {
			var raw bytes.Buffer
			if _, err := message.WriteTo(&raw); err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			return managed.SubmitEmail(ctx, email, raw.Bytes())
		}
		go func() { defer close(managedDone); managed.Run(managedCtx) }()
		if managedConfig.SMTPEnabled {
			smtpServer, err := sending.StartSMTP(managed)
			if err != nil {
				log.Fatalf("Managed SMTP startup failed: %v", err)
			}
			closeManagedSMTP = func() { _ = smtpServer.Close() }
			defer closeManagedSMTP()
		}
	}
	// Initialize task handlers
	taskHandler := tasks.NewTaskHandler(db_instance, cfg)

	// Initialize task client for automation engine
	taskClient := tasks.NewTaskClient(cfg.Redis.Addr, cfg.Redis.Username, cfg.Redis.Password, cfg.Redis.DB, cfg.Redis.UseTLS)

	// Initialize AI client if enabled
	var aiClient ai.AIClient
	if cfg.AI.Enabled {
		if cfg.AI.Provider == "anthropic" {
			aiClient = ai.NewAnthropicClient(cfg.AI.AnthropicAPIKey, cfg.AI.Model)
			logger.Success("Initialized Anthropic AI client with model: %s", cfg.AI.Model)
		} else {
			logger.Warn("Unknown AI provider: %s, AI features will be disabled", cfg.AI.Provider)
		}
	}

	// Initialize automation engine
	automationEngine := automation.NewEngine(db_instance, taskClient)

	// Register node processors
	automationEngine.RegisterProcessor(processors.NewStartProcessor(db_instance))
	automationEngine.RegisterProcessor(processors.NewEmailProcessor(db_instance, taskClient, cfg))
	automationEngine.RegisterProcessor(processors.NewExitProcessor(db_instance))
	automationEngine.RegisterProcessor(processors.NewWaitProcessor(db_instance))
	automationEngine.RegisterProcessor(processors.NewConditionProcessor(db_instance))
	automationEngine.RegisterProcessor(processors.NewAddToListProcessor(db_instance))
	automationEngine.RegisterProcessor(processors.NewTagProcessor(db_instance))
	automationEngine.RegisterProcessor(processors.NewPercentageSplitProcessor(db_instance))
	automationEngine.RegisterProcessor(processors.NewSetVariableProcessor(db_instance))
	automationEngine.RegisterProcessor(processors.NewWebhookProcessor(db_instance))
	automationEngine.RegisterProcessor(processors.NewUpdateSubscriberProcessor(db_instance))

	// Register segment processors
	automationEngine.RegisterProcessor(processors.NewSegmentFilterProcessor(db_instance))
	automationEngine.RegisterProcessor(processors.NewAddToSegmentProcessor(db_instance))
	automationEngine.RegisterProcessor(processors.NewRemoveFromSegmentProcessor(db_instance))
	logger.Success("Registered segment processors")

	// Register lead scoring processors
	automationEngine.RegisterProcessor(processors.NewScoreChangeProcessor(db_instance))
	automationEngine.RegisterProcessor(processors.NewScoreThresholdProcessor(db_instance))
	logger.Success("Registered lead scoring processors")

	// Register A/B testing processors
	automationEngine.RegisterProcessor(processors.NewABSplitProcessor(db_instance))
	logger.Success("Registered A/B testing processors")

	// Register goal tracking processors
	automationEngine.RegisterProcessor(processors.NewGoalTrackProcessor(db_instance))
	logger.Success("Registered goal tracking processors")

	// Register AI decision processor if AI is enabled
	if cfg.AI.Enabled && aiClient != nil {
		automationEngine.RegisterProcessor(processors.NewAIDecisionProcessor(db_instance, aiClient, cfg))
		logger.Success("Registered AI decision processor")
	}

	// Connect engine to task handler
	taskHandler.SetAutomationEngine(automationEngine)

	// Initialize trigger manager for event-based automations
	triggerManager := automation.NewTriggerManager(db_instance, taskClient)
	if err := triggerManager.Initialize(context.Background()); err != nil {
		log.Fatalf("Failed to initialize trigger manager: %v", err)
	}

	// Initialize task server
	taskServer := tasks.NewServer(
		cfg.Redis.Addr,
		cfg.Redis.Username,
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Redis.UseTLS,
		taskHandler,
		logger,
	)

	// Create a context for task server
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	// Start task server
	go func() {
		if err := taskServer.Start(serverCtx); err != nil {
			logger.Error("Task server error", err)
		}
	}()

	// Initialize task scheduler
	taskScheduler := tasks.NewScheduler(
		cfg.Redis.Addr,
		cfg.Redis.Username,
		cfg.Redis.Password,
		cfg.Redis.DB,
		logger,
		cfg.Redis.UseTLS,
	)

	// Start task scheduler
	go func() {
		if err := taskScheduler.Start(); err != nil {
			logger.Error("Task scheduler error", err)
		}
	}()

	// Initialize API server
	apiServer := api.NewServer(cfg, db_instance)
	managed.Register(apiServer.GetEcho(), cfg.JWT.Secret)

	routes.SetupMarketingRoutes(apiServer.GetEcho(), db_instance, cfg)

	// Register automation routes
	routes.SetupAutomationRoutes(apiServer.GetEcho(), db_instance, cfg, taskClient)

	// Register segment routes
	routes.SetupSegmentRoutes(apiServer.GetEcho(), db_instance, cfg.JWT.Secret)
	logger.Success("Registered segment routes")

	// Register lead scoring routes
	routes.SetupLeadScoringRoutes(apiServer.GetEcho(), db_instance, cfg.JWT.Secret)
	logger.Success("Registered lead scoring routes")

	// Register A/B testing routes
	routes.SetupABTestingRoutes(apiServer.GetEcho(), db_instance, cfg.JWT.Secret)
	logger.Success("Registered A/B testing routes")

	// Register goal tracking routes
	routes.SetupGoalTrackingRoutes(apiServer.GetEcho(), db_instance, cfg.JWT.Secret)
	logger.Success("Registered goal tracking routes")

	// Register AI routes if enabled
	if cfg.AI.Enabled && aiClient != nil {
		routes.SetupAIRoutes(apiServer.GetEcho(), db_instance, cfg, aiClient)
		logger.Success("Registered AI routes")
	}

	go func() {

		// Initialize S3 service
		s3Service, err := services.NewS3Service(
			cfg.Storage.S3.BucketName,
			cfg.Storage.S3.Endpoint,
			cfg.Storage.S3.Region,
			cfg.Storage.S3.AccessKey,
			cfg.Storage.S3.SecretKey,
		)

		if err != nil {
			log.Fatalf("Failed to initialize S3 service: %v", err)
		}

		// Register the URL generator
		models.RegisterFileURLGenerator(s3Service)
		handlers.RegisterStorageHandler(s3Service)

		if cfg.Airley.Enabled {
			// Seed Airley templates
			if err := airley.LoadAirleyTemplates(db_instance); err != nil {
				logger.Error("Warning: Failed to seed Airley templates: %v", err)
			} else {
				logger.Success("Successfully seeded Airley templates")
			}
		}
		logger.Success("API server started")

		// Swagger documentation
		swagger.SwaggerInfo.Title = "Posthoot API Documentation"
		swagger.SwaggerInfo.Description = "API documentation for Posthoot application"
		swagger.SwaggerInfo.Version = "1.0"
		swagger.SwaggerInfo.Host = "backyard.posthoot.com"
		swagger.SwaggerInfo.Schemes = []string{"https"}

		if err := apiServer.Start(); err != nil {
			logger.Error("API server error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop task scheduler
	taskScheduler.Stop()

	// Stop task server
	serverCancel()

	// Shutdown API server
	if err := apiServer.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown API server", err)
	}

	closeManagedSMTP()
	stopManaged()
	if managedConfig.Enabled {
		select {
		case <-managedDone:
		case <-time.After(15 * time.Second):
			logger.Info("Managed worker shutdown timed out; stale claims will be quarantined")
		}
	}
	logger.Info("Servers shutdown gracefully")
}
