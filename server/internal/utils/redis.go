package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"time"

	"kori/internal/config"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis/Valkey client with additional functionality
// Works with both Redis and Valkey (Redis-compatible)
type RedisClient struct {
	*redis.Client
	Type string // "redis" or "valkey"
}

// NewRedisClient creates a new Redis/Valkey client
// Both Redis and Valkey use the same wire protocol and are fully compatible
func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	clientType := cfg.Redis.Type
	if clientType == "" {
		clientType = "redis" // default
	}

	opts := &redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		Username: cfg.Redis.Username,
		DB:       cfg.Redis.DB,

		// Connection pool settings
		PoolSize:     10,
		MinIdleConns: 5,

		// Timeout settings
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,

		// Retry settings
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	}

	// Enable TLS if configured (required for ElastiCache Serverless)
	if cfg.Redis.UseTLS {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	client := redis.NewClient(opts)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", clientType, err)
	}

	return &RedisClient{
		Client: client,
		Type:   clientType,
	}, nil
}

// Close closes the Redis client
func (r *RedisClient) Close() error {
	return r.Client.Close()
}

// HealthCheck checks if Redis is healthy
func (r *RedisClient) HealthCheck(ctx context.Context) error {
	return r.Ping(ctx).Err()
}

// GetRateLimitKey returns a Redis key for rate limiting
func (r *RedisClient) GetRateLimitKey(clientID, endpointKey string) string {
	return fmt.Sprintf("rate_limit:%s:%s", clientID, endpointKey)
}

// GetRateLimitCount gets the current count for a rate limit key
func (r *RedisClient) GetRateLimitCount(ctx context.Context, key string) (int, error) {
	val, err := r.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// IncrementRateLimit increments the rate limit counter
func (r *RedisClient) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int, error) {
	pipe := r.Pipeline()

	// Increment the counter
	incrCmd := pipe.Incr(ctx, key)

	// Set expiry
	pipe.Expire(ctx, key, window)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return int(incrCmd.Val()), nil
}

// GetRateLimitTTL gets the TTL for a rate limit key
func (r *RedisClient) GetRateLimitTTL(ctx context.Context, key string) (time.Duration, error) {
	return r.TTL(ctx, key).Result()
}

// ClearRateLimit clears a rate limit key
func (r *RedisClient) ClearRateLimit(ctx context.Context, key string) error {
	return r.Del(ctx, key).Err()
}

// GetRateLimitStats gets statistics about rate limiting
func (r *RedisClient) GetRateLimitStats(ctx context.Context, pattern string) (map[string]int, error) {
	keys, err := r.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]int)
	for _, key := range keys {
		count, err := r.GetRateLimitCount(ctx, key)
		if err != nil {
			continue
		}
		stats[key] = count
	}

	return stats, nil
}
