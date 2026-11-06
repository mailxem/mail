package marketing_test

import (
	"context"
	"encoding/base64"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kori/internal/handlers"
	"kori/internal/marketing"
	"kori/internal/models"
	cryptoutil "kori/internal/utils/crypto"
)

func TestListCleanupAsksOncePausesAndKeepLinkReactivates(t *testing.T) {
	db := testDB(t)
	require.NoError(t, db.AutoMigrate(&models.EmailTracking{}))
	ctx := context.Background()
	now := time.Now().UTC()
	day := 24 * time.Hour
	create := func(value interface{}) { require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(value).Error) }
	team, list, offList, campaign, sender, category, tpl := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	off := false
	create(&models.Team{Base: models.Base{ID: team}, Name: "Acme Weekly"})
	create(&models.MailingList{Base: models.Base{ID: list}, TeamID: team, Name: "Readers"})
	create(&models.MailingList{Base: models.Base{ID: offList}, TeamID: team, Name: "Cleanup off", ListCleanup: &off})
	create(&models.Campaign{Base: models.Base{ID: campaign}, TeamID: team, Name: "Issue", ListID: list, TemplateID: tpl, SMTPConfigID: sender, PostalAddress: "42 Market Street, Bengaluru"})
	reader := func(listID string, sends int, firstAgo time.Duration) string {
		id := uuid.NewString()
		create(&models.Contact{Base: models.Base{ID: id, CreatedAt: now.Add(-200 * day)}, TeamID: team, ListID: listID, Email: id + "@example.com", FirstName: "Ana", Status: models.SubscriberStatusActive})
		for i := 0; i < sends; i++ {
			at := now.Add(-firstAgo + time.Duration(i)*day)
			create(&models.Email{Base: models.Base{ID: uuid.NewString(), CreatedAt: at}, TeamID: team, ContactID: id, CampaignID: campaign, To: id + "@example.com", From: "news@example.com", SMTPConfigID: sender, CategoryID: category, Status: models.EmailStatusSent})
		}
		return id
	}
	quiet := reader(list, 5, 70*day)
	clicker := reader(list, 5, 70*day)
	opener := reader(list, 5, 70*day)
	recent := reader(list, 5, 30*day)
	few := reader(list, 4, 70*day)
	quietOff := reader(offList, 5, 70*day)
	create(&models.EmailTracking{Base: models.Base{ID: uuid.NewString()}, EmailID: uuid.NewString(), ContactID: clicker, Event: models.EmailTrackingEventClick, Timestamp: now.Add(-10 * day)})
	create(&models.EmailTracking{Base: models.Base{ID: uuid.NewString()}, EmailID: uuid.NewString(), ContactID: opener, Event: models.EmailTrackingEventOpen, Timestamp: now.Add(-10 * day)})

	secret := strings.Repeat("s", 32)
	service := marketing.New(db, secret, "https://api.example.com")
	asked := func() []models.Email {
		var emails []models.Email
		require.NoError(t, db.Where("campaign_id IS NULL").Find(&emails).Error)
		return emails
	}
	status := func(id string) models.Contact {
		var c models.Contact
		require.NoError(t, db.First(&c, "id = ?", id).Error)
		return c
	}

	// Only the reader who ignored 5 sends over 60+ days is asked, and only once.
	require.NoError(t, service.CleanUpLists(ctx, now))
	require.NoError(t, service.CleanUpLists(ctx, now.Add(time.Hour)))
	require.NoError(t, service.CleanUpLists(ctx, now.Add(13*day)))
	emails := asked()
	require.Len(t, emails, 1)
	require.Equal(t, quiet, emails[0].ContactID)
	require.Equal(t, "Still want emails from Acme Weekly?", emails[0].Subject)
	require.Equal(t, models.EmailStatusPending, emails[0].Status)
	require.NotNil(t, emails[0].DeliveryKey)
	body, err := base64.StdEncoding.DecodeString(emails[0].Body)
	require.NoError(t, err)
	keepURL := "https://api.example.com/public/keep/" + team + "/" + quiet + "/" + service.KeepToken(team, quiet)
	require.Contains(t, string(body), keepURL)
	require.Contains(t, string(body), "Yes, keep sending")
	require.Contains(t, string(body), emails[0].UnsubscribeURL)
	require.Contains(t, emails[0].UnsubscribeURL, "https://api.example.com/public/unsubscribe/")
	for _, id := range []string{clicker, opener, recent, few, quietOff} {
		require.Nil(t, status(id).ReengageSentAt, id)
	}

	// No answer within 14 days: paused, not unsubscribed, and out of the subscriber count.
	require.NoError(t, service.CleanUpLists(ctx, now.Add(15*day)))
	require.Equal(t, models.SubscriberStatusInactive, status(quiet).Status)
	require.Equal(t, models.SubscriberStatusActive, status(quietOff).Status)
	var readers models.MailingList
	require.NoError(t, db.First(&readers, "id = ?", list).Error)
	require.EqualValues(t, 4, readers.SubscribersCount)
	require.Len(t, asked(), 1)

	// Paused readers get no newsletter edition.
	create(&models.Template{Base: models.Base{ID: tpl}, TeamID: team, CategoryID: category, Name: "Issue", Subject: "Hi", HTMLBody: "<p>Hi</p>"})
	password, err := cryptoutil.Encrypt("fixture-password")
	require.NoError(t, err)
	create(&models.SMTPConfig{Base: models.Base{ID: sender}, TeamID: team, FromEmail: "news@example.com", SupportsTLS: true, Password: password})
	newsletter := uuid.NewString()
	edition := models.Campaign{Base: models.Base{ID: uuid.NewString()}, NewsletterID: &newsletter, TeamID: team, Name: "Edition", ListID: list, TemplateID: tpl, SMTPConfigID: sender, Status: models.CampaignStatusScheduled, HTMLBody: "<p>Hi</p>", Subject: "Edition", PostalAddress: "42 Market Street"}
	create(&edition)
	require.NoError(t, service.PrepareEdition(ctx, edition.ID))
	var editionTo []string
	require.NoError(t, db.Model(&models.Email{}).Where("campaign_id = ?", edition.ID).Pluck("contact_id", &editionTo).Error)
	require.Len(t, editionTo, 4)
	require.NotContains(t, editionTo, quiet)

	// The keep link needs a POST and a valid signature, then turns mail back on.
	e := echo.New()
	handler := handlers.NewMarketingHandler(db, secret, "https://api.example.com")
	e.GET("/public/keep/:team/:contact/:token", handler.KeepSending)
	e.POST("/public/keep/:team/:contact/:token", handler.KeepSending)
	call := func(method, path string) int {
		r := httptest.NewRecorder()
		e.ServeHTTP(r, httptest.NewRequest(method, path, nil))
		return r.Code
	}
	path := strings.TrimPrefix(keepURL, "https://api.example.com")
	require.Equal(t, 200, call("GET", path))
	require.Equal(t, models.SubscriberStatusInactive, status(quiet).Status)
	require.Equal(t, 404, call("POST", "/public/keep/"+team+"/"+quiet+"/"+service.KeepToken(team, clicker)))
	require.Equal(t, 200, call("POST", path))
	kept := status(quiet)
	require.Equal(t, models.SubscriberStatusActive, kept.Status)
	require.Nil(t, kept.ReengageSentAt)
	// Their count restarts, so they aren't asked again straight away.
	require.NoError(t, service.CleanUpLists(ctx, now.Add(16*day)))
	require.Len(t, asked(), 1)
}

