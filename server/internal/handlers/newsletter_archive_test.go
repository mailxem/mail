package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/models"
)

func TestArchiveListsEditionsAndPublishedCampaignsOnly(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Newsletter{}, &models.Campaign{}, &models.Form{}))
	team, other := uuid.NewString(), uuid.NewString()
	slug := "founder-notes"
	n := models.Newsletter{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Founder Notes", ListID: uuid.NewString(), PublicSlug: &slug}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&n).Error)
	add := func(subject, teamID string, status models.CampaignStatus, edition, archive *string, at time.Time) {
		c := models.Campaign{Base: models.Base{ID: uuid.NewString()}, TeamID: teamID, Name: subject, Subject: subject, Status: status, NewsletterID: edition, ArchiveNewsletterID: archive, TemplateID: uuid.NewString(), ListID: uuid.NewString(), ScheduledFor: at}
		require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&c).Error)
	}
	now := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	add("Edition one", team, models.CampaignStatusCompleted, &n.ID, nil, now)
	add("Launch announcement", team, models.CampaignStatusCompleted, nil, &n.ID, now.Add(24*time.Hour))
	add("Draft announcement", team, models.CampaignStatusDraft, nil, &n.ID, now.Add(48*time.Hour))
	add("Not published", team, models.CampaignStatusCompleted, nil, nil, now.Add(72*time.Hour))
	add("Other workspace", other, models.CampaignStatusCompleted, nil, &n.ID, now.Add(96*time.Hour))

	h := NewMarketingHandler(db, "secret", "https://xem.email")
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(httptest.NewRequest("GET", "/", nil), rec)
	c.SetParamNames("slug")
	c.SetParamValues(slug)
	require.NoError(t, h.PublicNewsletter(c))
	var body struct {
		Issues []struct{ Subject string } `json:"issues"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	var subjects []string
	for _, i := range body.Issues {
		subjects = append(subjects, i.Subject)
	}
	require.Equal(t, []string{"Launch announcement", "Edition one"}, subjects)
}
