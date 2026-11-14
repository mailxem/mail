package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/models"
	"kori/internal/sending"
)

func TestAlertsAreScopedReadAndResolved(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent), DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Alert{}, &models.ServiceNotice{}, &models.ServiceNoticeEvent{}, &models.Campaign{}, &models.ContactImport{}, &sending.Domain{}, &sending.Account{}))
	team, other := uuid.NewString(), uuid.NewString()
	now := time.Now().UTC()
	quiet := db.Session(&gorm.Session{SkipHooks: true})

	campaign := models.Campaign{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Launch", Status: models.CampaignStatusPaused, TemplateID: uuid.NewString(), ListID: uuid.NewString()}
	require.NoError(t, quiet.Create(&campaign).Error)
	require.NoError(t, models.QueueAlert(db, team, "staged-sending", campaign.ID, "sending_paused", "Bounces from bob@example.com", now))
	domain := sending.Domain{ID: uuid.NewString(), TeamID: team, Name: "news.example.com"}
	require.NoError(t, db.Create(&domain).Error)
	require.NoError(t, models.RecordAlert(db, team, "domain_unready", domain.ID, domain.ID, "DNS records no longer verify", 1, now.Add(time.Second)))
	failed := models.ContactImport{Base: models.Base{ID: uuid.NewString(), CreatedAt: now}, TeamID: team, ListID: uuid.NewString(), Status: models.ContactImportStatusFailed}
	require.NoError(t, quiet.Create(&failed).Error)
	require.NoError(t, models.QueueAlert(db, team, "import", failed.ID, "import_failed", "The file has no data rows", now.Add(2*time.Second)))
	// Routine bounces reach the bell only once they would email the owner.
	for i := range 9 {
		require.NoError(t, models.QueueSendingNotice(db, team, "smtp", uuid.NewString(), "BOUNCED", "Mailbox full", now))
		if i == 0 {
			require.NoError(t, models.QueueSendingNotice(db, other, "smtp", uuid.NewString(), "SUSPENDED", "Suspended", now))
		}
	}

	h := NewMarketingHandler(db, "secret", "https://xem.email")
	call := func(as string, handler echo.HandlerFunc, body string) (*httptest.ResponseRecorder, error) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c := echo.New().NewContext(req, rec)
		c.Set("teamID", as)
		return rec, handler(c)
	}
	type listed struct {
		Alerts []struct {
			ID, Kind, Title, Action, Href, Reason string
			Resolved                              bool
			ReadAt                                *time.Time
		}
		Unread int64
	}
	list := func(as string) listed {
		rec, err := call(as, h.Alerts, "")
		require.NoError(t, err)
		var out listed
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
		return out
	}

	got := list(team)
	require.EqualValues(t, 3, got.Unread)
	require.Len(t, got.Alerts, 3, "nine bounces stay below the daily threshold; the other workspace's alert isn't listed")
	byKind := map[string]int{}
	for i, a := range got.Alerts {
		byKind[a.Kind] = i
		require.False(t, a.Resolved, a.Kind)
	}
	paused := got.Alerts[byKind["sending_paused"]]
	require.Equal(t, "Resume campaign", paused.Action)
	require.Equal(t, "/campaigns/"+campaign.ID, paused.Href)
	require.Equal(t, "Bounces from [address]", paused.Reason)
	require.Equal(t, "/audience/lists/"+failed.ListID, got.Alerts[byKind["import_failed"]].Href)
	require.Equal(t, "Check DNS", got.Alerts[byKind["domain_unready"]].Action)

	require.NoError(t, models.QueueSendingNotice(db, team, "smtp", uuid.NewString(), "BOUNCED", "Mailbox full", now))
	got = list(team)
	require.Len(t, got.Alerts, 4, "the tenth bounce of the day shows")

	// Read state: one, then all, and never another workspace's.
	theirs := list(other).Alerts[0].ID
	_, err = call(team, h.ReadAlerts, `{"id":"`+theirs+`"}`)
	require.Equal(t, 404, err.(*echo.HTTPError).Code)
	require.EqualValues(t, 1, list(other).Unread)
	_, err = call(team, h.ReadAlerts, `{"id":"`+paused.ID+`"}`)
	require.NoError(t, err)
	got = list(team)
	require.EqualValues(t, 3, got.Unread)
	for _, a := range got.Alerts {
		require.Equal(t, a.ID == paused.ID, a.ReadAt != nil, a.Kind)
	}
	_, err = call(team, h.ReadAlerts, `{}`)
	require.NoError(t, err)
	require.Zero(t, list(team).Unread)
	require.EqualValues(t, 1, list(other).Unread)

	// Resolved state follows the thing each alert is about.
	require.NoError(t, db.Model(&campaign).Update("status", models.CampaignStatusSending).Error)
	require.NoError(t, db.Model(&domain).Update("ready", true).Error)
	retry := models.ContactImport{Base: models.Base{ID: uuid.NewString(), CreatedAt: now.Add(time.Minute)}, TeamID: team, ListID: failed.ListID, Status: models.ContactImportStatusCompleted}
	require.NoError(t, quiet.Create(&retry).Error)
	for _, a := range list(team).Alerts {
		require.Equal(t, a.Kind != "sending_bounced", a.Resolved, a.Kind)
	}
	require.NoError(t, db.Create(&sending.Account{TeamID: other, Suspended: true}).Error)
	require.False(t, list(other).Alerts[0].Resolved)
	require.NoError(t, db.Model(&sending.Account{}).Where("team_id = ?", other).Update("suspended", false).Error)
	require.True(t, list(other).Alerts[0].Resolved)
}
