package models

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSeedSuperAdminSystemTemplatesBackfillsOnlyRootAndPreservesCustomizations(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}, &EmailCategory{}, &Template{}))
	rootTeam, regularTeam := "root-team", "regular-team"
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&User{Base: Base{ID: "root"}, TeamID: rootTeam, Email: "root@example.com", Role: UserRoleSuperAdmin}).Error)
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&User{Base: Base{ID: "admin"}, TeamID: regularTeam, Email: "admin@example.com", Role: UserRoleAdmin}).Error)
	category := EmailCategory{Base: Base{ID: "root-category"}, TeamID: rootTeam, Name: "Transactional"}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&category).Error)
	custom := Template{Base: Base{ID: "custom"}, TeamID: rootTeam, CategoryID: category.ID, StarterKey: "xem-welcome", Name: "Customized welcome", Subject: "Keep this", HTMLBody: "custom", DesignJSON: "custom"}
	require.NoError(t, db.Create(&custom).Error)
	deleted := Template{Base: Base{ID: "deleted", IsDeleted: true}, TeamID: rootTeam, CategoryID: category.ID, StarterKey: "xem-password_changed", Name: "Deleted password notice", Subject: "Do not restore", HTMLBody: "deleted", DesignJSON: "deleted"}
	require.NoError(t, db.Create(&deleted).Error)
	require.NoError(t, SeedSuperAdminSystemTemplates(db))
	require.NoError(t, SeedSuperAdminSystemTemplates(db))
	var rootCount, regularCount int64
	require.NoError(t, db.Model(&Template{}).Where("team_id = ? AND starter_key LIKE ?", rootTeam, "xem-%").Count(&rootCount).Error)
	require.EqualValues(t, 20, rootCount)
	var activeCount int64
	require.NoError(t, db.Model(&Template{}).Where("team_id = ? AND starter_key LIKE ? AND is_deleted = false", rootTeam, "xem-%").Count(&activeCount).Error)
	require.EqualValues(t, 19, activeCount)
	require.NoError(t, db.Model(&Template{}).Where("team_id = ?", regularTeam).Count(&regularCount).Error)
	require.Zero(t, regularCount)
	var preserved Template
	require.NoError(t, db.Where("id = ?", "custom").First(&preserved).Error)
	require.Equal(t, "Keep this", preserved.Subject)
	require.Equal(t, "custom", preserved.HTMLBody)
	var generated Template
	require.NoError(t, db.Where("team_id = ? AND starter_key = ?", rootTeam, "xem-domain_added").First(&generated).Error)
	require.NotEmpty(t, generated.HTMLBody)
	require.NotEmpty(t, generated.DesignJSON)
	require.Equal(t, category.ID, generated.CategoryID)
	decoded, err := base64.StdEncoding.DecodeString(generated.DesignJSON)
	require.NoError(t, err)
	require.True(t, json.Valid(decoded))
}

func TestSeedSuperAdminSystemTemplatesFreshSharedWorkspace(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}, &EmailCategory{}, &Template{}))
	for _, id := range []string{"root-a", "root-b"} {
		require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&User{Base: Base{ID: id}, TeamID: "shared-root-team", Email: id + "@example.com", Role: UserRoleSuperAdmin}).Error)
	}
	require.NoError(t, SeedSuperAdminSystemTemplates(db))
	var templates, categories int64
	require.NoError(t, db.Model(&Template{}).Where("team_id = ?", "shared-root-team").Count(&templates).Error)
	require.EqualValues(t, 20, templates)
	require.NoError(t, db.Model(&EmailCategory{}).Where("team_id = ? AND LOWER(name) = ?", "shared-root-team", "transactional").Count(&categories).Error)
	require.EqualValues(t, 1, categories)
}

func TestSeedSuperAdminSystemTemplatesDoesNotHideIDConflict(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}, &EmailCategory{}, &Template{}))
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&User{Base: Base{ID: "root"}, TeamID: "root-team", Email: "root@example.com", Role: UserRoleSuperAdmin}).Error)
	category := EmailCategory{Base: Base{ID: "category"}, TeamID: "root-team", Name: "Transactional"}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&category).Error)
	collision := Template{Base: Base{ID: rootSystemTemplateID("root-team", "xem-domain_added")}, TeamID: "root-team", CategoryID: category.ID, StarterKey: "unrelated", Name: "Unrelated", Subject: "Unrelated"}
	require.NoError(t, db.Create(&collision).Error)
	require.ErrorContains(t, SeedSuperAdminSystemTemplates(db), "verify system template xem-domain_added")
}

func TestSeedSuperAdminSystemTemplatesDoesNotUseCollidingCategory(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}, &EmailCategory{}, &Template{}))
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&User{Base: Base{ID: "root"}, TeamID: "root-team", Email: "root@example.com", Role: UserRoleSuperAdmin}).Error)
	collision := EmailCategory{Base: Base{ID: rootSystemTemplateCategoryID("root-team")}, TeamID: "other-team", Name: "Unrelated"}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&collision).Error)
	require.ErrorContains(t, SeedSuperAdminSystemTemplates(db), "verify transactional category")
	var templates int64
	require.NoError(t, db.Model(&Template{}).Where("team_id = ?", "root-team").Count(&templates).Error)
	require.Zero(t, templates)
}

func TestSeedSuperAdminSystemTemplatesPreservesDeletedCategory(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&User{}, &EmailCategory{}, &Template{}))
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&User{Base: Base{ID: "root"}, TeamID: "root-team", Email: "root@example.com", Role: UserRoleSuperAdmin}).Error)
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&EmailCategory{Base: Base{ID: "deleted-category", IsDeleted: true}, TeamID: "root-team", Name: "Transactional"}).Error)
	require.ErrorContains(t, SeedSuperAdminSystemTemplates(db), "preserving administrator intent")
	var templates int64
	require.NoError(t, db.Model(&Template{}).Where("team_id = ?", "root-team").Count(&templates).Error)
	require.Zero(t, templates)
}
