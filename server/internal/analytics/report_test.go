package analytics_test

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/analytics"
	"kori/internal/api/middleware"
	"kori/internal/handlers"
	"kori/internal/models"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func database(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("POSTHOOT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("POSTHOOT_TEST_DATABASE_URL required for PostgreSQL analytics integration tests")
	}
	admin, e := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, e)
	schema := "analytics_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	require.NoError(t, admin.Exec("CREATE SCHEMA "+schema).Error)
	db, e := gorm.Open(postgres.Open(dsn+" search_path="+schema), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, e)
	t.Cleanup(func() {
		raw, _ := db.DB()
		raw.Close()
		admin.Exec("DROP SCHEMA " + schema + " CASCADE")
		a, _ := admin.DB()
		a.Close()
	})
	require.NoError(t, db.AutoMigrate(&models.Team{}, &models.Contact{}, &models.MailingList{}, &models.Tag{}, &models.Campaign{}, &models.Email{}, &models.EmailTracking{}, &models.SuppressionList{}, &models.Resource{}, &models.ResourcePermission{}, &models.UserPermission{}, &models.FormSubmission{}, &models.EmailBounce{}, &models.ComplaintReport{}))
	require.NoError(t, db.Transaction(analytics.Install))
	return db
}
func put(t *testing.T, db *gorm.DB, v interface{}) {
	if email, ok := v.(*models.Email); ok {
		email.SMTPConfigID = uuid.NewString()
		email.CategoryID = uuid.NewString()
	}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(v).Error)
}
func base(id string) models.Base {
	return models.Base{ID: id, CreatedAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)}
}
func TestReportDeduplicationFiltersAndPagination(t *testing.T) {
	db := database(t)
	team, foreign, list, otherList, campaign := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	put(t, db, &models.MailingList{Base: base(list), TeamID: team, Name: "Community"})
	put(t, db, &models.MailingList{Base: base(otherList), TeamID: team, Name: "Journal"})
	a, b, dup := uuid.NewString(), uuid.NewString(), uuid.NewString()
	for _, c := range []models.Contact{{Base: base(a), TeamID: team, ListID: list, Email: "ada@example.com", Status: models.SubscriberStatusActive}, {Base: base(b), TeamID: team, ListID: list, Email: "ben@example.com", Status: models.SubscriberStatusActive}, {Base: base(dup), TeamID: team, ListID: otherList, Email: "ADA@example.com", Status: models.SubscriberStatusActive}} {
		put(t, db, &c)
	}
	put(t, db, &models.Campaign{Base: base(campaign), TeamID: team, ListID: list, Name: "Edition", TemplateID: uuid.NewString(), SMTPConfigID: uuid.NewString()})
	stamp := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	enabled := true
	for i, to := range []string{"ada@example.com", "ada@example.com", "ben@example.com"} {
		id := uuid.NewString()
		put(t, db, &models.Email{Base: base(id), TeamID: team, CampaignID: campaign, To: to, From: "news@example.com", SentAt: stamp, Status: models.EmailStatusSent, ClickTrackingEnabled: &enabled})
		if i < 2 {
			for j := 0; j < 3; j++ {
				put(t, db, &models.EmailTracking{Base: base(uuid.NewString()), EmailID: id, Event: models.EmailTrackingEventClick, Timestamp: stamp.Add(time.Hour), URL: "https://example.com/story"})
			}
		}
	}
	for _, s := range []models.EmailStatus{models.EmailStatusFailed, models.EmailStatusPending} {
		put(t, db, &models.Email{Base: base(uuid.NewString()), TeamID: team, CampaignID: campaign, To: "ignored@example.com", From: "news@example.com", SentAt: stamp, Status: s})
	}
	put(t, db, &models.Email{Base: base(uuid.NewString()), TeamID: team, CampaignID: campaign, To: "test@example.com", From: "news@example.com", SentAt: stamp, Status: models.EmailStatusSent, Test: true})
	// A corrupt foreign email pointing at this campaign must not enter aggregates.
	put(t, db, &models.Email{Base: base(uuid.NewString()), TeamID: foreign, CampaignID: campaign, To: "foreign@example.com", From: "news@example.com", SentAt: stamp, Status: models.EmailStatusSent})
	f, e := analytics.Parse(url.Values{"from": {"2026-09-01"}, "to": {"2026-09-10"}}, team, time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC))
	require.NoError(t, e)
	r, e := analytics.Build(context.Background(), db, f)
	require.NoError(t, e)
	require.EqualValues(t, 2, r.Summary.Active)
	require.EqualValues(t, 3, r.Summary.Accepted)
	require.EqualValues(t, 2, r.Summary.Reached)
	require.EqualValues(t, 1, r.Summary.Clicked)
	require.InDelta(t, 2.0/3, *r.Summary.ClickRate.Value, .0001)
	require.InDelta(t, .5, *r.Summary.AudienceClickRate.Value, .0001)
	require.Nil(t, r.Previous.ClickRate.Value)
	require.Len(t, r.Activity, 9)
	require.EqualValues(t, 1, r.Activity[4].Clicked)
	var total int64
	for _, c := range r.Cohorts {
		total += c.Count
	}
	require.Equal(t, r.Summary.Active, total)
	f.Kind = "lists"
	f.Limit = 1
	f.Page = 2
	p, e := analytics.Breakdown(context.Background(), db, f)
	require.NoError(t, e)
	require.EqualValues(t, 2, p.Total)
	require.Len(t, p.Items, 1)
	f.Kind = "links"
	f.Page = 1
	p, e = analytics.Breakdown(context.Background(), db, f)
	require.NoError(t, e)
	require.EqualValues(t, 2, p.Items[0].ClickedMessages)
	f.Kind = "sources"
	_, e = analytics.Breakdown(context.Background(), db, f)
	require.NoError(t, e)
	f.Kind = "campaigns"
	p, e = analytics.Breakdown(context.Background(), db, f)
	require.NoError(t, e)
	require.Len(t, p.Items, 1)
	f.Kind = "domains"
	_, e = analytics.Breakdown(context.Background(), db, f)
	require.NoError(t, e)
	f.Cohort = "recent"
	people, e := analytics.People(context.Background(), db, f)
	require.NoError(t, e)
	require.EqualValues(t, 1, people.Total)
	f.List = otherList
	r, e = analytics.Build(context.Background(), db, f)
	require.NoError(t, e)
	require.EqualValues(t, 1, r.Summary.Active)
	require.Zero(t, r.Summary.Accepted)
	// Public middleware rejects cross-team queries and foreign resource identifiers.
	h := handlers.NewTrackingHandler(db)
	echoServer := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/analytics/report?teamId="+foreign, nil)
	c := echoServer.NewContext(req, rec)
	c.Set("teamID", team)
	c.Set("hasAdminAccess", true)
	err := middleware.AnalyticsAccess(db)(h.AudienceReport)(c)
	require.Equal(t, 403, err.(*echo.HTTPError).Code)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/analytics/report?campaignId="+campaign, nil)
	c = echoServer.NewContext(req, rec)
	c.Set("teamID", foreign)
	c.Set("hasAdminAccess", true)
	err = middleware.AnalyticsAccess(db)(h.AudienceReport)(c)
	require.Equal(t, 404, err.(*echo.HTTPError).Code)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v1/analytics/report", nil)
	c = echoServer.NewContext(req, rec)
	c.Set("teamID", team)
	c.Set("userID", uuid.NewString())
	c.Set("scopes", []string{"read"})
	err = middleware.AnalyticsAccess(db)(h.AudienceReport)(c)
	require.Equal(t, http.StatusForbidden, err.(*echo.HTTPError).Code)
	if path := os.Getenv("ANALYTICS_FIXTURE_PATH"); path != "" {
		f.List = ""
		r, e = analytics.Build(context.Background(), db, f)
		require.NoError(t, e)
		data, _ := json.Marshal(r)
		require.NoError(t, os.WriteFile(path, data, 0600))
	}
}
func TestSubscriptionHistoryTracksBulkChangesAndRollbacks(t *testing.T) {
	db := database(t)
	team, list, id := uuid.NewString(), uuid.NewString(), uuid.NewString()
	put(t, db, &models.MailingList{Base: base(list), TeamID: team, Name: "History"})
	put(t, db, &models.Contact{Base: base(id), TeamID: team, ListID: list, Email: "history@example.com", Status: models.SubscriberStatusActive})
	require.NoError(t, db.Model(&models.Contact{}).Where("id=?", id).UpdateColumn("status", "UNSUBSCRIBED").Error)
	require.NoError(t, db.Model(&models.Contact{}).Where("id=?", id).UpdateColumn("status", "UNSUBSCRIBED").Error)
	var count int64
	require.NoError(t, db.Table("subscription_events").Count(&count).Error)
	require.EqualValues(t, 2, count)
	tx := db.Begin()
	require.NoError(t, tx.Model(&models.Contact{}).Where("id=?", id).UpdateColumn("status", "ACTIVE").Error)
	tx.Rollback()
	require.NoError(t, db.Table("subscription_events").Count(&count).Error)
	require.EqualValues(t, 2, count)
	now := time.Now().UTC()
	require.NoError(t, db.Exec("UPDATE analytics_coverage SET started_at=?", now.Add(-time.Hour)).Error)
	require.NoError(t, db.Exec("UPDATE subscription_events SET occurred_at=? WHERE kind='insert'", now.Add(-45*time.Minute)).Error)
	f, e := analytics.Parse(url.Values{"from": {now.Add(-30 * time.Minute).Format(time.RFC3339)}, "to": {now.Add(time.Minute).Format(time.RFC3339)}}, team, now)
	require.NoError(t, e)
	r, e := analytics.Build(context.Background(), db, f)
	require.NoError(t, e)
	require.True(t, r.Growth.Available)
	require.EqualValues(t, 1, *r.Growth.Start)
	require.EqualValues(t, 0, *r.Growth.End)
	require.EqualValues(t, -1, *r.Growth.Net)
	require.NoError(t, db.Transaction(analytics.Install))
	require.NoError(t, db.Table("subscription_events").Count(&count).Error)
	require.EqualValues(t, 2, count)
}
func TestFilterValidationAndNullRates(t *testing.T) {
	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	for _, q := range []url.Values{{"timezone": {"invalid"}}, {"page": {"-1"}}, {"limit": {"1000"}}, {"from": {"2025-01-01"}, "to": {"2026-03-10"}}, {"from": {"bad"}}, {"cohort": {"madeup"}}, {"asOf": {"2027-01-01"}}} {
		_, err := analytics.Parse(q, "team", now)
		require.Error(t, err)
	}
	f, e := analytics.Parse(url.Values{"timezone": {"America/New_York"}, "from": {"2026-03-07"}, "to": {"2026-03-10"}}, "team", now)
	require.NoError(t, e)
	require.Equal(t, 71*time.Hour, f.To.Sub(f.From))
	require.Nil(t, analytics.Ratio(0, 0).Value)
}

