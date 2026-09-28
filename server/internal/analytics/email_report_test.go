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

func TestEmailOverviewWithoutCampaigns(t *testing.T) {
	db := database(t)
	team, foreign := uuid.NewString(), uuid.NewString()
	stamp := time.Date(2026, 9, 5, 23, 30, 0, 0, time.UTC)
	add := func(status models.EmailStatus, sent time.Time) string {
		id := uuid.NewString()
		b := base(id)
		b.CreatedAt = stamp
		put(t, db, &models.Email{Base: b, TeamID: team, From: "sender@example.com", To: "reader@example.com", Status: status, SentAt: sent})
		return id
	}
	sent := add(models.EmailStatusSent, stamp)
	bounced := add(models.EmailStatusBounced, stamp)
	add("QUEUED", time.Time{})
	add(models.EmailStatusPending, time.Time{})
	add(models.EmailStatusFailed, time.Time{})
	add("DRAFT", time.Time{})
	add(models.EmailStatusSent, time.Time{}) // Legacy sent status without timestamp is unknown.
	foreignID := add(models.EmailStatusSent, stamp)
	require.NoError(t, db.Table("emails").Where("id=?", foreignID).Update("team_id", foreign).Error)
	testID := add(models.EmailStatusSent, stamp)
	require.NoError(t, db.Table("emails").Where("id=?", testID).Update("test", true).Error)
	deleted := add(models.EmailStatusSent, stamp)
	require.NoError(t, db.Table("emails").Where("id=?", deleted).Update("is_deleted", true).Error)
	old := add(models.EmailStatusSent, stamp.AddDate(0, 0, -20))
	require.NoError(t, db.Table("emails").Where("id=?", old).Update("created_at", stamp.AddDate(0, 0, -21)).Error)
	// This send belongs in the period even though its record was created earlier.
	require.NoError(t, db.Table("emails").Where("id=?", sent).Update("created_at", stamp.AddDate(0, 0, -20)).Error)
	track := func(id string, event models.EmailTrackingEvent, at time.Time) {
		put(t, db, &models.EmailTracking{Base: base(uuid.NewString()), EmailID: id, Event: event, Timestamp: at})
	}
	track(sent, "open", stamp.Add(time.Hour))
	track(sent, "open", stamp.Add(2*time.Hour))
	track(sent, "click", stamp.Add(time.Hour))
	track(sent, "click", stamp.Add(2*time.Hour))
	track(bounced, "click", stamp.Add(-time.Hour))
	track(bounced, "open", stamp.AddDate(0, 0, 20))
	track(foreignID, "click", stamp.Add(time.Hour))
	f, err := analytics.Parse(url.Values{"from": {"2026-09-01"}, "to": {"2026-09-10"}, "timezone": {"Asia/Kolkata"}}, team, time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	r, err := analytics.BuildEmailReport(context.Background(), db, f)
	require.NoError(t, err)
	require.EqualValues(t, 7, r.Summary.Total)
	require.EqualValues(t, 2, r.Summary.Sent)
	require.EqualValues(t, 2, r.Summary.Queued)
	require.EqualValues(t, 1, r.Summary.Failed)
	require.EqualValues(t, 1, r.Summary.Drafts)
	require.EqualValues(t, 1, r.Summary.Unknown)
	require.EqualValues(t, 1, r.Summary.Opened)
	require.EqualValues(t, 1, r.Summary.Clicked)
	require.EqualValues(t, 1, r.Summary.Bounced)
	require.EqualValues(t, 0, r.Summary.Campaigns)
	require.EqualValues(t, 2, r.Summary.Other)
	require.Equal(t, 0.5, *r.ClickRate.Value)
	require.Len(t, r.Series, 9)
	require.Equal(t, "2026-09-06", r.Series[5].Date)
	require.EqualValues(t, 7, r.Series[5].Total)
	require.Equal(t, r.Summary.Total, r.Summary.Sent+r.Summary.Queued+r.Summary.Failed+r.Summary.Drafts+r.Summary.Unknown)
	// Campaign-specific analytics remains empty while generic email activity exists.
	campaigns, err := analytics.Build(context.Background(), db, f)
	require.NoError(t, err)
	require.Zero(t, campaigns.Summary.Accepted)
	// Linking an existing email to a campaign cannot inflate the overall sent count.
	campaign := uuid.NewString()
	put(t, db, &models.Campaign{Base: base(campaign), TeamID: team, ListID: uuid.NewString(), Name: "Launch", TemplateID: uuid.NewString(), SMTPConfigID: uuid.NewString()})
	require.NoError(t, db.Table("emails").Where("id=?", sent).Update("campaign_id", campaign).Error)
	r, err = analytics.BuildEmailReport(context.Background(), db, f)
	require.NoError(t, err)
	require.EqualValues(t, 2, r.Summary.Sent)
	require.EqualValues(t, 1, r.Summary.Campaigns)
	require.EqualValues(t, 1, r.Summary.Other)
	f.Team = uuid.NewString()
	r, err = analytics.BuildEmailReport(context.Background(), db, f)
	require.NoError(t, err)
	require.Zero(t, r.Summary.Total)
	require.Nil(t, r.ClickRate.Value)
	require.Len(t, r.Series, 9)
	f.List = uuid.NewString()
	_, err = analytics.BuildEmailReport(context.Background(), db, f)
	require.Error(t, err)
}
