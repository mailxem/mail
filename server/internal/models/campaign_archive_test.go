package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCampaignArchiveMustBelongToTheSameWorkspace(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Newsletter{}))
	team, other := uuid.NewString(), uuid.NewString()
	own := Newsletter{Base: Base{ID: uuid.NewString()}, TeamID: team, Name: "Ours"}
	foreign := Newsletter{Base: Base{ID: uuid.NewString()}, TeamID: other, Name: "Theirs"}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&own).Error)
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&foreign).Error)

	ok := Campaign{TeamID: team, ArchiveNewsletterID: &own.ID}
	require.NoError(t, ok.checkArchiveNewsletter(db, true))
	require.NoError(t, ok.BeforeUpdate(db))

	stolen := Campaign{TeamID: team, ArchiveNewsletterID: &foreign.ID}
	require.EqualError(t, stolen.checkArchiveNewsletter(db, true), "archive newsletter not found")
	require.EqualError(t, stolen.BeforeUpdate(db), "archive newsletter not found")

	// Deleting the newsletter later must not block updates to the campaign.
	require.NoError(t, db.Model(&own).Update("is_deleted", true).Error)
	require.Error(t, ok.checkArchiveNewsletter(db, true))
	require.NoError(t, ok.BeforeUpdate(db))

	empty := ""
	none := Campaign{TeamID: team, ArchiveNewsletterID: &empty}
	require.NoError(t, none.checkArchiveNewsletter(db, true))
	require.Nil(t, none.ArchiveNewsletterID)
}
