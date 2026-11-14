package notifications

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/analytics"
	"kori/internal/models"
)

func TestBuildSummaryOnPostgres(t *testing.T) {
	dsn := os.Getenv("POSTHOOT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("POSTHOOT_TEST_DATABASE_URL required for the summary's analytics queries")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	schema := "summary_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	require.NoError(t, admin.Exec("CREATE SCHEMA "+schema).Error)
	db, err := gorm.Open(postgres.Open(dsn+" search_path="+schema), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	t.Cleanup(func() { raw, _ := db.DB(); raw.Close(); admin.Exec("DROP SCHEMA " + schema + " CASCADE"); a, _ := admin.DB(); a.Close() })
	require.NoError(t, db.AutoMigrate(&models.Team{}, &models.Contact{}, &models.MailingList{}, &models.Tag{}, &models.Campaign{}, &models.Email{}, &models.EmailTracking{}, &models.SuppressionList{}, &models.EmailBounce{}, &models.ComplaintReport{}))
	require.NoError(t, db.Transaction(analytics.Install))
	put := func(v any) { require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(v).Error) }

	team, list, campaign := uuid.NewString(), uuid.NewString(), uuid.NewString()
	ctx := context.Background()
	p, _ := LastPeriod("weekly", "UTC", time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC))
	quiet, err := BuildSummary(ctx, db, team, p)
	require.NoError(t, err)
	require.True(t, quiet.Quiet(), "nothing sent and no subscriber changes")

	put(&models.MailingList{Base: models.Base{ID: list}, TeamID: team, Name: "Readers"})
	put(&models.Campaign{Base: models.Base{ID: campaign}, TeamID: team, ListID: list, Name: "Launch", TemplateID: uuid.NewString(), SMTPConfigID: uuid.NewString()})
	tracked := true
	send := func(to string, at time.Time, opened bool) {
		id := uuid.NewString()
		put(&models.Email{Base: models.Base{ID: id, CreatedAt: at}, TeamID: team, CampaignID: campaign, To: to, From: "news@example.com", SentAt: at, Status: models.EmailStatusSent, ClickTrackingEnabled: &tracked, SMTPConfigID: uuid.NewString(), CategoryID: uuid.NewString()})
		if opened {
			put(&models.EmailTracking{Base: models.Base{ID: uuid.NewString()}, EmailID: id, Event: "open", Timestamp: at.Add(time.Hour)})
		}
	}
	send("old@example.com", p.Previous().Add(time.Hour), true)
	send("OLD@example.com ", p.From.Add(time.Hour), false)
	send("new@example.com", p.From.Add(2*time.Hour), true)
	send("later@example.com", p.To.Add(time.Hour), true) // after the period

	s, err := BuildSummary(ctx, db, team, p)
	require.NoError(t, err)
	require.False(t, s.Quiet())
	require.EqualValues(t, 2, s.Now.Sent)
	require.EqualValues(t, 1, s.Before.Sent)
	require.EqualValues(t, 1, s.FirstTime, "old@ heard from the workspace before")
	require.Equal(t, 0.5, *s.Now.OpenRate.Value)
	require.Equal(t, "Launch", s.Top.Name)
	require.Equal(t, "Opens fell 50 points to 50.0%; 1 person got their first email from you.", s.Headline())
}
