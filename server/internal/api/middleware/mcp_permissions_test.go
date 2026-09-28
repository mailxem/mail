package middleware

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/models"
	"testing"
)

func TestMCPResourceMapping(t *testing.T) {
	m := NewAuthMiddleware("test")
	for path, want := range map[string]string{
		"/api/v1/marketing/contact-batch":         "contacts",
		"/api/v1/marketing/campaign-drafts":       "campaigns",
		"/api/v1/marketing/newsletters/abc/pause": "campaigns",
		"/api/v1/marketing/options":               "lists",
		"/api/v1/marketing/forms":                 "marketing",
		"/api/v1/marketing/forms/abc/analytics":   "marketing",
		"/api/v1/marketing/forms/abc/journey":     "marketing",
		"/api/v1/marketing/form-draft":            "marketing",
		"/api/v1/mailing-lists":                   "lists",
		"/api/v1/analytics/report":                "analytics",
	} {
		require.Equal(t, want, m.getResourceFromPath(path), path)
	}
}
func TestMCPKeyPermissions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	raw, _ := db.DB()
	t.Cleanup(func() { raw.Close() })
	require.NoError(t, db.AutoMigrate(&models.Resource{}, &models.ResourcePermission{}, &models.APIKeyPermission{}))
	keyID := uuid.NewString()
	resource := models.Resource{Base: models.Base{ID: uuid.NewString()}, Name: "contacts", Action: "read"}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&resource).Error)
	permission := models.ResourcePermission{Base: models.Base{ID: uuid.NewString()}, ResourceID: resource.ID}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&permission).Error)
	// Use a map so this test covers the persisted foreign-key contract.
	require.NoError(t, db.Table("api_key_permissions").Create(map[string]any{"id": uuid.NewString(), "is_deleted": false, "key_id": keyID, "resource_permission_id": permission.ID}).Error)
	require.NoError(t, ValidateAPIKeyPermissions(context.Background(), db, keyID, []string{"contacts:read"}))
	require.Error(t, ValidateAPIKeyPermissions(context.Background(), db, keyID, []string{"contacts:create"}))
	require.Error(t, ValidateAPIKeyPermissions(context.Background(), db, keyID, []string{"campaigns:read"}))
	require.NoError(t, db.Model(&resource).UpdateColumn("action", "create").Error)
	require.NoError(t, ValidateAPIKeyPermissions(context.Background(), db, keyID, []string{"contacts:write"}))
	require.NoError(t, db.Model(&resource).UpdateColumn("is_deleted", true).Error)
	require.Error(t, ValidateAPIKeyPermissions(context.Background(), db, keyID, []string{"contacts:read"}))
}
