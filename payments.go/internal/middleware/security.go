package middleware

import (
	"crypto/subtle"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// SecurityHeaders adds security-related HTTP headers
func SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Prevent clickjacking
			c.Response().Header().Set("X-Frame-Options", "DENY")

			// Prevent MIME type sniffing
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")

			// Enable XSS protection (legacy browsers)
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

			// Referrer policy - send origin only for cross-origin requests
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Content Security Policy
			// Strict CSP for API - no scripts/styles should be needed
			c.Response().Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

			// Permissions Policy - disable unnecessary features
			c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=()")

			return next(c)
		}
	}
}

// HSTS adds HTTP Strict Transport Security header (only in production)
func HSTS(enabled bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if enabled {
				// Enforce HTTPS for 1 year, include subdomains
				c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}
			return next(c)
		}
	}
}

// AdminAuth validates admin secret with constant-time comparison
func AdminAuth(adminSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			providedSecret := c.Request().Header.Get("X-Admin-Secret")

			// Constant-time comparison to prevent timing attacks
			if subtle.ConstantTimeCompare([]byte(providedSecret), []byte(adminSecret)) != 1 {
				// Log failed authentication attempt
				log.Printf("SECURITY: Failed admin authentication from IP: %s, Path: %s", c.RealIP(), c.Request().URL.Path)

				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "unauthorized",
				})
			}

			// Log successful admin access
			log.Printf("SECURITY: Admin access granted to IP: %s, Path: %s", c.RealIP(), c.Request().URL.Path)

			return next(c)
		}
	}
}

// RequestSizeLimit limits the size of request bodies to prevent memory exhaustion
func RequestSizeLimit(maxBytes int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, maxBytes)
			return next(c)
		}
	}
}

// AuditLog logs security-sensitive operations
type AuditLogger struct {
	enabled bool
}

func NewAuditLogger(enabled bool) *AuditLogger {
	return &AuditLogger{enabled: enabled}
}

func (a *AuditLogger) LogEvent(event, ip, userID, resource string, metadata map[string]interface{}) {
	if !a.enabled {
		return
	}

	log.Printf("AUDIT: event=%s ip=%s user=%s resource=%s metadata=%v",
		event, ip, userID, resource, metadata)
}

// AuditMiddleware logs all requests to sensitive endpoints
func AuditMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Log request
			log.Printf("AUDIT: method=%s path=%s ip=%s",
				c.Request().Method,
				c.Request().URL.Path,
				c.RealIP())

			return next(c)
		}
	}
}
