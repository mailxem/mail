package rate

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimit struct {
	Window  time.Duration // e.g., 1 minute, 1 hour
	MaxJobs int           // max jobs per window
}

type QueueConfig struct {
	Name      string
	RateLimit RateLimit
}

type QueueRateLimiter struct {
	redis  *redis.Client
	config QueueConfig
}

func NewQueueRateLimiter(redis *redis.Client, config QueueConfig) *QueueRateLimiter {
	return &QueueRateLimiter{
		redis:  redis,
		config: config,
	}
}

func (qrl *QueueRateLimiter) Allow(ctx context.Context, identifier string) (bool, error) {
	key := fmt.Sprintf("queue_rate_limit:%s:%s", qrl.config.Name, identifier)

	pipe := qrl.redis.Pipeline()
	now := time.Now().Unix()
	windowStart := now - int64(qrl.config.RateLimit.Window.Seconds())

	// Remove old entries
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))

	// Count current window
	pipe.ZCard(ctx, key)

	// Add new entry
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

	// Set expiration
	pipe.Expire(ctx, key, qrl.config.RateLimit.Window*2)

	results, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("redis pipeline error: %w", err)
	}

	count := results[1].(*redis.IntCmd).Val()
	return count <= int64(qrl.config.RateLimit.MaxJobs), nil
}
