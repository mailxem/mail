package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"kori/internal/models"
	"net/http"
	"strings"
)

// AnalyticsAccess applies resource permissions and ownership to every report,
// including legacy routes and exports. Query parameters never select a tenant.
func AnalyticsAccess(db *gorm.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			team := GetTeamID(c)
			if team == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Sign in to view analytics")
			}
			if requested := c.QueryParam("teamId"); requested != "" && requested != team {
				return echo.NewHTTPError(http.StatusForbidden, "Workspace mismatch")
			}
			admin, _ := c.Get("hasAdminAccess").(bool)
			if !admin {
				if IsAPIKey(c) {
					id, _ := c.Get("apiKeyID").(string)
					if err := ValidateAPIKeyPermissions(c.Request().Context(), db, id, []string{"analytics:read"}); err != nil {
						return err
					}
				} else {
					var count int64
					err := db.WithContext(c.Request().Context()).Table("user_permissions up").
						Joins("JOIN resource_permissions rp ON rp.id = up.resource_permission_id AND rp.is_deleted = false").
						Joins("JOIN resources r ON r.id = rp.resource_id AND r.is_deleted = false").
						Where("up.user_id = ? AND up.is_deleted = false AND r.name = 'analytics' AND r.action IN ('read','admin')", GetUserID(c)).Count(&count).Error
					if err != nil {
						return echo.NewHTTPError(http.StatusInternalServerError, "Unable to verify analytics access")
					}
					if count == 0 {
						return echo.NewHTTPError(http.StatusForbidden, "Missing analytics:read permission")
					}
				}
			}
			for _, resource := range []struct {
				key   string
				model interface{}
			}{
				{"emailId", &models.Email{}}, {"campaignId", &models.Campaign{}}, {"campaignIds", &models.Campaign{}},
				{"listId", &models.MailingList{}}, {"tagId", &models.Tag{}},
			} {
				raw := c.QueryParam(resource.key)
				if raw == "" {
					continue
				}
				ids := strings.Split(raw, ",")
				if len(ids) > 20 {
					return echo.NewHTTPError(400, "Select at most 20 resources")
				}
				for _, id := range ids {
					if _, err := uuid.Parse(id); err != nil {
						return echo.NewHTTPError(400, "Invalid "+resource.key)
					}
					var n int64
					if err := db.WithContext(c.Request().Context()).Model(resource.model).Where("id = ? AND team_id = ? AND is_deleted = false", id, team).Count(&n).Error; err != nil {
						return err
					}
					if n != 1 {
						return echo.NewHTTPError(404, "Analytics resource not found")
					}
				}
			}
			q := c.Request().URL.Query()
			q.Set("teamId", team)
			c.Request().URL.RawQuery = q.Encode()
			c.Response().Header().Set("Cache-Control", "private, no-store")
			return next(c)
		}
	}
}