func TestReportScale(t *testing.T) {
	if os.Getenv("ANALYTICS_SCALE_TEST") != "1" {
		t.Skip("Opt-in representative dataset check")
	}
	db := database(t)
	team := uuid.NewString()
	require.NoError(t, db.Exec(`INSERT INTO mailing_lists(id,team_id,name,created_at,is_deleted) SELECT md5('list'||i)::uuid,?,'Community '||i,now(),false FROM generate_series(1,15) i`, team).Error)
	require.NoError(t, db.Exec(`INSERT INTO contacts(id,team_id,list_id,email,status,created_at,is_deleted) SELECT md5('person'||i)::uuid,?,md5('list'||(1+i%15))::uuid,'reader'||i||'@example.com','ACTIVE','2026-09-02',false FROM generate_series(1,10000) i`, team).Error)
	require.NoError(t, db.Exec(`INSERT INTO campaigns(id,team_id,list_id,name,template_id,smtp_config_id,created_at,is_deleted) SELECT md5('edition'||i)::uuid,?,md5('list'||i)::uuid,'September dispatch '||i,md5('template')::uuid,md5('smtp')::uuid,'2026-09-02',false FROM generate_series(1,15) i`, team).Error)
	require.NoError(t, db.Exec(`INSERT INTO emails(id,team_id,campaign_id,contact_id,"to","from",subject,body,status,sent_at,created_at,test,is_deleted,click_tracking_enabled,smtp_config_id,category_id) SELECT md5('message'||i)::uuid,?,md5('edition'||(1+(1+(i-1)%10000)%15))::uuid,md5('person'||(1+(i-1)%10000))::uuid,'reader'||(1+(i-1)%10000)||'@example.com','news@example.com','Dispatch','','SENT','2026-09-03'::timestamptz+(i%6)*interval '1 day','2026-09-02',false,false,true,md5('smtp')::uuid,md5('category')::uuid FROM generate_series(1,50000) i`, team).Error)
	require.NoError(t, db.Exec(`INSERT INTO email_trackings(id,email_id,event,timestamp,url,is_deleted) SELECT md5('click'||i)::uuid,md5('message'||i)::uuid,'click','2026-09-03'::timestamptz+(i%6)*interval '1 day'+interval '1 hour','https://example.com/stories/'||(i%3),false FROM generate_series(1,50000) i WHERE i%7=0`).Error)
	require.NoError(t, db.Exec(`UPDATE campaigns SET newsletter_id=md5('newsletter-'||name)::uuid WHERE name IN ('September dispatch 3','September dispatch 6','September dispatch 9','September dispatch 12','September dispatch 15')`).Error)
	require.NoError(t, db.Exec(`UPDATE email_trackings SET device_type=CASE WHEN substring(id::text,1,1) IN ('0','1','2','3','4','5','6','7') THEN 'mobile' WHEN substring(id::text,1,1) IN ('8','9','a','b') THEN 'desktop' WHEN substring(id::text,1,1) IN ('c','d') THEN 'tablet' ELSE '' END, timestamp=timestamp+(ascii(substring(id::text,1,1))%20)*interval '1 hour'`).Error)
	require.NoError(t, db.Exec(`ANALYZE`).Error)
	f, err := analytics.Parse(url.Values{"from": {"2026-09-01"}, "to": {"2026-09-10"}}, team, time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	started := time.Now()
	r, err := analytics.Build(context.Background(), db, f)
	require.NoError(t, err)
	t.Logf("10,000 contacts / 50,000 messages report: %s", time.Since(started))
	require.EqualValues(t, 50000, r.Summary.Accepted)
	require.EqualValues(t, 10000, r.Summary.Reached)
	if path := os.Getenv("ANALYTICS_BROWSER_FIXTURE"); path != "" {
		pages := map[string]interface{}{}
		for _, kind := range []string{"lists", "sources", "campaigns", "domains", "links"} {
			f.Kind = kind
			for page := 1; page <= 2; page++ {
				f.Page = page
				t.Logf("Checking %s page %d", kind, page)
				p, e := analytics.Breakdown(context.Background(), db, f)
				require.NoError(t, e)
				pages[kind+strconv.Itoa(page)] = p
			}
		}
		f.Page = 1
		f.Limit = 20
		f.Cohort = "recent"
		people, e := analytics.People(context.Background(), db, f)
		require.NoError(t, e)
		var lists []map[string]interface{}
		require.NoError(t, db.Table("mailing_lists").Select("id,name").Order("name").Find(&lists).Error)
		data, _ := json.Marshal(map[string]interface{}{"report": r, "pages": pages, "people": people, "options": map[string]interface{}{"lists": lists, "tags": []string{}}})
		require.NoError(t, os.WriteFile(path, data, 0600))
	}
}

func TestTagSuppressionAndHistoryTombstones(t *testing.T) {
	db := database(t)
	team, list, tag, a, b := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	put(t, db, &models.MailingList{Base: base(list), TeamID: team, Name: "Filtered"})
	put(t, db, &models.Tag{Base: base(tag), TeamID: team, Name: "VIP"})
	for _, c := range []models.Contact{{Base: base(a), TeamID: team, ListID: list, Email: "a@example.com", Status: models.SubscriberStatusActive}, {Base: base(b), TeamID: team, ListID: list, Email: "b@example.com", Status: models.SubscriberStatusActive}} {
		put(t, db, &c)
	}
	require.NoError(t, db.Exec("INSERT INTO contact_tags(contact_id,tag_id) VALUES(?,?)", a, tag).Error)
	now := time.Now().UTC()
	f, e := analytics.Parse(url.Values{}, team, now)
	require.NoError(t, e)
	f.Tag = tag
	r, e := analytics.Build(context.Background(), db, f)
	require.NoError(t, e)
	require.EqualValues(t, 1, r.Summary.Active)
	require.False(t, r.Growth.Available)
	put(t, db, &models.SuppressionList{Base: base(uuid.NewString()), TeamID: team, EmailAddress: "a@example.com", IsActive: true, Reason: models.SuppressionReasonManual, AddedAt: now.Add(-time.Hour)})
	r, e = analytics.Build(context.Background(), db, f)
	require.NoError(t, e)
	require.Zero(t, r.Summary.Active)
	people, e := analytics.People(context.Background(), db, f)
	require.NoError(t, e)
	require.Zero(t, people.Total)
	require.NoError(t, db.Model(&models.MailingList{}).Where("id=?", list).UpdateColumn("is_deleted", true).Error)
	var deleted int64
	require.NoError(t, db.Table("subscription_events").Where("kind='list_status' AND is_deleted=true").Count(&deleted).Error)
	require.EqualValues(t, 2, deleted)
	require.NoError(t, db.Model(&models.Contact{}).Where("id=?", a).UpdateColumn("status", "UNSUBSCRIBED").Error)
	require.NoError(t, db.Model(&models.MailingList{}).Where("id=?", list).UpdateColumn("is_deleted", false).Error)
	var active int64
	require.NoError(t, db.Raw("SELECT COUNT(*) FROM (SELECT DISTINCT ON(contact_id) status,is_deleted FROM subscription_events ORDER BY contact_id,occurred_at DESC,id DESC) s WHERE status='ACTIVE' AND is_deleted=false").Scan(&active).Error)
	require.EqualValues(t, 1, active)
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Where("id=?", b).Delete(&models.Contact{}).Error)
	require.NoError(t, db.Table("subscription_events").Where("contact_id=? AND kind='delete' AND is_deleted=true", b).Count(&deleted).Error)
	require.EqualValues(t, 1, deleted)
}

