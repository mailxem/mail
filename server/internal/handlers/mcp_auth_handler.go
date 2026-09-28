package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"kori/internal/assistant"
	"kori/internal/models"
)

// MCPAuthorize verifies a workspace API key without requiring a resource grant.
// Tool endpoints still enforce their own resource permissions on every call.
// This endpoint returns no identity, permissions, or credential data.
func MCPAuthorize(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-store")
		key := c.Request().Header.Get("X-API-Key")
		if key == "" || len(key) > 4096 {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key")
		}
		var row models.APIKey
		if err := db.WithContext(c.Request().Context()).Model(&models.APIKey{}).
			Where("key = ? AND is_deleted = false AND team_id IS NOT NULL", assistant.Lookup(key)).First(&row).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(401, "Invalid API key")
			}
			return echo.NewHTTPError(503, "Authentication unavailable")
		}
		// Compare actual instants after loading. SQLite stores offsets as text,
		// so a SQL comparison against UTC can accept expired local-time keys.
		if !row.ExpiresAt.IsZero() && !row.ExpiresAt.After(time.Now()) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key")
		}
		if !assistant.ValidateActor(db.WithContext(c.Request().Context()), &row) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key")
		}
		return c.NoContent(http.StatusNoContent)
	}
}