func TestListCleanupClickReactivatesPausedReader(t *testing.T) {
	db := testDB(t)
	require.NoError(t, db.AutoMigrate(&models.EmailTracking{}))
	now := time.Now().UTC()
	team, list, id := uuid.NewString(), uuid.NewString(), uuid.NewString()
	paused := now.Add(-48 * time.Hour)
	require.NoError(t, db.Create(&models.MailingList{Base: models.Base{ID: list}, TeamID: team, Name: "Readers"}).Error)
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&models.Contact{Base: models.Base{ID: id}, TeamID: team, ListID: list, Email: "back@example.com", Status: models.SubscriberStatusInactive, CleanupFrom: &paused}).Error)
	service := marketing.New(db, strings.Repeat("s", 32), "https://api.example.com")
	require.NoError(t, service.CleanUpLists(context.Background(), now))
	var c models.Contact
	require.NoError(t, db.First(&c, "id = ?", id).Error)
	require.Equal(t, models.SubscriberStatusInactive, c.Status)
	require.NoError(t, db.Create(&models.EmailTracking{EmailID: uuid.NewString(), ContactID: id, Event: models.EmailTrackingEventClick, Timestamp: now.Add(-time.Hour)}).Error)
	require.NoError(t, service.CleanUpLists(context.Background(), now))
	require.NoError(t, db.First(&c, "id = ?", id).Error)
	require.Equal(t, models.SubscriberStatusActive, c.Status)
}
