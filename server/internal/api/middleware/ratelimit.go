package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	// Redis client for storing rate limit data
	RedisClient *redis.Client

	// Default rate limits
	DefaultLimit rate.Limit
	DefaultBurst int

	// Endpoint-specific limits
	EndpointLimits map[string]EndpointLimit

	// IP-based limits
	IPLimits map[string]IPLimit

	// Authentication-based limits
	AuthLimits map[string]AuthLimit
}

// EndpointLimit defines rate limits for specific endpoints
type EndpointLimit struct {
	Limit  rate.Limit
	Burst  int
	Window time.Duration
}

// IPLimit defines rate limits for specific IP addresses
type IPLimit struct {
	Limit  rate.Limit
	Burst  int
	Window time.Duration
}

// AuthLimit defines rate limits for authenticated users
type AuthLimit struct {
	Limit  rate.Limit
	Burst  int
	Window time.Duration
}

// RateLimitResponse holds rate limit information
type RateLimitResponse struct {
	Limit      int       `json:"limit"`
	Remaining  int       `json:"remaining"`
	Reset      time.Time `json:"reset"`
	RetryAfter int       `json:"retry_after,omitempty"`
}

// Default rate limit configurations
var defaultEndpointLimits = map[string]EndpointLimit{
	// Authentication endpoints - stricter limits
	"POST:/auth/login": {
		Limit:  5.0 / 60.0, // 5 requests per minute
		Burst:  3,
		Window: time.Minute,
	},
	"POST:/auth/register": {
		Limit:  3.0 / 3600.0, // 3 requests per hour
		Burst:  1,
		Window: time.Hour,
	},
	"POST:/auth/password-reset": {
		Limit:  3.0 / 3600.0, // 3 requests per hour
		Burst:  1,
		Window: time.Hour,
	},
	"POST:/auth/refresh": {
		Limit:  10.0 / 60.0, // 10 requests per minute
		Burst:  5,
		Window: time.Minute,
	},

	// Analytics endpoints - moderate limits
	"GET:/api/v1/analytics": {
		Limit:  100.0 / 60.0, // 100 requests per minute
		Burst:  50,
		Window: time.Minute,
	},

	// Email sending - moderate limits
	"POST:/email": {
		Limit:  100.0 / 60.0, // 100 requests per minute
		Burst:  50,
		Window: time.Minute,
	},

	// File upload - stricter limits
	"POST:/api/v1/files/upload": {
		Limit:  20.0 / 60.0, // 20 requests per minute
		Burst:  10,
		Window: time.Minute,
	},

	// Tracking endpoints - high limits (used by email clients)
	"GET:/api/v1/t/open": {
		Limit:  1000.0 / 60.0, // 1000 requests per minute
		Burst:  500,
		Window: time.Minute,
	},
	"GET:/api/v1/t/click": {
		Limit:  1000.0 / 60.0, // 1000 requests per minute
		Burst:  500,
		Window: time.Minute,
	},
}

// RateLimiter creates a new rate limiting middleware
func RateLimiter(config RateLimitConfig) echo.MiddlewareFunc {
	// Set default values if not provided
	if config.DefaultLimit == 0 {
		config.DefaultLimit = 100.0 / 60.0 // 100 requests per minute
	}
	if config.DefaultBurst == 0 {
		config.DefaultBurst = 50
	}
	if config.EndpointLimits == nil {
		config.EndpointLimits = defaultEndpointLimits
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip rate limiting for health check endpoints
			path := c.Request().URL.Path
			if path == "/health" || path == "/api/v1/health" {
				return next(c)
			}

			// Get client identifier (IP address or user ID)
			clientID := getClientID(c)

			// Get endpoint key
			endpointKey := getEndpointKey(c)

			// Get rate limit configuration for this endpoint
			limitConfig := getLimitConfig(endpointKey, config)

			// Check rate limit
			allowed, remaining, reset, retryAfter := checkRateLimit(
				c.Request().Context(),
				config.RedisClient,
				clientID,
				endpointKey,
				limitConfig,
			)

			// Set rate limit headers
			setRateLimitHeaders(c, limitConfig.Limit, remaining, reset)

			if !allowed {
				// Return rate limit exceeded response
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error":       "rate_limit_exceeded",
					"message":     "Rate limit exceeded. Try again later.",
					"retry_after": retryAfter,
					"limit":       int(float64(limitConfig.Limit) * limitConfig.Window.Seconds()),
					"window":      limitConfig.Window.Seconds(),
				})
			}

			return next(c)
		}
	}
}

