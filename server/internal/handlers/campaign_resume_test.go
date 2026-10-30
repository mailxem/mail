package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/models"
)

func TestResumeOnlyContinuesOwnPausedCampaignsWithoutStages(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Campaign{}))
	team, other := uuid.NewString(), uuid.NewString()
	add := func(teamID string, status models.CampaignStatus) models.Campaign {
		c := models.Campaign{Base: models.Base{ID: uuid.NewString()}, TeamID: teamID, Name: "Launch", Status: status, PauseReason: "Paused after 9 of 100 emails failed", TemplateID: uuid.NewString(), ListID: uuid.NewString()}
		require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&c).Error)
		return c
	}
	paused, sending, foreign := add(team, models.CampaignStatusPaused), add(team, models.CampaignStatusSending), add(other, models.CampaignStatusPaused)
	h := NewMarketingHandler(db, "secret", "https://xem.email")
	resume := func(id string) error {
		c := echo.New().NewContext(httptest.NewRequest("POST", "/", nil), httptest.NewRecorder())
		c.Set("teamID", team)
		c.SetParamNames("id")
		c.SetParamValues(id)
		return h.ResumeCampaign(c)
	}
	require.NoError(t, resume(paused.ID))
	var got models.Campaign
	require.NoError(t, db.First(&got, "id = ?", paused.ID).Error)
	require.Equal(t, models.CampaignStatusSending, got.Status)
	require.Empty(t, got.PauseReason)
	require.NotNil(t, got.SendInStages)
	require.False(t, *got.SendInStages)
	require.Error(t, resume(sending.ID), "only paused campaigns resume")
	require.Error(t, resume(foreign.ID), "another workspace's campaign is not found")
}
