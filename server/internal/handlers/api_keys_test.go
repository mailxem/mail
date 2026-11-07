package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"kori/internal/db"
	"kori/internal/models"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCreateAPIKeyExpiryAndAdmin(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, database.AutoMigrate(&models.APIKey{}, &models.Resource{}, &models.ResourcePermission{}, &models.APIKeyPermission{}))
	db.DB = database
	team := "11111111-1111-1111-1111-111111111111"

	create := func(body, role string) (int, models.APIKey) {
		req := httptest.NewRequest("POST", "/api-keys", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		c.Set("teamID", team)
		c.Set("role", role)
		c.Set("isAPIKey", false)
		if err := CreateAPIKey(c); err != nil {
			return err.(*echo.HTTPError).Code, models.APIKey{}
		}
		var key models.APIKey
		require.NoError(t, database.Preload("Permissions").Order("created_at DESC").First(&key).Error)
		return rec.Code, key
	}

	code, key := create(`{"name":"forever","expiresIn":"never"}`, "MEMBER")
	require.Equal(t, http.StatusCreated, code)
	require.True(t, key.ExpiresAt.IsZero(), "no expiry stays zero")
	require.False(t, key.Admin)

	code, key = create(`{"name":"week","expiresIn":"7d"}`, "MEMBER")
	require.Equal(t, http.StatusCreated, code)
	require.WithinDuration(t, time.Now().Add(7*24*time.Hour), key.ExpiresAt, time.Minute)

	code, _ = create(`{"name":"bad","expiresIn":"2w"}`, "MEMBER")
	require.Equal(t, http.StatusBadRequest, code)

	code, _ = create(`{"name":"sneaky","admin":true}`, "MEMBER")
	require.Equal(t, http.StatusForbidden, code)

	code, key = create(`{"name":"ops","expiresIn":"1y","admin":true}`, string(models.UserRoleAdmin))
	require.Equal(t, http.StatusCreated, code)
	require.True(t, key.Admin)
	require.Len(t, key.Permissions, len(models.AdminKeyResources))
}
