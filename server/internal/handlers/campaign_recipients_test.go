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
	"kori/internal/marketing"
	"kori/internal/models"
)

func TestRecipientOutcomes(t *testing.T) {
	sent := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	off := false
	emails := []recipientEmail{
		{ID: "delivered-then-bounced", To: "a@x.com", Status: "DELIVERED", SentAt: sent},
		{ID: "bounced", To: "b@x.com", Status: "BOUNCED", SentAt: sent},
		{ID: "machines-only", To: "c@x.com", Status: "SENT", SentAt: sent},
		{ID: "reader", To: "d@x.com", Status: "SENT", SentAt: sent},
		{ID: "failed", To: "e@x.com", Status: "FAILED"},
		{ID: "pending", To: "f@x.com", Status: "PENDING"},
		{ID: "spam", To: "g@x.com", Status: "COMPLAINED", SentAt: sent},
		{ID: "untracked", To: "h@x.com", Status: "SENT", SentAt: sent, ClickTrackingEnabled: &off},
	}
	ev := func(email string, event models.EmailTrackingEvent, at time.Duration, machine, url string) models.EmailTracking {
		return models.EmailTracking{EmailID: email, Event: event, Timestamp: sent.Add(at), Machine: machine, URL: url}
	}
	rows := recipientOutcomes(emails, []models.EmailTracking{
		ev("delivered-then-bounced", models.EmailTrackingEventBounce, time.Hour, "", ""),
		ev("machines-only", models.EmailTrackingEventOpen, time.Minute, models.TrackingMachineBot, ""),
		ev("machines-only", models.EmailTrackingEventOpen, time.Minute, models.TrackingMachineProxy, ""),
		ev("machines-only", models.EmailTrackingEventClick, time.Minute, models.TrackingMachineBot, "https://x.com/a"),
		ev("reader", models.EmailTrackingEventOpen, 2*time.Hour, "", ""),
		ev("reader", models.EmailTrackingEventOpen, time.Hour, "", ""),
		ev("reader", models.EmailTrackingEventClick, 3*time.Hour, "", "https://x.com/a"),
		ev("reader", models.EmailTrackingEventClick, 4*time.Hour, "", "https://x.com/a"),
		ev("reader", models.EmailTrackingEventClick, 5*time.Hour, models.TrackingMachineBot, "https://x.com/b"),
		ev("reader", models.EmailTrackingEventUnsubscribe, 6*time.Hour, "", ""),
	})
	by := map[string]recipientRow{}
	for _, r := range rows {
		by[r.EmailID] = r
	}

	require.Equal(t, "bounced", by["delivered-then-bounced"].Delivery, "a bounce wins over a delivered status")
	require.Equal(t, "bounced", by["bounced"].Delivery)
	require.Equal(t, "received", by["machines-only"].Delivery)
	require.Nil(t, by["machines-only"].OpenedAt, "bot and proxy opens are not opens")
	require.Nil(t, by["machines-only"].ClickedAt, "scanner clicks are not clicks")
	require.True(t, by["machines-only"].matches("not_opened", ""))

	r := by["reader"]
	require.Equal(t, 2, r.Opens)
	require.Equal(t, sent.Add(time.Hour), *r.OpenedAt, "the first open")
	require.Equal(t, []string{"https://x.com/a"}, r.Links, "each link once, scanner clicks left out")
	require.NotNil(t, r.UnsubscribedAt)
	require.True(t, r.matches("clicked", "https://x.com/a"))
	require.False(t, r.matches("clicked", "https://x.com/b"))
	require.False(t, r.matches("not_opened", ""))

	require.Equal(t, "not_delivered", by["failed"].Delivery)
	require.Equal(t, "waiting", by["pending"].Delivery)
	require.Equal(t, "received", by["spam"].Delivery)
	require.True(t, by["spam"].matches("spam", ""))
	require.False(t, by["untracked"].matches("not_opened", ""), "untracked email can't say whether it was opened")
	require.False(t, by["bounced"].matches("not_opened", ""))
}

