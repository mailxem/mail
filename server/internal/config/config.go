package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	JWT        JWTConfig
	Storage    StorageConfig
	Worker     WorkerConfig
	Redis      RedisConfig
	S3         S3Config
	Crypto     CryptoConfig
	SMTP       SMTPConfig
	Monitor    MonitorConfig
	Airley     AirleyConfig
	AI         AIConfig
	Automation AutomationConfig
}

type CryptoConfig struct {
	PrivateKey string
}

type ServerConfig struct {
	Host      string
	Port      int
	PublicURL string
}

type DatabaseConfig struct {
	Host                   string
	Port                   int
	User                   string
	Password               string
	Name                   string
	SSLMode                string
	InstanceConnectionName string
	UseIAMAuth             bool
}

type JWTConfig struct {
	Secret string
}

type StorageConfig struct {
	Provider string // local, s3, etc.
	BasePath string
	S3       S3Config
}

type S3Config struct {
	BucketName string `env:"S3_BUCKET_NAME" required:"true"`
	Endpoint   string `env:"S3_ENDPOINT"`
	Region     string `env:"S3_REGION" required:"true"`
	AccessKey  string `env:"S3_ACCESS_KEY" required:"true"`
	SecretKey  string `env:"S3_SECRET_KEY" required:"true"`
}

type SMTPConfig struct {
	Host      string `env:"SMTP_HOST" required:"true"`
	Port      int    `env:"SMTP_PORT" required:"true"`
	User      string `env:"SMTP_USER" required:"true"`
	Password  string `env:"SMTP_PASSWORD" required:"true"`
	FromEmail string `env:"SMTP_FROM_EMAIL" required:"true"`
	Provider  string `env:"SMTP_PROVIDER" required:"true"`
}

type WorkerConfig struct {
	Concurrency int
	QueueSize   int
}

type RedisConfig struct {
	Addr     string
	Password string
	Username string
	DB       int
	Type     string // "redis" or "valkey"
	UseTLS   bool
}

type MonitorConfig struct {
	DiscordWebhookURL string
}

type AirleyConfig struct {
	Enabled bool
}

type AIConfig struct {
	Enabled         bool
	Provider        string // "anthropic", "openai"
	AnthropicAPIKey string
	Model           string // "claude-4-opus", "claude-4-sonnet"
	AutoOptimize    bool
	MaxTokens       int
}

type AutomationConfig struct {
	MaxExecutionsPerMinute int
	ExecutionTimeout       int // in seconds
	EnableEventTriggers    bool
}

var (
	config *Config
	once   sync.Once
)

