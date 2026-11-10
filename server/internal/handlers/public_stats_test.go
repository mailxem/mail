package handlers

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/models"
)

func TestPublicStatsCountRealSendsOnly(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent), DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Email{}, &models.ComplaintReport{}))
	now := time.Now().UTC()
	team := uuid.NewString()
	add := func(status string, test bool, at time.Time) {
		e := models.Email{Base: models.Base{ID: uuid.NewString()}, TeamID: team, To: "a@b.co", Status: models.EmailStatus(status), Test: test, SMTPConfigID: uuid.NewString(), CategoryID: uuid.NewString()}
		require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&e).Error)
		require.NoError(t, db.Model(&e).UpdateColumn("created_at", at).Error)
	}
	for range 8 {
		add("SENT", false, now)
	}
	add("OPENED", false, now)
	add("CLICKED", false, now.Add(-40*24*time.Hour)) // all time only
	add("BOUNCED", false, now)
	add("SENT", true, now) // tests don't count
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&models.ComplaintReport{Base: models.Base{ID: uuid.NewString()}, TeamID: team, EmailAddress: "a@b.co"}).Error)

	s, err := publicStats(db, now)
	require.NoError(t, err)
	require.EqualValues(t, 10, s.Delivered)
	require.EqualValues(t, 9, s.Delivered30d)
	require.InDelta(t, 0.9, s.DeliveryRate, 1e-9)
	require.InDelta(t, 1.0/9, s.ComplaintRate, 1e-9)
}