// getClientID returns a unique identifier for the client
func getClientID(c echo.Context) string {
	// Try to get user ID from JWT token first
	if userID := getUserIDFromContext(c); userID != "" {
		return fmt.Sprintf("user:%s", userID)
	}

	// Fall back to IP address
	ip := getClientIP(c)
	return fmt.Sprintf("ip:%s", ip)
}

// getUserIDFromContext extracts user ID from JWT context
func getUserIDFromContext(c echo.Context) string {
	// This would need to be implemented based on your JWT middleware
	// For now, we'll return empty string and use IP-based limiting
	return ""
}

// getClientIP returns the client's IP address
func getClientIP(c echo.Context) string {
	// Check for X-Forwarded-For header (for proxy setups)
	if forwardedFor := c.Request().Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check for X-Real-IP header
	if realIP := c.Request().Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to remote address
	return c.RealIP()
}

// getEndpointKey returns a unique key for the endpoint
func getEndpointKey(c echo.Context) string {
	method := c.Request().Method
	path := c.Request().URL.Path

	// Normalize path to remove dynamic parts
	path = normalizePath(path)

	return fmt.Sprintf("%s:%s", method, path)
}

// normalizePath removes dynamic parts from the path for better grouping
func normalizePath(path string) string {
	// Replace UUIDs with placeholder
	path = strings.ReplaceAll(path, "/", "/")

	// Replace common dynamic patterns
	path = strings.ReplaceAll(path, "/api/v1/analytics/campaign", "/api/v1/analytics/campaign")
	path = strings.ReplaceAll(path, "/api/v1/analytics/email", "/api/v1/analytics/email")
	path = strings.ReplaceAll(path, "/auth/users/", "/auth/users/")

	return path
}

// getLimitConfig returns the rate limit configuration for an endpoint
func getLimitConfig(endpointKey string, config RateLimitConfig) EndpointLimit {
	// Check for endpoint-specific limit
	if limit, exists := config.EndpointLimits[endpointKey]; exists {
		return limit
	}

	// Return default limit
	return EndpointLimit{
		Limit:  config.DefaultLimit,
		Burst:  config.DefaultBurst,
		Window: time.Minute,
	}
}

// checkRateLimit checks if the request is within rate limits
func checkRateLimit(
	ctx context.Context,
	redisClient *redis.Client,
	clientID string,
	endpointKey string,
	limitConfig EndpointLimit,
) (allowed bool, remaining int, reset time.Time, retryAfter int) {
	// Create Redis key
	key := fmt.Sprintf("rate_limit:%s:%s", clientID, endpointKey)

	// Get current timestamp
	now := time.Now()

	// Use Redis pipeline for atomic operations
	pipe := redisClient.Pipeline()

	// Get current count
	getCmd := pipe.Get(ctx, key)

	// Set expiry if key doesn't exist
	pipe.Expire(ctx, key, limitConfig.Window)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		// Redis error, allow request but log error
		return true, int(float64(limitConfig.Limit) * limitConfig.Window.Seconds()), now.Add(limitConfig.Window), 0
	}

	// Get current count
	var currentCount int
	if getCmd.Val() != "" {
		currentCount, _ = strconv.Atoi(getCmd.Val())
	}

	// Check if limit exceeded
	limit := int(float64(limitConfig.Limit) * limitConfig.Window.Seconds())
	if currentCount >= limit {
		// Use window duration as retry after
		retryAfter = int(limitConfig.Window.Seconds())
		return false, 0, now.Add(limitConfig.Window), retryAfter
	}

	// Increment counter
	pipe = redisClient.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, limitConfig.Window)
	pipe.Exec(ctx)

	// Get new count
	newCount := int(incrCmd.Val())
	remaining = limit - newCount

	// Calculate reset time
	reset = now.Add(limitConfig.Window)

	return true, remaining, reset, 0
}