// GetConfig returns the singleton config instance
func GetConfig() *Config {
	once.Do(func() {
		config = &Config{}
		config.JWT.Secret = os.Getenv("JWT_SECRET")
		config.Server.PublicURL = os.Getenv("PUBLIC_URL")
	})
	return config
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:      getEnv("SERVER_HOST", "localhost"),
			Port:      getEnvAsInt("SERVER_PORT", 8080),
			PublicURL: getEnv("PUBLIC_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			Host:                   getEnv("POSTGRES_HOST", "localhost"),
			Port:                   getEnvAsInt("POSTGRES_PORT", 5432),
			User:                   getEnv("POSTGRES_USER", "postgres"),
			Password:               getEnv("POSTGRES_PASSWORD", ""),
			Name:                   getEnv("POSTGRES_DB", "kori"),
			SSLMode:                getEnv("POSTGRES_SSLMODE", "disable"),
			InstanceConnectionName: getEnv("INSTANCE_CONNECTION_NAME", ""),
			UseIAMAuth:             getEnvAsBool("CLOUDSQL_IAM_AUTH", false),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key"),
		},
		Storage: StorageConfig{
			Provider: getEnv("STORAGE_PROVIDER", "local"),
			BasePath: getEnv("STORAGE_BASE_PATH", "./storage"),
			S3: S3Config{
				BucketName: getEnv("S3_BUCKET_NAME", ""),
				Endpoint:   getEnv("S3_ENDPOINT", ""),
				Region:     getEnv("S3_REGION", ""),
				AccessKey:  getEnv("S3_ACCESS_KEY", ""),
				SecretKey:  getEnv("S3_SECRET_KEY", ""),
			},
		},
		Worker: WorkerConfig{
			Concurrency: getEnvAsInt("WORKER_CONCURRENCY", 5),
			QueueSize:   getEnvAsInt("WORKER_QUEUE_SIZE", 100),
		},
		Redis: parseRedisConfig(),
		Crypto: CryptoConfig{
			PrivateKey: getEnv("PRIVATE_KEY", ""),
		},
		SMTP: SMTPConfig{
			Host:      getEnv("SMTP_HOST", ""),
			Port:      getEnvAsInt("SMTP_PORT", 587),
			User:      getEnv("SMTP_USER", ""),
			Password:  getEnv("SMTP_PASSWORD", ""),
			FromEmail: getEnv("SMTP_FROM_EMAIL", ""),
			Provider:  getEnv("SMTP_PROVIDER", "CUSTOM"),
		},
		Monitor: MonitorConfig{
			DiscordWebhookURL: getEnv("DISCORD_WEBHOOK_URL", ""),
		},
		Airley: AirleyConfig{
			Enabled: getEnvAsBool("AIRLEY_ENABLED", false),
		},
		AI: AIConfig{
			Enabled:         getEnvAsBool("AI_ENABLED", false),
			Provider:        getEnv("AI_PROVIDER", "anthropic"),
			AnthropicAPIKey: getEnv("ANTHROPIC_API_KEY", ""),
			Model:           getEnv("AI_MODEL", "claude-4-sonnet"),
			AutoOptimize:    getEnvAsBool("AI_AUTO_OPTIMIZE", false),
			MaxTokens:       getEnvAsInt("AI_MAX_TOKENS", 4096),
		},
		Automation: AutomationConfig{
			MaxExecutionsPerMinute: getEnvAsInt("AUTOMATION_MAX_EXECUTIONS_PER_MINUTE", 100),
			ExecutionTimeout:       getEnvAsInt("AUTOMATION_EXECUTION_TIMEOUT", 1800),
			EnableEventTriggers:    getEnvAsBool("AUTOMATION_ENABLE_EVENT_TRIGGERS", true),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return value == "true"
	}
	return defaultValue
}

// parseRedisConfig parses Redis configuration from REDIS_URL or individual env vars
func parseRedisConfig() RedisConfig {
	// Try REDIS_URL first (supports redis://, rediss://, valkey://, valkeys://)
	if redisURL := getEnv("REDIS_URL", ""); redisURL != "" {
		if parsed, err := url.Parse(redisURL); err == nil {
			config := RedisConfig{
				Type: "redis", // default
			}

			// Determine type and TLS from scheme
			scheme := strings.ToLower(parsed.Scheme)
			switch scheme {
			case "rediss", "valkeys":
				config.UseTLS = true
				if scheme == "valkeys" {
					config.Type = "valkey"
				}
			case "redis":
				config.UseTLS = false
			case "valkey":
				config.UseTLS = false
				config.Type = "valkey"
			}

			// Get host and port
			config.Addr = parsed.Host
			if parsed.Port() == "" {
				config.Addr = fmt.Sprintf("%s:6379", parsed.Hostname())
			}

			// Get credentials
			if parsed.User != nil {
				config.Username = parsed.User.Username()
				if password, ok := parsed.User.Password(); ok {
					config.Password = password
				}
			}

			// Get DB from path (e.g., /0, /1)
			if parsed.Path != "" && len(parsed.Path) > 1 {
				if db, err := strconv.Atoi(strings.TrimPrefix(parsed.Path, "/")); err == nil {
					config.DB = db
				}
			}

			return config
		}
	}

	// Fall back to individual env vars
	return RedisConfig{
		Addr:     fmt.Sprintf("%s:%d", getEnv("REDIS_HOST", "localhost"), getEnvAsInt("REDIS_PORT", 6379)),
		Password: getEnv("REDIS_PASSWORD", ""),
		Username: getEnv("REDIS_USERNAME", ""),
		DB:       getEnvAsInt("REDIS_DB", 0),
		Type:     getEnv("REDIS_TYPE", "redis"),
		UseTLS:   getEnvAsBool("REDIS_USE_TLS", false),
	}
}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
