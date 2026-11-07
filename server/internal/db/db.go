package db

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"kori/internal/billing"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"kori/internal/analytics"
	"kori/internal/config"
	"kori/internal/hygiene"
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
		// SQL diagnostics must never interpolate credentials or mailbox content.
		Logger: logger.New(stdlog.New(os.Stdout, "", stdlog.LstdFlags), logger.Config{
			LogLevel: logger.Warn, SlowThreshold: time.Second, ParameterizedQueries: true,
		}),
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

func databaseDSN(cfg config.DatabaseConfig) string {
	quote := func(value string) string {
		return "'" + strings.NewReplacer("\\", "\\\\", "'", "\\'").Replace(value) + "'"
	}
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		quote(cfg.Host), quote(cfg.User), quote(cfg.Password), quote(cfg.Name), cfg.Port, quote(cfg.SSLMode))
}

func connectWithDSN(cfg *config.Config, gormCfg *gorm.Config) error {
	dsn := databaseDSN(cfg.Database)

	log.Info("Using direct Postgres connection with DSN")

	gormDB, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}), gormCfg)
	if err != nil {
		return err
	}

	DB = gormDB
	if err := finalizeConnection(cfg.Database.Name); err != nil {
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

	cloudConfig := cfg.Database
	cloudConfig.Host, cloudConfig.Port, cloudConfig.SSLMode = "localhost", 5432, "disable"
	if cfg.Database.UseIAMAuth {
		cloudConfig.Password = ""
	}
	dsn := databaseDSN(cloudConfig)

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

	if err := finalizeConnection(cfg.Database.Name); err != nil {
		sqlDB.Close()
		dialer.Close()
		cloudSQLDialer = nil
		return err
	}

	return nil
}

func finalizeConnection(expectedDatabase string) error {
	sqlDB, err := DB.DB()
	if err != nil {
		return log.Error("Failed to get underlying *sql.DB instance", err)
	}
	var actualDatabase string
	if err := DB.Raw("SELECT current_database()").Scan(&actualDatabase).Error; err != nil {
		return fmt.Errorf("cannot verify selected database: %w", err)
	}
	if actualDatabase != expectedDatabase {
		return fmt.Errorf("refusing to migrate unexpected database %q; configured database is %q", actualDatabase, expectedDatabase)
	}

	open, idle, lifetime, err := poolSettings()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(open)
	sqlDB.SetMaxIdleConns(idle)
	sqlDB.SetConnMaxLifetime(lifetime)
	sqlDB.SetConnMaxIdleTime(time.Minute * 15) // Maximum amount of time a connection may be idle

	// Run migrations
	if err := runMigrations(); err != nil {
		return log.Error("Failed to run migrations", err)
	}

	log.Success("Migrations completed")
	return nil
}

// poolSettings reads DB_MAX_OPEN_CONNS (default 50), DB_MAX_IDLE_CONNS (default
// the open limit, so busy pools do not close and reopen connections) and
// DB_CONN_MAX_LIFETIME (default 30m). Every replica opens up to its own limit:
// keep replicas x DB_MAX_OPEN_CONNS (plus workers and migrations) under the
// server's max_connections, or put PgBouncer in transaction mode in front. That
// is safe because connections use the simple protocol and GORM's PrepareStmt is
// off, so no prepared statement outlives a transaction.
func poolSettings() (open, idle int, lifetime time.Duration, err error) {
	open, idle, lifetime = 50, -1, 30*time.Minute
	for _, v := range []struct {
		key string
		dst *int
	}{{"DB_MAX_OPEN_CONNS", &open}, {"DB_MAX_IDLE_CONNS", &idle}} {
		if raw := os.Getenv(v.key); raw != "" {
			if *v.dst, err = strconv.Atoi(raw); err != nil || *v.dst <= 0 {
				return 0, 0, 0, fmt.Errorf("invalid %s", v.key)
			}
		}
	}
	if idle < 0 {
		idle = open
	}
	if raw := os.Getenv("DB_CONN_MAX_LIFETIME"); raw != "" {
		if lifetime, err = time.ParseDuration(raw); err != nil || lifetime <= 0 {
			return 0, 0, 0, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME")
		}
	}
	return open, idle, lifetime, nil
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

	firstVersions := !tx.Migrator().HasTable(&models.AutomationVersion{})
	firstCleanup := !tx.Migrator().HasColumn(&models.Contact{}, "cleanup_since")
	if err := tx.AutoMigrate(
		// Base models without foreign keys
		&models.User{},
		&models.Team{},
		&models.BrandingSettings{},
		&models.Resource{},
		&models.Model{},

		// Models with single foreign key dependencies
		&models.PasswordReset{},
		&models.ServiceNotice{}, &models.ServiceNoticeEvent{},
		&billing.EmailUsage{},
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
		&models.Newsletter{}, &models.NewsletterSignup{}, &models.ContactNote{}, &models.FormReceipt{}, &models.TrackingConsent{}, &models.ContactEvent{},

		// Subscriber models
		&models.ContactImport{},
		&hygiene.NeverEmailEntry{},

		// Email-related models
		&models.Email{},
		&models.EmailContent{},
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
		&models.AutomationVersion{},
		&models.AIOptimization{},
		&models.LLMEmailWriterJob{},

		// Segmentation models
		&models.Segment{},
		&models.SegmentContact{},

		// Affiliate program
		&models.Affiliate{},
		&models.AffiliateClick{},
		&models.AffiliateReferral{},
		&models.AffiliateCommission{},
		&models.AffiliatePayout{},

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
		&models.MailConnection{},
		&models.MailOAuthState{},
		&models.CloudflareRelay{},
	); err != nil {
		tx.Rollback()
		return err
	}

	if err := models.BackfillTagWorkspaces(tx); err != nil {
		tx.Rollback()
		return err
	}
	if firstVersions {
		if err := models.BackfillAutomationVersions(tx); err != nil {
			return err
		}
	}
	// List cleanup must not judge readers on mail sent before it existed.
	if firstCleanup {
		if err := tx.Model(&models.Contact{}).Where("1 = 1").UpdateColumn("cleanup_since", time.Now().UTC()).Error; err != nil {
			return err
		}
	}
	if err := analytics.Install(tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := campaignEmailIndex(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// campaignEmailIndex speeds up the campaign sender's check for contacts that
// already have an email in the campaign. It isn't unique: recurring campaigns
// and API sends tagged with a campaign may email the same contact again.
func campaignEmailIndex(tx *gorm.DB) error {
	return tx.Exec("CREATE INDEX IF NOT EXISTS emails_campaign_contact ON emails (campaign_id, contact_id) WHERE campaign_id IS NOT NULL").Error
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
