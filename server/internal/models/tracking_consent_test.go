package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestShouldTrackConsentModes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	off := false

	// Lookup error (no tables yet): marketing sends fail closed, transactional ignores consent.
	require.False(t, ShouldTrack(db, uuid.NewString(), "a@example.com", nil, true))
	require.True(t, ShouldTrack(db, uuid.NewString(), "a@example.com", nil, false))
	require.False(t, ShouldTrack(db, uuid.NewString(), "a@example.com", &off, false))

	require.NoError(t, db.AutoMigrate(&Team{}, &TrackingConsent{}))
	team := uuid.NewString()
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&Team{Base: Base{ID: team}, Name: "Acme"}).Error)
	require.True(t, ShouldTrack(db, team, "a@example.com", nil, true), "mode none follows the switch")
	require.False(t, ShouldTrack(db, team, "a@example.com", &off, true))

	require.NoError(t, db.Model(&Team{}).Where("id = ?", team).Update("tracking_consent_mode", TrackingConsentOptIn).Error)
	require.False(t, ShouldTrack(db, team, "a@example.com", nil, true), "opt_in without consent")
	require.True(t, ShouldTrack(db, team, "a@example.com", nil, false), "transactional ignores consent mode")
	require.NoError(t, SetTrackingConsent(db, team, " A@Example.com ", true, "preferences_page"))
	require.True(t, ShouldTrack(db, team, "a@example.com", nil, true))
	require.False(t, ShouldTrack(db, team, "a@example.com", &off, true), "switch off still wins")

	require.NoError(t, db.Model(&Team{}).Where("id = ?", team).Update("tracking_consent_mode", TrackingConsentOptOut).Error)
	require.True(t, ShouldTrack(db, team, "b@example.com", nil, true), "opt_out without objection")
	require.NoError(t, SetTrackingConsent(db, team, "b@example.com", false, "preferences_page"))
	require.False(t, ShouldTrack(db, team, "b@example.com", nil, true), "opt_out after objection")

	var row TrackingConsent
	require.NoError(t, db.First(&row, "team_id = ? AND email = ?", team, "b@example.com").Error)
	require.Equal(t, "preferences_page", row.Source)
	require.False(t, row.ChangedAt.IsZero())
}
