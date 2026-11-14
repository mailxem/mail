package notifications

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kori/internal/analytics"
	"kori/internal/models"
)

func TestSummaryPeriods(t *testing.T) {
	// Monday 16 Nov 2026, 03:00 UTC is 08:30 the same Monday in Kolkata.
	now := time.Date(2026, 11, 16, 3, 0, 0, 0, time.UTC)
	kolkata, _ := time.LoadLocation("Asia/Kolkata")

	p, due := LastPeriod("weekly", "Asia/Kolkata", now)
	require.True(t, due)
	require.Equal(t, time.Date(2026, 11, 9, 0, 0, 0, 0, kolkata), p.From)
	require.Equal(t, time.Date(2026, 11, 16, 0, 0, 0, 0, kolkata), p.To)
	require.Equal(t, time.Date(2026, 11, 2, 0, 0, 0, 0, kolkata), p.Previous())

	p, due = LastPeriod("daily", "Asia/Kolkata", now)
	require.True(t, due)
	require.Equal(t, time.Date(2026, 11, 15, 0, 0, 0, 0, kolkata), p.From)
	require.Equal(t, time.Date(2026, 11, 14, 0, 0, 0, 0, kolkata), p.Previous())

	// Same instant is 03:00 in UTC: too early. No time zone falls back to UTC.
	p, due = LastPeriod("daily", "", now)
	require.False(t, due)
	require.Equal(t, "UTC", p.Timezone)
	require.Equal(t, time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC), p.From)

	// Weekly summaries only go out on Monday, still for the week before.
	p, due = LastPeriod("weekly", "UTC", time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC))
	require.False(t, due)
	require.Equal(t, time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC), p.From)

	_, due = LastPeriod("off", "UTC", now.Add(10*time.Hour))
	require.False(t, due)

	// A day across a daylight saving change keeps local midnights.
	ny, _ := time.LoadLocation("America/New_York")
	p, _ = LastPeriod("daily", "America/New_York", time.Date(2026, 11, 2, 15, 0, 0, 0, time.UTC))
	require.Equal(t, time.Date(2026, 11, 1, 0, 0, 0, 0, ny), p.From)
	require.Equal(t, 25*time.Hour, p.To.Sub(p.From))

	back, err := ParsePeriod(p.Reference())
	require.NoError(t, err)
	require.True(t, back.From.Equal(p.From) && back.To.Equal(p.To))
	require.Equal(t, p.Frequency, back.Frequency)
	_, err = ParsePeriod("monthly|2026-11-01|UTC")
	require.Error(t, err)
}

func rate(n, d int64) analytics.Rate { return analytics.Ratio(n, d) }

func TestSummaryQuiet(t *testing.T) {
	require.True(t, Summary{}.Quiet())
	require.False(t, Summary{Now: Numbers{Sent: 1}}.Quiet())
	require.False(t, Summary{Changes: 2}.Quiet(), "an unsubscribe with no sends is still news")
	// The period before doesn't count: a busy week followed by a quiet one is skipped.
	require.True(t, Summary{Before: Numbers{Sent: 500, Bounced: 4}}.Quiet())
}

func TestSummaryHeadline(t *testing.T) {
	weekly := Period{Frequency: "weekly", Timezone: "UTC"}
	steady := Numbers{Sent: 1000, OpenRate: rate(400, 1000), ClickRate: rate(50, 1000), BounceRate: rate(5, 1000)}
	for name, tc := range map[string]struct {
		s    Summary
		want string
	}{
		"opens fell after new readers": {
			Summary{Period: weekly, Before: steady, Now: Numbers{Sent: 3000, OpenRate: rate(1020, 3000), ClickRate: rate(150, 3000), BounceRate: rate(15, 3000)}, FirstTime: 2000},
			"Opens fell 6 points to 34.0%; 2,000 people got their first email from you.",
		},
		"opens fell with more volume but no new readers": {
			Summary{Period: weekly, Before: steady, Now: Numbers{Sent: 2000, OpenRate: rate(680, 2000), ClickRate: rate(100, 2000), BounceRate: rate(10, 2000)}, FirstTime: 10},
			"Opens fell 6 points to 34.0%; you sent 100% more email than the week before.",
		},
		"opens fell with no supported reason": {
			Summary{Period: weekly, Before: steady, Now: Numbers{Sent: 1000, OpenRate: rate(340, 1000), ClickRate: rate(50, 1000), BounceRate: rate(5, 1000)}},
			"Opens fell 6 points to 34.0%.",
		},
		"clicks rose without a reason": {
			Summary{Period: weekly, Before: steady, Now: Numbers{Sent: 1000, OpenRate: rate(400, 1000), ClickRate: rate(70, 1000), BounceRate: rate(5, 1000)}},
			"Clicks rose 2 points to 7.0%.",
		},
		"complaints come first": {
			Summary{Period: weekly, Before: steady, Now: Numbers{Sent: 1000, Complaints: 4, OpenRate: rate(300, 1000), ClickRate: rate(50, 1000), BounceRate: rate(5, 1000)}},
			"Spam complaints rose to 4 last week, from 0 the week before.",
		},
		"bounces before engagement": {
			Summary{Period: weekly, Before: steady, Now: Numbers{Sent: 1000, OpenRate: rate(300, 1000), ClickRate: rate(50, 1000), BounceRate: rate(35, 1000)}, FirstTime: 600},
			"Bounces rose 3 points to 3.5% of sends; 600 people got their first email from you.",
		},
		"volume up from one campaign": {
			Summary{Period: weekly, Before: steady, Now: Numbers{Sent: 1500, OpenRate: rate(600, 1500), ClickRate: rate(75, 1500), BounceRate: rate(7, 1500)}, Top: &analytics.Row{Name: "Spring sale", Accepted: 900, Reached: 900}},
			"You sent 50% more email than the week before; Spring sale went to 900 people.",
		},
		"volume down": {
			Summary{Period: weekly, Before: steady, Now: Numbers{Sent: 500, OpenRate: rate(200, 500), ClickRate: rate(25, 500), BounceRate: rate(2, 500)}},
			"You sent 50% less email than the week before.",
		},
		"steady": {
			Summary{Period: weekly, Before: steady, Now: steady},
			"You sent 1,000 emails last week, about as many as the week before.",
		},
		"first sends": {
			Summary{Period: Period{Frequency: "daily"}, Now: Numbers{Sent: 1, OpenRate: rate(1, 1)}},
			"You sent 1 email yesterday, after none the day before.",
		},
		"no sends, new contacts": {
			Summary{Period: weekly, Before: steady, Now: Numbers{NewContacts: 12}, Changes: 12},
			"No emails went out last week, and 12 new contacts joined.",
		},
	} {
		t.Run(name, func(t *testing.T) { require.Equal(t, tc.want, tc.s.Headline()) })
	}
}