func TestHourlyEngagement(t *testing.T) {
	start := time.Date(2026, 11, 1, 10, 20, 0, 0, time.UTC)
	ev := func(email string, event models.EmailTrackingEvent, at time.Time, machine string) models.EmailTracking {
		return models.EmailTracking{EmailID: email, Event: event, Timestamp: at, Machine: machine}
	}
	at := func(h, m int) time.Time { return time.Date(2026, 11, 1, h, m, 0, 0, time.UTC) }
	buckets, later := hourlyEngagement(start, at(13, 20), []models.EmailTracking{
		ev("a", models.EmailTrackingEventOpen, at(10, 50), ""),
		ev("a", models.EmailTrackingEventOpen, at(12, 5), ""), // a repeat open
		ev("b", models.EmailTrackingEventOpen, at(12, 10), ""),
		ev("c", models.EmailTrackingEventOpen, at(11, 0), models.TrackingMachineProxy),
		ev("c", models.EmailTrackingEventClick, at(11, 1), models.TrackingMachineBot),
		ev("a", models.EmailTrackingEventClick, at(11, 30), ""),
		ev("d", models.EmailTrackingEventOpen, at(9, 59), ""), // clock skew lands in the first hour
		ev("e", models.EmailTrackingEventOpen, start.Add(200*time.Hour), ""),
	})
	require.Len(t, buckets, 4, "10:00 through 13:00")
	require.Equal(t, at(10, 0), buckets[0].Hour)
	require.Equal(t, []int{2, 0, 1, 0}, []int{buckets[0].Opens, buckets[1].Opens, buckets[2].Opens, buckets[3].Opens})
	require.Equal(t, []int{0, 1, 0, 0}, []int{buckets[0].Clicks, buckets[1].Clicks, buckets[2].Clicks, buckets[3].Clicks})
	require.Equal(t, 1, later.Opens, "past the end of the timeline")

	buckets, _ = hourlyEngagement(start, start.Add(30*24*time.Hour), nil)
	require.Len(t, buckets, maxTimelineHours)
}

func TestCampaignRecipientsAndFollowUp(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Campaign{}, &models.Email{}, &models.EmailTracking{}, &models.Contact{}, &models.Segment{}, &models.SegmentContact{}))
	quiet := db.Session(&gorm.Session{SkipHooks: true})
	team, list := uuid.NewString(), uuid.NewString()
	camp := models.Campaign{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: list, Name: "Launch", Status: models.CampaignStatusCompleted, Timezone: "UTC", Schedule: models.CampaignScheduleOneTime}
	require.NoError(t, quiet.Create(&camp).Error)
	var contacts []models.Contact
	for i, status := range []models.EmailStatus{"SENT", "SENT", "BOUNCED"} {
		c := models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: list, Email: []string{"ann@x.com", "bob@x.com", "cat@x.com"}[i], Status: models.SubscriberStatusActive}
		require.NoError(t, quiet.Create(&c).Error)
		contacts = append(contacts, c)
		e := models.Email{Base: models.Base{ID: uuid.NewString()}, TeamID: team, CampaignID: camp.ID, ContactID: c.ID, To: c.Email, Status: status, SentAt: time.Now()}
		require.NoError(t, quiet.Create(&e).Error)
		if i == 0 {
			require.NoError(t, quiet.Create(&models.EmailTracking{Base: models.Base{ID: uuid.NewString()}, EmailID: e.ID, Event: models.EmailTrackingEventOpen, Timestamp: time.Now()}).Error)
		}
	}
	h := NewMarketingHandler(db, "secret", "https://api.xem.email")
	call := func(fn echo.HandlerFunc, method, target, body, teamID string) (*httptest.ResponseRecorder, error) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c := echo.New().NewContext(req, rec)
		c.Set("teamID", teamID)
		c.SetParamNames("id")
		c.SetParamValues(camp.ID)
		return rec, fn(c)
	}

	rec, err := call(h.CampaignRecipients, "GET", "/?filter=not_opened", "", team)
	require.NoError(t, err)
	var list1 struct {
		Recipients []recipientRow `json:"recipients"`
		Total      int            `json:"total"`
		Counts     map[string]int `json:"counts"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list1))
	require.Equal(t, 1, list1.Total)
	require.Equal(t, "bob@x.com", list1.Recipients[0].Email)
	require.Equal(t, map[string]int{"all": 3, "received": 2, "opened": 1, "not_opened": 1, "bounced": 1}, list1.Counts)

	rec, err = call(h.CampaignRecipients, "GET", "/?q=ANN", "", team)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list1))
	require.Equal(t, 1, list1.Total, "search by address ignores case")

	_, err = call(h.CampaignRecipients, "GET", "/", "", uuid.NewString())
	require.Error(t, err, "another workspace's campaign is not found")

	rec, err = call(h.CreateFollowUp, "POST", "/", `{"name":"Didn't open Launch","filter":"not_opened"}`, team)
	require.NoError(t, err)
	var draft models.Campaign
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &draft))
	require.Equal(t, models.CampaignStatusDraft, draft.Status)
	require.NotNil(t, draft.SegmentID)
	var segment models.Segment
	require.NoError(t, db.First(&segment, "id = ?", *draft.SegmentID).Error)
	require.False(t, segment.IsDynamic)
	require.Equal(t, 1, segment.ContactCount)

	// The follow-up goes only to the segment.
	old, fresh := marketing.Recipients(db, &draft)
	var ids []string
	require.NoError(t, old.Pluck("id", &ids).Error)
	var more []string
	require.NoError(t, fresh.Pluck("id", &more).Error)
	require.Equal(t, []string{contacts[1].ID}, append(ids, more...))

	_, err = call(h.SaveRecipientSegment, "POST", "/", `{"name":"Clicked","filter":"clicked"}`, team)
	require.Error(t, err, "nobody clicked, so there's no segment to save")
}
