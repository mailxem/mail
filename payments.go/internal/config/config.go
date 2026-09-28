package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type PlanConfig struct {
	FreePlanID string
}

type Config struct {
	// Server
	Port        string
	Environment string

	// Database
	DatabaseURL string

	// Dodo Payments
	DodoTestMode      bool
	DodoWebhookSecret string

	// Security
	JWTSecret   string
	AdminSecret string

	// CORS
	AllowedOrigins []string

	Plans             PlanConfig
	DashboardURL      string
	KoriEmailAPIURL   string
	KoriEmailAPIKey   string
	KoriEmailProvider string

	Templates TemplatesConfig

	POSTGRES_MAX_OPEN_CONNS int
}

type TemplatesConfig struct {
	FreeTrialStart               string
	FreeTrialEndIsGoingToEndSoon string
	FreeTrialEnded               string
	SubscriptionStatusChanged    string
	ProPlanActivated             string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:              getEnv("PORT", "8080"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/diagonal_payments?sslmode=disable"),
		DodoTestMode:      getEnvBool("DODO_TEST_MODE", true),
		DodoWebhookSecret: getEnv("DODO_WEBHOOK_SECRET", ""),
		JWTSecret:         getEnv("JWT_SECRET", "your-secret-key"),
		AdminSecret:       getEnv("ADMIN_SECRET", ""),
		AllowedOrigins: []string{
			getEnv("FRONTEND_URL", "http://localhost:3000"),
		},
		Plans: PlanConfig{
			FreePlanID: getEnv("FREE_PLAN_ID", "starter-plan-001"),
		},
		POSTGRES_MAX_OPEN_CONNS: getEnvInt("POSTGRES_MAX_OPEN_CONNS", 100),
		DashboardURL:            getEnv("DASHBOARD_URL", "http://localhost:3000"),
		KoriEmailAPIURL:         getEnv("KORI_EMAIL_API_URL", "https://api.posthoot.com/v1"),
		KoriEmailAPIKey:         getEnv("KORI_EMAIL_API_KEY", ""),
		KoriEmailProvider:       getEnv("KORI_EMAIL_PROVIDER", "AMAZON"),
		Templates: TemplatesConfig{
			FreeTrialStart:               getEnv("FREE_TRIAL_START_TEMPLATE_ID", "afad3b75-dfe0-4104-9896-175ea9375998"),
			FreeTrialEndIsGoingToEndSoon: getEnv("FREE_TRIAL_END_IS_GOING_TO_END_SOON_TEMPLATE_ID", "c7cd10b6-bad4-4f13-a842-e4d8e5eed615"),
			FreeTrialEnded:               getEnv("FREE_TRIAL_END_TEMPLATE_ID", "c57ce46d-de66-41c2-8bfb-b33aa552d980"),
			SubscriptionStatusChanged:    getEnv("SUBSCRIPTION_STATUS_CHANGED_TEMPLATE_ID", "18cd1ab7-1b87-4454-b2f0-b75c5850f36b"),
			ProPlanActivated:             getEnv("PRO_PLAN_ACTIVATED_TEMPLATE_ID", "44463897-deb2-4289-900e-db95ef42dfeb"),
		},
	}

	// Validate critical security configuration in production
	if cfg.Environment == "production" {
		if err := validateProductionConfig(cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

// validateProductionConfig ensures all critical secrets are properly set in production
func validateProductionConfig(cfg *Config) error {
	// Check for weak or default secrets
	weakSecrets := []string{"your-secret-key", "secret", "password", "admin", "test"}

	if cfg.JWTSecret == "" || len(cfg.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters in production")
	}
	for _, weak := range weakSecrets {
		if cfg.JWTSecret == weak {
			return fmt.Errorf("JWT_SECRET cannot be a default/weak value in production")
		}
	}

	if cfg.AdminSecret == "" || len(cfg.AdminSecret) < 32 {
		return fmt.Errorf("ADMIN_SECRET must be at least 32 characters in production")
	}
	for _, weak := range weakSecrets {
		if cfg.AdminSecret == weak {
			return fmt.Errorf("ADMIN_SECRET cannot be a default/weak value in production")
		}
	}

	if cfg.DodoWebhookSecret == "" {
		return fmt.Errorf("DODO_WEBHOOK_SECRET is required in production")
	}

	// Database URL should use SSL in production
	if !contains(cfg.DatabaseURL, "sslmode=require") && !contains(cfg.DatabaseURL, "sslmode=verify-") {
		log.Printf("WARNING: Database connection should use SSL in production (sslmode=require)")
	}

	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
