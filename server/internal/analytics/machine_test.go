package analytics_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"kori/internal/analytics"
	"kori/internal/models"
)

func TestReportsLeaveMachinesOut(t *testing.T) {
	db := database(t)
	team, list, campaign := uuid.NewString(), uuid.NewString(), uuid.NewString()
	put(t, db, &models.MailingList{Base: base(list), TeamID: team, Name: "Community"})
	put(t, db, &models.Campaign{Base: base(campaign), TeamID: team, ListID: list, Name: "Edition", TemplateID: uuid.NewString(), SMTPConfigID: uuid.NewString()})
	stamp := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	enabled := true
	event := func(emailID string, kind models.EmailTrackingEvent, machine, link string) {
		put(t, db, &models.EmailTracking{Base: base(uuid.NewString()), EmailID: emailID, Event: kind, Timestamp: stamp.Add(time.Hour), URL: link, Machine: machine})
	}
	ada, ben := uuid.NewString(), uuid.NewString()
	for id, to := range map[string]string{ada: "ada@example.com", ben: "ben@example.com"} {
		put(t, db, &models.Contact{Base: base(uuid.NewString()), TeamID: team, ListID: list, Email: to, Status: models.SubscriberStatusActive})
		put(t, db, &models.Email{Base: base(id), TeamID: team, CampaignID: campaign, To: to, From: "news@example.com", SentAt: stamp, Status: models.EmailStatusSent, ClickTrackingEnabled: &enabled})
	}
	// Ada clicked for real; a gateway also followed both links, and her mail app fetched the pixel through Apple's proxy.
	event(ada, models.EmailTrackingEventClick, "", "https://example.com/story")
	event(ada, models.EmailTrackingEventClick, models.TrackingMachineBot, "https://example.com/story")
	event(ada, models.EmailTrackingEventClick, models.TrackingMachineBot, "https://example.com/scan-only")
	event(ada, models.EmailTrackingEventOpen, models.TrackingMachineProxy, "")
	// Ben only ever met a scanner.
	event(ben, models.EmailTrackingEventClick, models.TrackingMachineBot, "https://example.com/story")
	event(ben, models.EmailTrackingEventOpen, models.TrackingMachineBot, "")

	f, e := analytics.Parse(url.Values{"from": {"2026-09-01"}, "to": {"2026-09-10"}}, team, time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC))
	require.NoError(t, e)
	r, e := analytics.Build(context.Background(), db, f)
	require.NoError(t, e)
	require.EqualValues(t, 2, r.Summary.Accepted)
	require.EqualValues(t, 1, r.Summary.ClickedMessages)
	require.EqualValues(t, 1, r.Summary.Clicked)
	require.Zero(t, r.Summary.Opened)
	require.EqualValues(t, 3, r.Summary.BotClicks)
	require.EqualValues(t, 1, r.Summary.BotOpens)
	require.EqualValues(t, 1, r.Summary.ProxyOpens)
	require.EqualValues(t, 1, r.Charts.ClickEvents)
	require.EqualValues(t, 1, r.Activity[4].Clicked)
	for _, c := range r.Cohorts {
		if c.Key == "recent" {
			require.EqualValues(t, 1, c.Count, "a scanner click does not make Ben a recent clicker")
		}
	}

	f.Kind = "links"
	p, e := analytics.Breakdown(context.Background(), db, f)
	require.NoError(t, e)
	require.Len(t, p.Items, 1)
	require.Equal(t, "https://example.com/story", p.Items[0].Name)
	f.Kind = "campaigns"
	p, e = analytics.Breakdown(context.Background(), db, f)
	require.NoError(t, e)
	require.EqualValues(t, 1, p.Items[0].ClickedMessages)
	require.EqualValues(t, 3, p.Items[0].BotClicks)

	f.Kind = ""
	er, e := analytics.BuildEmailReport(context.Background(), db, f)
	require.NoError(t, e)
	require.EqualValues(t, 1, er.Summary.Clicked)
	require.Zero(t, er.Summary.Opened)
	require.EqualValues(t, 3, er.Summary.BotClicks)
	require.EqualValues(t, 1, er.Summary.BotOpens)
	require.EqualValues(t, 1, er.Summary.ProxyOpens)
}
