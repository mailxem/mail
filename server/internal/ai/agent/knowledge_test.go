package agent

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/models"
)

// Seed the current email/event model; tracking rows are events, not deliveries.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Team{}, &models.MailingList{}, &models.Contact{}, &models.Campaign{}, &models.Email{}, &models.EmailTracking{}))
	raw, err := db.DB()
	require.NoError(t, err)
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() { require.NoError(t, raw.Close()) })
	return db
}

type analyticsFixture struct {
	db       *gorm.DB
	team     string
	contact  models.Contact
	campaign models.Campaign
}

func newAnalyticsFixture(t *testing.T) analyticsFixture {
	t.Helper()
	db := setupTestDB(t)
	team := uuid.NewString()
	list := models.MailingList{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Readers"}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&list).Error)
	contact := models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: list.ID, Email: "reader@example.com", FirstName: "Test", LastName: "Reader"}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&contact).Error)
	campaign := models.Campaign{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: list.ID, Name: "Test Campaign", Subject: "Test Subject", Status: models.CampaignStatusCompleted}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&campaign).Error)
	return analyticsFixture{db, team, contact, campaign}
}
func (f analyticsFixture) message(t *testing.T, device string, events ...models.EmailTrackingEvent) models.Email {
	t.Helper()
	now := time.Now().UTC()
	email := models.Email{Base: models.Base{ID: uuid.NewString(), CreatedAt: now}, TeamID: f.team, ContactID: f.contact.ID, CampaignID: f.campaign.ID, From: "sender@example.com", To: f.contact.Email, Subject: "Hello", Body: "Hello", Status: models.EmailStatusSent}
	require.NoError(t, f.db.Session(&gorm.Session{SkipHooks: true}).Create(&email).Error)
	for _, event := range events {
		row := models.EmailTracking{Base: models.Base{ID: uuid.NewString(), CreatedAt: now}, EmailID: email.ID, ContactID: f.contact.ID, CampaignID: f.campaign.ID, Timestamp: now, Event: event, DeviceType: device}
		require.NoError(t, f.db.Session(&gorm.Session{SkipHooks: true}).Create(&row).Error)
	}
	return email
}

