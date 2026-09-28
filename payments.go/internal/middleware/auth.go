package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// TeamAuth validates that the requesting team has access to the resource
// Expects X-Team-ID header to match the resource being accessed
func TeamAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get team ID from header
			teamIDHeader := c.Request().Header.Get("X-Team-ID")
			if teamIDHeader == "" {
				log.Printf("SECURITY: Missing X-Team-ID header from IP: %s, Path: %s", c.RealIP(), c.Request().URL.Path)
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "missing team authentication",
				})
			}

			// Get team ID from path parameter
			teamIDParam := c.Param("teamId")
			if teamIDParam != "" {
				// Validate team ID matches
				if teamIDHeader != teamIDParam {
					log.Printf("SECURITY: Team ID mismatch - Header: %s, Param: %s, IP: %s",
						teamIDHeader, teamIDParam, c.RealIP())
					return c.JSON(http.StatusForbidden, map[string]string{
						"error": "access denied",
					})
				}
			}

			// Store team ID in context for handlers
			c.Set("team_id", teamIDHeader)

			return next(c)
		}
	}
}

// APIKeyAuth validates API key authentication
// Expects X-API-Key header with valid API key
func APIKeyAuth(validAPIKeys map[string]string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apiKey := c.Request().Header.Get("X-API-Key")
			if apiKey == "" {
				// Also check Authorization header with Bearer scheme
				authHeader := c.Request().Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					apiKey = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if apiKey == "" {
				log.Printf("SECURITY: Missing API key from IP: %s, Path: %s", c.RealIP(), c.Request().URL.Path)
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "missing API key",
				})
			}

			// Validate API key
			teamID, valid := validAPIKeys[apiKey]
			if !valid {
				log.Printf("SECURITY: Invalid API key from IP: %s, Path: %s", c.RealIP(), c.Request().URL.Path)
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid API key",
				})
			}

			// Store team ID in context
			c.Set("team_id", teamID)

			return next(c)
		}
	}
}
