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

func TestCampaignReportCountsDeliveryPeopleAndStages(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Campaign{}, &models.Email{}, &models.EmailTracking{}, &models.Contact{}))
	team, list := uuid.NewString(), uuid.NewString()
	staged := true
	camp := models.Campaign{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: list, Name: "Launch", Status: models.CampaignStatusSending, SendInStages: &staged, Timezone: "UTC", Schedule: models.CampaignScheduleOneTime}
	quiet := db.Session(&gorm.Session{SkipHooks: true})
	require.NoError(t, quiet.Create(&camp).Error)
	for i := 0; i < 50; i++ {
		c := models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: list, Email: uuid.NewString() + "@example.com", Status: models.SubscriberStatusActive}
		require.NoError(t, quiet.Create(&c).Error)
		if i >= 3 {
			continue
		}
		status := []models.EmailStatus{models.EmailStatusSent, models.EmailStatusBounced, models.EmailStatusSent}[i]
		e := models.Email{Base: models.Base{ID: uuid.NewString()}, TeamID: team, CampaignID: camp.ID, ContactID: c.ID, To: c.Email, Status: status, SentAt: time.Now(), ClickTrackingEnabled: &staged}
		require.NoError(t, quiet.Create(&e).Error)
		if i == 0 {
			// A person opens; a scanner clicks. Only the open counts.
			require.NoError(t, quiet.Create(&models.EmailTracking{Base: models.Base{ID: uuid.NewString()}, EmailID: e.ID, Event: models.EmailTrackingEventOpen, Timestamp: time.Now()}).Error)
			require.NoError(t, quiet.Create(&models.EmailTracking{Base: models.Base{ID: uuid.NewString()}, EmailID: e.ID, Event: models.EmailTrackingEventClick, URL: "https://example.com", Timestamp: time.Now(), Machine: models.TrackingMachineBot}).Error)
		}
	}
	// A test send is not part of the campaign's numbers.
	require.NoError(t, quiet.Create(&models.Email{Base: models.Base{ID: uuid.NewString()}, TeamID: team, CampaignID: camp.ID, To: "me@example.com", Status: models.EmailStatusSent, Test: true}).Error)

	h := NewMarketingHandler(db, "secret", "https://api.xem.email")
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(httptest.NewRequest("GET", "/", nil), rec)
	c.Set("teamID", team)
	c.SetParamNames("id")
	c.SetParamValues(camp.ID)
	require.NoError(t, h.CampaignReport(c))
	var r campaignReport
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &r))
	require.EqualValues(t, 3, r.Delivery.Total)
	require.EqualValues(t, 2, r.Delivery.Sent)
	require.EqualValues(t, 1, r.Delivery.Bounced)
	require.Equal(t, 1, r.Engagement.UniqueOpens)
	require.Equal(t, 0, r.Engagement.UniqueClicks)
	require.Equal(t, 1, r.Engagement.BotClicks)
	require.InDelta(t, 100.0/3, r.Engagement.OpenRate, 0.01)
	require.NotNil(t, r.Stages)
	require.EqualValues(t, 50, r.Stages.New, "nobody was emailed before, so all 50 are new")
	require.EqualValues(t, 3, r.Stages.SentToNew)
	require.Equal(t, [2]int{1, 3}, r.Stages.Bounds)

	c = echo.New().NewContext(httptest.NewRequest("GET", "/", nil), httptest.NewRecorder())
	c.Set("teamID", uuid.NewString())
	c.SetParamNames("id")
	c.SetParamValues(camp.ID)
	require.Error(t, h.CampaignReport(c), "another workspace's campaign is not found")
}
