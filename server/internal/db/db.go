package db

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"kori/internal/analytics"
	"kori/internal/config"
	"kori/internal/models"
	console "kori/internal/utils/logger"
)

var DB *gorm.DB
var log = console.New("DB")
var discordWebhookURL string
var cloudSQLDialer *cloudsqlconn.Dialer

// SetDiscordWebhook sets the Discord webhook URL for monitoring
func SetDiscordWebhook(webhookURL string) {
	discordWebhookURL = webhookURL
}

// sendToDiscord sends connection pool stats to Discord
func sendToDiscord(stats sql.DBStats) error {
	if discordWebhookURL == "" {
		return nil
	}

	message := fmt.Sprintf("```\n🗄️ DB Connection Pool Stats:\n"+
		"🔝 Max Open Connections: %d\n"+
		"📊 Open Connections: %d\n"+
		"🔄 In Use Connections: %d\n"+
		"💤 Idle Connections: %d\n"+
		"⏳ Waited for Connection: %d\n"+
		"⌛ Wait Duration: %v\n"+
		"🚫 Max Idle Closed: %d\n"+
		"⏰ Max Lifetime Closed: %d\n```",
		stats.MaxOpenConnections,
		stats.OpenConnections,
		stats.InUse,
		stats.Idle,
		stats.WaitCount,
		stats.WaitDuration,
		stats.MaxIdleClosed,
		stats.MaxLifetimeClosed)

	payload := map[string]interface{}{
		"content": message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord payload: %w", err)
	}

	resp, err := http.Post(discordWebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send to Discord: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Discord API returned status code: %d", resp.StatusCode)
	}

	return nil
}

func Connect(cfg *config.Config) error {
	log.Info("Connecting to database...")
	maxRetries := 5
	gormCfg := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
		// Schema migrations change result shapes; do not cache prepared statements.
		PrepareStmt:       false,
		AllowGlobalUpdate: false,
	}

	var err error
	for i := 0; i < maxRetries; i++ {
		if cfg.Database.InstanceConnectionName != "" {
			err = connectWithCloudSQL(cfg, gormCfg)
		} else {
			err = connectWithDSN(cfg, gormCfg)
		}

		if err == nil {
			log.Success("Connected to database")
			return nil
		}

		log.Warn("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(time.Second * 5)
	}
	return log.Error("failed to connect to database after %d attempts", fmt.Errorf("failed to connect to database after %d attempts", maxRetries))
}

func connectWithDSN(cfg *config.Config, gormCfg *gorm.Config) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	log.Info("Using direct Postgres connection with DSN")

	gormDB, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}), gormCfg)
	if err != nil {
		return err
	}

	DB = gormDB
	if err := finalizeConnection(); err != nil {
		sqlDB, dbErr := DB.DB()
		if dbErr == nil {
			sqlDB.Close()
		}
		return err
	}

	return nil
}

func connectWithCloudSQL(cfg *config.Config, gormCfg *gorm.Config) error {
	ctx := context.Background()
	opts := []cloudsqlconn.Option{}
	if cfg.Database.UseIAMAuth {
		opts = append(opts, cloudsqlconn.WithIAMAuthN())
	}

	dialer, err := cloudsqlconn.NewDialer(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create Cloud SQL dialer: %w", err)
	}

	dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	)
	if cfg.Database.UseIAMAuth {
		dsn = fmt.Sprintf("user=%s dbname=%s sslmode=disable",
			cfg.Database.User,
			cfg.Database.Name,
		)
	}

	log.Info("Using Cloud SQL Connector for instance %s (IAM auth: %t)", cfg.Database.InstanceConnectionName, cfg.Database.UseIAMAuth)

	pgxConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		dialer.Close()
		return fmt.Errorf("failed to parse pgx config: %w", err)
	}

	// Conn is supplied below, so the GORM driver cannot configure pgx for us.
	// Match the direct connection: no implicit prepared-statement cache.
	pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pgxConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.Dial(ctx, cfg.Database.InstanceConnectionName)
	}

	sqlDB := stdlib.OpenDB(*pgxConfig)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), gormCfg)
	if err != nil {
		dialer.Close()
		sqlDB.Close()
		return fmt.Errorf("failed to open gorm connection via Cloud SQL: %w", err)
	}

	DB = gormDB
	cloudSQLDialer = dialer

	if err := finalizeConnection(); err != nil {
		sqlDB.Close()
		dialer.Close()
		cloudSQLDialer = nil
		return err
	}

	return nil
}

func finalizeConnection() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return log.Error("Failed to get underlying *sql.DB instance", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(300)                 // Maximum number of open connections to the database
	sqlDB.SetMaxIdleConns(5)                   // Maximum number of idle connections in the pool
	sqlDB.SetConnMaxLifetime(time.Hour)        // Maximum amount of time a connection may be reused
	sqlDB.SetConnMaxIdleTime(time.Minute * 15) // Maximum amount of time a connection may be idle

	// Run migrations
	if err := runMigrations(); err != nil {
		return log.Error("Failed to run migrations", err)
	}

	log.Success("Migrations completed")
	return nil
}