func TestSummaryEmailRender(t *testing.T) {
	s := Summary{Period: Period{Frequency: "weekly", Timezone: "UTC", From: time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC), To: time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC)},
		Before: Numbers{Sent: 1000, OpenRate: rate(400, 1000)}, Now: Numbers{Sent: 1100, OpenRate: rate(340, 1100)},
		Top: &analytics.Row{Name: "<b>Launch</b>", Accepted: 800, ClickRate: rate(40, 800)}}
	subject, html, text := s.Email("Studio\r\nBcc: x").Render("https://app.example.net/dashboard")
	require.Equal(t, "Your weekly summary for Studio Bcc: x", subject)
	require.Contains(t, html, "&lt;b&gt;Launch&lt;/b&gt;: 800 sent, 5.0% clicked")
	require.NotContains(t, html, "<b>Launch")
	require.Contains(t, text, "Emails sent: 1,100 (up 10% from 1,000)")
	require.Contains(t, text, "Opened: 30.9% (down 9.1 points from 40.0%)")
	require.Contains(t, text, "Period: Mon 9 Nov – Sun 15 Nov 2026 (UTC)")
	require.True(t, strings.Index(text, "Opens fell 9 points") < strings.Index(text, "Emails sent"), "the headline opens the email")
}

func TestSummariesQueueOncePerPeriodAndSkipWhenTurnedOff(t *testing.T) {
	w, owner := fixture(t)
	require.NoError(t, w.DB.AutoMigrate(&models.TeamSendTimeDefaults{}))
	require.NoError(t, w.DB.Model(&models.Team{}).Where("id = ?", owner.TeamID).Update("summary_frequency", "daily").Error)
	admin := models.User{Base: models.Base{ID: uuid.NewString()}, TeamID: owner.TeamID, Role: models.UserRoleAdmin, Email: "admin@example.net", SummaryEmails: true}
	member := models.User{Base: models.Base{ID: uuid.NewString()}, TeamID: owner.TeamID, Role: models.UserRoleMember, Email: "member@example.net", SummaryEmails: true}
	require.NoError(t, w.DB.Session(&gorm.Session{SkipHooks: true}).Create([]models.User{admin, member}).Error)

	ctx := context.Background()
	require.NoError(t, QueueSummaries(ctx, w.DB, w.Now()))
	require.NoError(t, QueueSummaries(ctx, w.DB, w.Now().Add(15*time.Minute)))
	var notices []models.ServiceNotice
	require.NoError(t, w.DB.Where("kind = ?", models.SummaryNoticeKind).Order("account_email").Find(&notices).Error)
	require.Len(t, notices, 2, "owner and opted-in admin, once each; members never")
	require.Equal(t, "admin@example.net", notices[0].AccountEmail)
	require.Equal(t, "daily|2026-09-29|UTC", notices[1].ReferenceID)

	// Turned off before it went out: skipped without sending.
	require.NoError(t, w.DB.Model(&models.Team{}).Where("id = ?", owner.TeamID).Update("summary_frequency", "off").Error)
	w.Send = func(context.Context, string, string, string, string, string) (bool, error) {
		t.Fatal("summary was sent")
		return false, nil
	}
	require.NoError(t, w.ProcessOne(ctx))
	require.NoError(t, w.ProcessOne(ctx))
	var statuses []string
	require.NoError(t, w.DB.Model(&models.ServiceNotice{}).Where("kind = ?", models.SummaryNoticeKind).Pluck("status", &statuses).Error)
	require.Equal(t, []string{"SKIPPED", "SKIPPED"}, statuses)
}
