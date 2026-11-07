package handlers

import (
	"net/http"
	"strings"
	"time"

	"kori/internal/db"
	"kori/internal/models"

	"github.com/labstack/echo/v4"
)

var apiKeyLifetimes = map[string]time.Duration{
	"7d":  7 * 24 * time.Hour,
	"30d": 30 * 24 * time.Hour,
	"90d": 90 * 24 * time.Hour,
	"1y":  365 * 24 * time.Hour,
}

// CreateAPIKey creates a key that expires after expiresIn ("never", "7d", "30d",
// "90d" or "1y"; 90 days when empty). Only workspace admins signed in to the
// app can create an admin key.
func CreateAPIKey(c echo.Context) error {
	var in struct {
		Name      string `json:"name"`
		ExpiresIn string `json:"expiresIn"`
		Admin     bool   `json:"admin"`
	}
	if c.Bind(&in) != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid API key")
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || len(in.Name) > 100 {
		return echo.NewHTTPError(http.StatusBadRequest, "Give the key a name of up to 100 characters")
	}
	teamID, _ := c.Get("teamID").(string)
	key := models.APIKey{Name: in.Name, TeamID: teamID, Admin: in.Admin}
	switch in.ExpiresIn {
	case "never":
		key.NeverExpires = true
	case "":
		key.ExpiresAt = time.Now().UTC().Add(apiKeyLifetimes["90d"])
	default:
		d, ok := apiKeyLifetimes[in.ExpiresIn]
		if !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "Choose no expiry, 7 days, 30 days, 90 days or 1 year")
		}
		key.ExpiresAt = time.Now().UTC().Add(d)
	}
	if in.Admin {
		isAPIKey, _ := c.Get("isAPIKey").(bool)
		role, _ := c.Get("role").(string)
		if isAPIKey || (role != string(models.UserRoleAdmin) && role != string(models.UserRoleSuperAdmin)) {
			return echo.NewHTTPError(http.StatusForbidden, "Only workspace admins can create admin API keys")
		}
	}
	if err := db.DB.WithContext(c.Request().Context()).Create(&key).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to create the API key")
	}
	return c.JSON(http.StatusCreated, key)
}