// setRateLimitHeaders sets rate limit headers in the response
func setRateLimitHeaders(c echo.Context, limit rate.Limit, remaining int, reset time.Time) {
	limitInt := int(float64(limit) * 60) // Convert to requests per minute

	c.Response().Header().Set("X-RateLimit-Limit", strconv.Itoa(limitInt))
	c.Response().Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Response().Header().Set("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
}

// IPBasedRateLimiter creates a rate limiter specifically for IP addresses
func IPBasedRateLimiter(config RateLimitConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := getClientIP(c)

			// Check if this IP has specific limits
			if ipLimit, exists := config.IPLimits[ip]; exists {
				allowed, remaining, reset, retryAfter := checkRateLimit(
					c.Request().Context(),
					config.RedisClient,
					fmt.Sprintf("ip:%s", ip),
					"ip_limit",
					EndpointLimit{
						Limit:  ipLimit.Limit,
						Burst:  ipLimit.Burst,
						Window: ipLimit.Window,
					},
				)

				setRateLimitHeaders(c, ipLimit.Limit, remaining, reset)

				if !allowed {
					return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
						"error":       "rate_limit_exceeded",
						"message":     "IP rate limit exceeded. Try again later.",
						"retry_after": retryAfter,
					})
				}
			}

			return next(c)
		}
	}
}

// AuthBasedRateLimiter creates a rate limiter for authenticated users
func AuthBasedRateLimiter(config RateLimitConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Only apply to authenticated endpoints
			if !isAuthenticatedEndpoint(c) {
				return next(c)
			}

			userID := getUserIDFromContext(c)
			if userID == "" {
				return next(c)
			}

			// Check if this user has specific limits
			if authLimit, exists := config.AuthLimits[userID]; exists {
				allowed, remaining, reset, retryAfter := checkRateLimit(
					c.Request().Context(),
					config.RedisClient,
					fmt.Sprintf("user:%s", userID),
					"auth_limit",
					EndpointLimit{
						Limit:  authLimit.Limit,
						Burst:  authLimit.Burst,
						Window: authLimit.Window,
					},
				)

				setRateLimitHeaders(c, authLimit.Limit, remaining, reset)

				if !allowed {
					return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
						"error":       "rate_limit_exceeded",
						"message":     "User rate limit exceeded. Try again later.",
						"retry_after": retryAfter,
					})
				}
			}

			return next(c)
		}
	}
}

// isAuthenticatedEndpoint checks if the endpoint requires authentication
func isAuthenticatedEndpoint(c echo.Context) bool {
	path := c.Request().URL.Path

	// List of endpoints that require authentication
	authEndpoints := []string{
		"/api/v1/",
		"/email",
		"/auth/me",
		"/auth/users",
	}

	for _, endpoint := range authEndpoints {
		if strings.HasPrefix(path, endpoint) {
			return true
		}
	}

	return false
}

// CreateDefaultRateLimitConfig creates a default rate limit configuration
func CreateDefaultRateLimitConfig(redisClient *redis.Client) RateLimitConfig {
	return RateLimitConfig{
		RedisClient:    redisClient,
		DefaultLimit:   100.0 / 60.0, // 100 requests per minute
		DefaultBurst:   50,
		EndpointLimits: defaultEndpointLimits,
		IPLimits:       make(map[string]IPLimit),
		AuthLimits:     make(map[string]AuthLimit),
	}
}