func runMigrations() error {
	log.Info("Running migrations...")
	// Begin transaction for migrations
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Rollback on every unsuccessful exit, including panics. After commit this
	// is a harmless no-op; panics must propagate to the caller.
	defer tx.Rollback()

	if err := tx.AutoMigrate(
		// Base models without foreign keys
		&models.User{},
		&models.Team{},
		&models.BrandingSettings{},
		&models.Resource{},
		&models.Model{},

		// Models with single foreign key dependencies
		&models.PasswordReset{},
		&models.TeamSettings{},
		&models.Contact{},
		&models.Tag{},
		&models.MailingList{},
		&models.SMTPConfig{},
		&models.Domain{},
		&models.Webhook{},
		&models.Template{},
		&models.APIKey{},
		&models.TeamInvite{},
		&models.RateLimit{},
		&models.AuthTransaction{},
		&models.Campaign{},
		&models.Newsletter{}, &models.ContactNote{}, &models.FormReceipt{},

		// Subscriber models
		&models.ContactImport{},

		// Email-related models
		&models.Email{},
		&models.EmailTracking{},
		&models.Delivery{},

		// Permission models
		&models.UserPermission{},
		&models.ResourcePermission{},
		&models.APIKeyPermission{},

		// Usage and monitoring
		&models.APIKeyUsage{},

		// Automation models
		&models.Automation{},
		&models.AutomationNode{},
		&models.AutomationNodeEdge{},
		&models.AutomationExecution{},
		&models.AIOptimization{},
		&models.LLMEmailWriterJob{},

		// Segmentation models
		&models.Segment{},
		&models.SegmentContact{},

		// Lead scoring models
		&models.LeadScore{},
		&models.ScoreActivity{},
		&models.ScoreRule{},

		// A/B testing models
		&models.ABTest{},
		&models.TestVariant{},
		&models.TestResult{},
		&models.VariantAssignment{},

		// Goal tracking models
		&models.AutomationGoal{},
		&models.GoalConversion{},

		// Bounce management models
		&models.EmailBounce{},
		&models.SuppressionList{},
		&models.ComplaintReport{},
		&models.BounceRule{},

		// Form & Landing Page models
		&models.Form{},
		&models.FormField{},
		&models.FormSubmission{},
		&models.FormRevision{}, &models.FormProgress{}, &models.FormAnalyticsEvent{}, &models.FormCompletionEvent{},
		&models.LandingPage{},
		&models.FormABTest{},

		// Send Time Optimization models
		&models.ContactEngagementPattern{},
		&models.EmailOpenEvent{},
		&models.TeamSendTimeDefaults{},
		&models.SendTimeQueue{},

		// Subscription models

		// IMAP models
		&models.IMAPConfig{},
	); err != nil {
		tx.Rollback()
		return err
	}

	if err := models.BackfillTagWorkspaces(tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := analytics.Install(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func Close() error {
	var closeErr error

	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			closeErr = err
		} else if err := sqlDB.Close(); err != nil {
			closeErr = err
		}
	}

	if cloudSQLDialer != nil {
		cloudSQLDialer.Close()
		cloudSQLDialer = nil
	}

	return closeErr
}

func GetDB() *gorm.DB {
	return DB
}

// MonitorConnectionPool logs database connection pool statistics periodically
func MonitorConnectionPool(interval time.Duration) {
	sqlDB, err := DB.DB()
	if err != nil {
		log.Error("Failed to get underlying *sql.DB instance", err)
		return
	}

	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			stats := sqlDB.Stats()

			// Log to console
			log.Info("DB Connection Pool Stats:")
			log.Info("  Max Open Connections: %d", stats.MaxOpenConnections)
			log.Info("  Open Connections: %d", stats.OpenConnections)
			log.Info("  In Use Connections: %d", stats.InUse)
			log.Info("  Idle Connections: %d", stats.Idle)
			log.Info("  Waited for Connection: %d", stats.WaitCount)
			log.Info("  Wait Duration: %v", stats.WaitDuration)
			log.Info("  Max Idle Closed: %d", stats.MaxIdleClosed)
			log.Info("  Max Lifetime Closed: %d", stats.MaxLifetimeClosed)

			// Send to Discord if configured
			if err := sendToDiscord(stats); err != nil {
				log.Error("Failed to send stats to Discord", err)
			}
		}
	}()
}

// GetConnectionStats returns the current database connection pool statistics
func GetConnectionStats() (*sql.DBStats, error) {
	sqlDB, err := DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying *sql.DB instance: %w", err)
	}
	stats := sqlDB.Stats()
	return &stats, nil
}
