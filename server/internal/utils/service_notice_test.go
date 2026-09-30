package utils

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	appdb "kori/internal/db"
	"kori/internal/models"
)

func TestSMTPStatusAndAlertCommitTogether(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, database.AutoMigrate(&models.Email{}, &models.ServiceNotice{}, &models.ServiceNoticeEvent{}))
	raw, err := database.DB()
	require.NoError(t, err)
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() { raw.Close() })
	old := appdb.DB
	appdb.DB = database
	t.Cleanup(func() { appdb.DB = old })
	email := models.Email{Base: models.Base{ID: uuid.NewString()}, TeamID: uuid.NewString(), Status: models.EmailStatusPending}
	require.NoError(t, database.Session(&gorm.Session{SkipHooks: true}).Create(&email).Error)
	handler := NewEmailHandler(1)
	handler.NotifyFailures = true
	email.Status = models.EmailStatusFailed
	email.Error = "private provider diagnostic"
	require.NoError(t, handler.UpdateEmail(&email))
	require.NoError(t, handler.UpdateEmail(&email))
	var n models.ServiceNotice
	require.NoError(t, database.First(&n).Error)
	require.Equal(t, "sending_failed", n.Kind)
	require.EqualValues(t, 1, n.EventCount)
	require.NoError(t, database.Migrator().DropTable(&models.ServiceNoticeEvent{}))
	email.Status = "DELIVERY_UNKNOWN"
	require.Error(t, handler.UpdateEmail(&email))
	var saved models.Email
	require.NoError(t, database.First(&saved, "id = ?", email.ID).Error)
	require.Equal(t, models.EmailStatusFailed, saved.Status)
}