func TestChartBucketsUseTenantScopeSendCohortsAndEventTime(t *testing.T) {
	db := database(t)
	team, foreign, list, otherList, campaign, edition, newsletter := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	put(t, db, &models.MailingList{Base: base(list), TeamID: team, Name: "Journal"})
	put(t, db, &models.MailingList{Base: base(otherList), TeamID: team, Name: "Other"})
	put(t, db, &models.Campaign{Base: base(campaign), TeamID: team, ListID: list, Name: "Launch", TemplateID: uuid.NewString(), SMTPConfigID: uuid.NewString()})
	put(t, db, &models.Campaign{Base: base(edition), TeamID: team, ListID: otherList, Name: "Edition", NewsletterID: &newsletter, TemplateID: uuid.NewString(), SMTPConfigID: uuid.NewString()})
	stamp := time.Date(2026, 9, 5, 23, 30, 0, 0, time.UTC)
	message, news, old, corrupt := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	for _, v := range []struct {
		id, team, campaign string
		sent               time.Time
	}{{message, team, campaign, stamp}, {news, team, edition, stamp}, {old, team, campaign, stamp.AddDate(0, 0, -7)}, {corrupt, foreign, campaign, stamp}} {
		put(t, db, &models.Email{Base: base(v.id), TeamID: v.team, CampaignID: v.campaign, To: "reader@example.com", From: "sender@example.com", SentAt: v.sent, Status: models.EmailStatusSent})
	}
	click := func(email, device string, at time.Time) {
		put(t, db, &models.EmailTracking{Base: base(uuid.NewString()), EmailID: email, Event: models.EmailTrackingEventClick, Timestamp: at, DeviceType: device})
	}
	click(message, "mobile", stamp.Add(time.Hour))
	click(message, "mobile", stamp.Add(2*time.Hour))
	click(news, "desktop", stamp.Add(time.Hour))
	click(old, "", stamp.Add(time.Hour))
	click(corrupt, "tablet", stamp.Add(time.Hour))    // Foreign email pointing to this team's campaign.
	click(message, "desktop", stamp.Add(-time.Hour))  // Before message was sent.
	click(message, "desktop", stamp.AddDate(0, 0, 8)) // Beyond report cutoff.
	f, err := analytics.Parse(url.Values{"from": {"2026-09-01"}, "to": {"2026-09-10"}, "timezone": {"Asia/Kolkata"}}, team, time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	r, err := analytics.Build(context.Background(), db, f)
	require.NoError(t, err)
	require.EqualValues(t, 2, r.Summary.Accepted)
	require.Len(t, r.Charts.Volume, 9)
	require.Len(t, r.Charts.ClickHours, 168)
	require.Len(t, r.Charts.Devices, 5)
	require.Equal(t, "2026-09-06", r.Charts.Volume[5].Date)
	require.EqualValues(t, 1, r.Charts.Volume[5].Campaigns)
	require.EqualValues(t, 1, r.Charts.Volume[5].Newsletters)
	require.EqualValues(t, 1, r.Charts.Volume[5].CampaignClicks)
	require.EqualValues(t, 1, r.Charts.Volume[5].NewsletterClicks)
	require.EqualValues(t, 4, r.Charts.ClickEvents)
	var deviceTotal, hourTotal, volumeTotal, clickedTotal int64
	for _, d := range r.Charts.Devices {
		deviceTotal += d.Count
		if d.Device == "unknown" {
			require.EqualValues(t, 1, d.Count)
		}
	}
	for _, h := range r.Charts.ClickHours {
		hourTotal += h.Count
		if h.Day == 6 && h.Hour == 6 {
			require.EqualValues(t, 3, h.Count)
		}
	}
	for _, v := range r.Charts.Volume {
		volumeTotal += v.Campaigns + v.Newsletters
		clickedTotal += v.CampaignClicks + v.NewsletterClicks
	}
	require.Equal(t, r.Charts.ClickEvents, deviceTotal)
	require.Equal(t, deviceTotal, hourTotal)
	require.Equal(t, r.Summary.Accepted, volumeTotal)
	require.Equal(t, r.Summary.ClickedMessages, clickedTotal)
	f.List = list
	r, err = analytics.Build(context.Background(), db, f)
	require.NoError(t, err)
	require.EqualValues(t, 1, r.Summary.Accepted)
	require.EqualValues(t, 3, r.Charts.ClickEvents)
	f.Campaign = edition
	r, err = analytics.Build(context.Background(), db, f)
	require.NoError(t, err)
	require.Zero(t, r.Summary.Accepted)
	require.Zero(t, r.Charts.ClickEvents)
}
