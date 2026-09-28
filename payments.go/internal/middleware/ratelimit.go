package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// RateLimiter stores rate limiters per IP
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	cleanup  time.Duration
}

// NewRateLimiter creates a new rate limiter
// rps: requests per second, burst: burst size
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(rps),
		burst:    burst,
		cleanup:  5 * time.Minute,
	}

	// Cleanup old entries periodically
	go rl.cleanupLoop()

	return rl
}

// getLimiter returns limiter for given IP
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
		rl.mu.Unlock()
	}

	return limiter
}

// cleanupLoop removes old limiters to prevent memory leak
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		// Simple cleanup: clear all limiters periodically
		// More sophisticated: track last access time
		rl.limiters = make(map[string]*rate.Limiter)
		rl.mu.Unlock()
	}
}

// Middleware returns Echo middleware for rate limiting
func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			limiter := rl.getLimiter(ip)

			if !limiter.Allow() {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "rate limit exceeded",
				})
			}

			return next(c)
		}
	}
}

// StrictRateLimiter returns strict rate limiting (lower limits for sensitive endpoints)
func StrictRateLimiter() echo.MiddlewareFunc {
	// 10 requests per minute (0.166 rps) with burst of 3
	limiter := NewRateLimiter(0.166, 3)
	return limiter.Middleware()
}

// ModerateRateLimiter returns moderate rate limiting for normal endpoints
func ModerateRateLimiter() echo.MiddlewareFunc {
	// 60 requests per minute (1 rps) with burst of 10
	limiter := NewRateLimiter(1.0, 10)
	return limiter.Middleware()
}

// GenerousRateLimiter returns generous rate limiting for high-traffic endpoints
func GenerousRateLimiter() echo.MiddlewareFunc {
	// 300 requests per minute (5 rps) with burst of 20
	limiter := NewRateLimiter(5.0, 20)
	return limiter.Middleware()
}
