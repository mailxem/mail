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

func TestRatesCountTrackedSendsOnly(t *testing.T) {
	db := database(t)
	team, list, campaign := uuid.NewString(), uuid.NewString(), uuid.NewString()
	put(t, db, &models.MailingList{Base: base(list), TeamID: team, Name: "Readers"})
	put(t, db, &models.Campaign{Base: base(campaign), TeamID: team, ListID: list, Name: "Edition", TemplateID: uuid.NewString(), SMTPConfigID: uuid.NewString()})
	stamp := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	on, off := true, false
	var opened string
	// Two tracked sends (one opened and clicked), two untracked, one legacy send with no flag (counted as tracked).
	for i, flag := range []*bool{&on, &on, &off, &off, nil} {
		id := uuid.NewString()
		put(t, db, &models.Email{Base: base(id), TeamID: team, CampaignID: campaign, To: uuid.NewString() + "@example.com", From: "news@example.com", SentAt: stamp, Status: models.EmailStatusSent, ClickTrackingEnabled: flag})
		if i == 0 {
			opened = id
		}
	}
	put(t, db, &models.EmailTracking{Base: base(uuid.NewString()), EmailID: opened, Event: models.EmailTrackingEventOpen, Timestamp: stamp.Add(time.Hour)})
	put(t, db, &models.EmailTracking{Base: base(uuid.NewString()), EmailID: opened, Event: models.EmailTrackingEventClick, Timestamp: stamp.Add(time.Hour), URL: "https://example.com"})
	f, err := analytics.Parse(url.Values{"from": {"2026-09-01"}, "to": {"2026-09-10"}}, team, time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	r, err := analytics.Build(context.Background(), db, f)
	require.NoError(t, err)
	require.EqualValues(t, 5, r.Summary.Accepted)
	require.EqualValues(t, 2, r.Summary.Untracked)
	require.InDelta(t, 1.0/3, *r.Summary.OpenRate.Value, .0001)
	require.InDelta(t, 1.0/3, *r.Summary.ClickRate.Value, .0001)
	require.InDelta(t, 1.0/3, *r.Summary.AudienceClickRate.Value, .0001)
	require.InDelta(t, 0, *r.Summary.BounceRate.Value, .0001) // delivery rates still use every send

	f.Kind, f.Sort, f.Direction, f.Page, f.Limit = "campaigns", "clickRate", "DESC", 1, 10
	p, err := analytics.Breakdown(context.Background(), db, f)
	require.NoError(t, err)
	require.Len(t, p.Items, 1)
	require.EqualValues(t, 2, p.Items[0].Untracked)
	require.InDelta(t, 1.0/3, *p.Items[0].ClickRate.Value, .0001)

	e, err := analytics.BuildEmailReport(context.Background(), db, f)
	require.NoError(t, err)
	require.EqualValues(t, 2, e.Summary.Untracked)
	require.InDelta(t, 1.0/3, *e.OpenRate.Value, .0001)
}
