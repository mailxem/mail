package middleware

import (
	"context"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/assistant"
	appdb "kori/internal/db"
	"kori/internal/models"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAssistantCredentials(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	raw, _ := db.DB()
	t.Cleanup(func() { raw.Close() })
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.APIKey{}, &models.Resource{}, &models.ResourcePermission{}, &models.APIKeyPermission{}, &models.APIKeyUsage{}))
	old := appdb.DB
	appdb.DB = db
	t.Cleanup(func() { appdb.DB = old })
	user := models.User{Base: models.Base{ID: uuid.NewString()}, TeamID: uuid.NewString(), Role: models.UserRoleAdmin, Email: "admin@example.test", Password: "unused"}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&user).Error)
	key, expires, err := assistant.Issue(db, user)
	require.NoError(t, err)
	require.WithinDuration(t, time.Now().Add(5*time.Minute), expires, time.Second)
	var stored models.APIKey
	require.NoError(t, db.Where("key = ?", assistant.Lookup(key)).First(&stored).Error)
	require.False(t, stored.IsDeleted, "new key unexpectedly deleted; expires %v", stored.ExpiresAt)
	require.NotEqual(t, key, stored.Key)
	require.NotEqual(t, stored.Key, assistant.Lookup(stored.Key))
	require.NoError(t, ValidateAPIKeyPermissions(context.Background(), db, stored.ID, []string{"contacts:read", "contacts:create", "sending:create"}))
	require.Error(t, ValidateAPIKeyPermissions(context.Background(), db, stored.ID, []string{"api_keys:create"}))
	require.Error(t, ValidateAPIKeyPermissions(context.Background(), db, stored.ID, []string{"users:create"}))
	check := func(value string) error {
		c := echo.New().NewContext(httptest.NewRequest("GET", "/api/v1/contacts", nil), httptest.NewRecorder())
		return NewAuthMiddleware("test").validateAPIKey(c, value, func(c echo.Context) error {
			require.Equal(t, user.TeamID, c.Get("teamID"))
			require.Equal(t, user.ID, c.Get("userID"))
			return nil
		})
	}
	require.NoError(t, check(key))
	require.Error(t, check(stored.Key))
	require.NoError(t, db.Model(&models.User{}).Where("id = ?", user.ID).UpdateColumn("role", models.UserRoleMember).Error)
	require.Error(t, check(key), "demotion must invalidate a write credential")
	user.Role = models.UserRoleMember
	// Legacy API-key hooks normally grant email writes. Assistant read keys
	// must not inherit that default even when the legacy resource is present.
	resource := models.Resource{Name: "emails", Action: "create"}
	require.NoError(t, db.Create(&resource).Error)
	require.NoError(t, db.Create(&models.ResourcePermission{ResourceID: resource.ID, Scope: "emails:create"}).Error)
	readKey, _, err := assistant.Issue(db, user)
	require.NoError(t, err)
	var read models.APIKey
	require.NoError(t, db.Where("key = ?", assistant.Lookup(readKey)).First(&read).Error)
	require.NoError(t, ValidateAPIKeyPermissions(context.Background(), db, read.ID, []string{"contacts:read"}))
	require.Error(t, ValidateAPIKeyPermissions(context.Background(), db, read.ID, []string{"contacts:create"}))
	require.Error(t, ValidateAPIKeyPermissions(context.Background(), db, read.ID, []string{"emails:create"}))
	require.NoError(t, check(readKey))
	require.NoError(t, db.Model(&models.User{}).Where("id = ?", user.ID).UpdateColumn("team_id", uuid.NewString()).Error)
	require.Error(t, check(readKey), "changing teams must invalidate old credentials")
}