func TestAnalyticsAggregator_GetEmailAnalytics(t *testing.T) {
	f := newAnalyticsFixture(t)
	f.message(t, "mobile", models.EmailTrackingEventOpen, models.EmailTrackingEventOpen, models.EmailTrackingEventClick)
	f.message(t, "desktop", models.EmailTrackingEventOpen)
	f.message(t, "desktop")
	f.message(t, "desktop", models.EmailTrackingEventBounce)
	// An unrelated workspace must not change any count or rate.
	other := f
	other.team = uuid.NewString()
	other.message(t, "mobile", models.EmailTrackingEventOpen, models.EmailTrackingEventClick)
	got, err := NewAnalyticsAggregator(f.db).GetEmailAnalytics(context.Background(), AnalyticsScope{TeamID: f.team, StartDate: time.Now().AddDate(0, 0, -30)})
	require.NoError(t, err)
	require.Equal(t, int64(4), got.TotalSent)
	require.Equal(t, int64(3), got.TotalOpened)
	require.Equal(t, int64(2), got.UniqueOpens)
	require.Equal(t, int64(1), got.UniqueClicks)
	require.Equal(t, 50.0, got.OpenRate)
	require.Equal(t, 25.0, got.ClickRate)
	require.Equal(t, 25.0, got.BounceRate)
}
func TestAnalyticsAggregator_GetCampaignAnalytics(t *testing.T) {
	f := newAnalyticsFixture(t)
	f.message(t, "mobile", models.EmailTrackingEventOpen)
	f.message(t, "desktop", models.EmailTrackingEventOpen)
	f.message(t, "mobile", models.EmailTrackingEventOpen)
	got, err := NewAnalyticsAggregator(f.db).GetCampaignAnalytics(context.Background(), f.campaign.ID)
	require.NoError(t, err)
	require.Equal(t, int64(3), got.EmailsSent)
	require.Equal(t, 100.0, got.OpenRate)
	require.Equal(t, "mobile", got.BestDevice)
	require.Equal(t, f.campaign.Name, got.CampaignName)
}
func TestAnalyticsAggregator_GetContactInsights(t *testing.T) {
	f := newAnalyticsFixture(t)
	f.message(t, "mobile", models.EmailTrackingEventOpen, models.EmailTrackingEventOpen, models.EmailTrackingEventClick)
	f.message(t, "desktop", models.EmailTrackingEventOpen)
	f.message(t, "desktop")
	got, err := NewAnalyticsAggregator(f.db).GetContactInsights(context.Background(), f.contact.ID)
	require.NoError(t, err)
	require.Equal(t, f.contact.ID, got.ContactID)
	require.Equal(t, int64(3), got.TotalEmailsReceived)
	require.Equal(t, int64(2), got.EmailsOpened)
	require.Equal(t, int64(1), got.EmailsClicked)
	require.InDelta(t, 53.333, got.EngagementScore, 0.01)
	require.NotNil(t, got.LastOpenedAt)
	require.NotNil(t, got.LastClickedAt)
}
func TestAnalyticsAggregator_EngagementScore(t *testing.T) {
	for _, tt := range []struct {
		received, opens, clicks int
		want                    float64
	}{{10, 8, 5, 68}, {10, 4, 3, 36}, {10, 1, 0, 6}, {0, 0, 0, 0}} {
		t.Run(fmt.Sprintf("%d-%d-%d", tt.received, tt.opens, tt.clicks), func(t *testing.T) {
			f := newAnalyticsFixture(t)
			for i := 0; i < tt.received; i++ {
				events := []models.EmailTrackingEvent{}
				if i < tt.opens {
					events = append(events, models.EmailTrackingEventOpen)
				}
				if i < tt.clicks {
					events = append(events, models.EmailTrackingEventClick)
				}
				f.message(t, "mobile", events...)
			}
			got, err := NewAnalyticsAggregator(f.db).GetContactInsights(context.Background(), f.contact.ID)
			require.NoError(t, err)
			require.InDelta(t, tt.want, got.EngagementScore, 0.01)
		})
	}
}
func TestKnowledgeBase_BuildTeamContext(t *testing.T) {
	f := newAnalyticsFixture(t)
	f.message(t, "mobile", models.EmailTrackingEventOpen, models.EmailTrackingEventClick)
	got, err := NewKnowledgeBase(f.db).BuildTeamContext(context.Background(), f.team, 30)
	require.NoError(t, err)
	for _, text := range []string{"Team Performance (Last 30 days)", "Total Emails Sent: 1", "Average Open Rate: 100.00%", "Average Click Rate: 100.00%"} {
		require.Contains(t, got, text)
	}
}
func TestKnowledgeBase_BuildCampaignContext(t *testing.T) {
	f := newAnalyticsFixture(t)
	f.message(t, "mobile", models.EmailTrackingEventOpen)
	got, err := NewKnowledgeBase(f.db).BuildCampaignContext(context.Background(), f.campaign.ID)
	require.NoError(t, err)
	for _, text := range []string{f.campaign.Name, "Emails Sent: 1", "Open Rate: 100.00%", "Most Used Device: mobile", "Performance:"} {
		require.Contains(t, got, text)
	}
}
func TestKnowledgeBase_BuildContactContext(t *testing.T) {
	f := newAnalyticsFixture(t)
	f.message(t, "mobile", models.EmailTrackingEventOpen, models.EmailTrackingEventClick)
	got, err := NewKnowledgeBase(f.db).BuildContactContext(context.Background(), f.contact.ID)
	require.NoError(t, err)
	for _, text := range []string{f.contact.Email, "Engagement Score: 100.0/100", "Emails Opened: 1 (100.0%)", "Emails Clicked: 1 (100.0%)"} {
		require.Contains(t, got, text)
	}
}
func TestKnowledgeBase_EmptyContactContext(t *testing.T) {
	f := newAnalyticsFixture(t)
	got, err := NewKnowledgeBase(f.db).BuildContactContext(context.Background(), f.contact.ID)
	require.NoError(t, err)
	require.Contains(t, got, "Emails Opened: 0 (0.0%)")
	require.NotContains(t, got, "NaN")
}
func TestKnowledgeBase_ClassifyEngagement(t *testing.T) {
	for _, tt := range []struct {
		score float64
		want  string
	}{{80, "Highly Engaged"}, {75, "Highly Engaged"}, {65, "Moderately Engaged"}, {50, "Moderately Engaged"}, {30, "Lightly Engaged"}, {25, "Lightly Engaged"}, {20, "Disengaged"}, {0, "Disengaged"}} {
		require.Equal(t, tt.want, classifyEngagement(tt.score))
	}
}
func TestKnowledgeBase_BenchmarkPerformance(t *testing.T) {
	for _, tt := range []struct {
		open, click         float64
		openText, clickText string
	}{{25, 3, "above", "above"}, {20, 2.5, "below", "below"}, {15, 4, "below", "above"}} {
		got := benchmarkPerformance(tt.open, tt.click)
		require.Contains(t, got, "Open rate is "+tt.openText+" industry average")
		require.Contains(t, got, "Click rate is "+tt.clickText+" industry average")
	}
}
