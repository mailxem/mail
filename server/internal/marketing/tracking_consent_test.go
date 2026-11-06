package marketing_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"kori/internal/automation"
	"kori/internal/automation/processors"
	"kori/internal/config"
	"kori/internal/handlers"
	"kori/internal/marketing"
	"kori/internal/models"
	"kori/internal/utils"
	cryptoutil "kori/internal/utils/crypto"
)

func TestEditionTrackingFollowsSwitchAndConsent(t *testing.T) {
	db := testDB(t)
	require.NoError(t, db.AutoMigrate(&models.TrackingConsent{}))
	ctx := context.Background()
	team, list, tpl, sender := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	seed(t, db, &models.Team{Base: models.Base{ID: team}, Name: "Acme News", TrackingConsentMode: models.TrackingConsentOptIn})
	seed(t, db, &models.MailingList{Base: models.Base{ID: list}, TeamID: team, Name: "Readers"})
	seed(t, db, &models.Template{Base: models.Base{ID: tpl}, TeamID: team, CategoryID: uuid.NewString(), Name: "T", Subject: "S", HTMLBody: "x"})
	password, err := cryptoutil.Encrypt("fixture-password")
	require.NoError(t, err)
	seed(t, db, &models.SMTPConfig{Base: models.Base{ID: sender}, TeamID: team, FromEmail: "news@example.com", SupportsTLS: true, Password: password})
	ada, ben := uuid.NewString(), uuid.NewString()
	seed(t, db, &models.Contact{Base: models.Base{ID: ada}, TeamID: team, ListID: list, Email: "ada@example.com", Status: models.SubscriberStatusActive})
	seed(t, db, &models.Contact{Base: models.Base{ID: ben}, TeamID: team, ListID: list, Email: "ben@example.com", Status: models.SubscriberStatusActive})
	require.NoError(t, models.SetTrackingConsent(db, team, "ben@example.com", true, "test"))
	service := marketing.New(db, strings.Repeat("s", 32), "https://api.example.com")
	t.Setenv("PUBLIC_API_URL", "https://api.example.com")
	oldRender := utils.RenderConfig
	t.Cleanup(func() { utils.RenderConfig = oldRender })
	utils.RenderConfig = &config.Config{JWT: config.JWTConfig{Secret: strings.Repeat("s", 32)}}
	newsletter := uuid.NewString()
	send := func(track *bool) map[string]models.Email {
		c := models.Campaign{Base: models.Base{ID: uuid.NewString()}, NewsletterID: &newsletter, TeamID: team, TemplateID: tpl, ListID: list, SMTPConfigID: sender, Name: "Edition", Status: models.CampaignStatusScheduled, HTMLBody: `<a href="https://example.com/story">Read</a>`, PostalAddress: "42 Market Street", TrackEngagement: track}
		seed(t, db, &c)
		require.NoError(t, service.PrepareEdition(ctx, c.ID))
		var rows []models.Email
		require.NoError(t, db.Where("campaign_id = ?", c.ID).Find(&rows).Error)
		byContact := map[string]models.Email{}
		for _, r := range rows {
			byContact[r.ContactID] = r
		}
		return byContact
	}
	body := func(e models.Email) string {
		b, err := utils.EmailHTML(db, &e)
		require.NoError(t, err)
		return b
	}

	// opt_in: only the reader who agreed is tracked; the other still gets a working unsubscribe link.
	sent := send(nil)
	require.False(t, *sent[ada].ClickTrackingEnabled)
	require.NotContains(t, body(sent[ada]), "/t/open")
	require.NotContains(t, body(sent[ada]), "/t/click/")
	require.Contains(t, body(sent[ada]), sent[ada].UnsubscribeURL)
	require.True(t, *sent[ben].ClickTrackingEnabled)
	require.Contains(t, body(sent[ben]), "/t/open?token=")

	// opt_out: a reader objects on the preferences page and is no longer tracked.
	require.NoError(t, db.Model(&models.Team{}).Where("id = ?", team).Update("tracking_consent_mode", models.TrackingConsentOptOut).Error)
	handler := handlers.NewMarketingHandler(db, strings.Repeat("s", 32), "https://api.example.com")
	e := echo.New()
	e.GET("/unsubscribe/:team/:contact/:token", handler.Unsubscribe)
	e.POST("/unsubscribe/:team/:contact/:token", handler.Unsubscribe)
	target := "/unsubscribe/" + team + "/" + ada + "/" + service.Token(team, ada)
	r := httptest.NewRecorder()
	e.ServeHTTP(r, httptest.NewRequest("GET", target, nil))
	require.Equal(t, 200, r.Code)
	require.Contains(t, r.Body.String(), "Let Acme News see when I open emails and click links")
	require.Contains(t, r.Body.String(), "checked") // opt_out defaults to allowed
	r = httptest.NewRecorder()
	req := httptest.NewRequest("POST", target, strings.NewReader(url.Values{"tracking_choice": {"1"}}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	e.ServeHTTP(r, req)
	require.Equal(t, 200, r.Code)
	require.Contains(t, r.Body.String(), "will not be tracked")
	var contact models.Contact
	require.NoError(t, db.First(&contact, "id = ?", ada).Error)
	require.Equal(t, models.SubscriberStatusActive, contact.Status, "saving a preference does not unsubscribe")
	sent = send(nil)
	require.False(t, *sent[ada].ClickTrackingEnabled)
	require.True(t, *sent[ben].ClickTrackingEnabled)

	// Switch off: nobody is tracked, whatever the consent.
	require.NoError(t, db.Model(&models.Team{}).Where("id = ?", team).Update("tracking_consent_mode", models.TrackingConsentNone).Error)
	off := false
	sent = send(&off)
	require.False(t, *sent[ben].ClickTrackingEnabled)
	require.NotContains(t, body(sent[ben]), "/t/open")

	// Mode none hides the choice on the preferences page.
	r = httptest.NewRecorder()
	e.ServeHTTP(r, httptest.NewRequest("GET", target, nil))
	require.NotContains(t, r.Body.String(), "see when I open")
}

func TestAutomationEmailTrackingSwitchAndTransactionalConsent(t *testing.T) {
	db := testDB(t)
	require.NoError(t, db.AutoMigrate(&models.TrackingConsent{}))
	teamID, senderID, categoryID, templateID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	seed(t, db, &models.Team{Base: models.Base{ID: teamID}, Name: "Acme", TrackingConsentMode: models.TrackingConsentOptIn})
	password, err := cryptoutil.Encrypt("fixture")
	require.NoError(t, err)
	seed(t, db, &models.EmailCategory{Base: models.Base{ID: categoryID}, TeamID: teamID, Name: "Transactional"})
	seed(t, db, &models.SMTPConfig{Base: models.Base{ID: senderID}, TeamID: teamID, Password: password, FromEmail: "hello@example.com", SupportsTLS: true})
	seed(t, db, &models.Template{Base: models.Base{ID: templateID}, TeamID: teamID, CategoryID: categoryID, HTMLBody: `<a href="https://example.com/r">Receipt</a>`, Subject: "Receipt"})
	cfg := &config.Config{}
	cfg.JWT.Secret = strings.Repeat("x", 32)
	processor := processors.NewEmailProcessor(db, nil, cfg)
	oldRender := utils.RenderConfig
	t.Cleanup(func() { utils.RenderConfig = oldRender })
	utils.RenderConfig = cfg
	run := func(data map[string]any) models.Email {
		raw, _ := json.Marshal(data)
		node := &models.AutomationNode{Base: models.Base{ID: uuid.NewString()}, Type: models.NodeTypeEmail, Data: raw}
		contact := &models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: teamID, Email: "reader@example.com", Status: models.SubscriberStatusActive}
		ctx := &automation.ExecutionContext{TeamID: teamID, ContactID: contact.ID, Contact: contact, AutomationID: uuid.NewString(), Execution: &models.AutomationExecution{Base: models.Base{ID: uuid.NewString()}}}
		result, err := processor.Process(ctx, node)
		require.NoError(t, err)
		var email models.Email
		require.NoError(t, db.First(&email, "id = ?", result.Data["emailId"]).Error)
		return email
	}
	decode := func(e models.Email) string {
		b, err := utils.EmailHTML(db, &e)
		require.NoError(t, err)
		return b
	}
	// Transactional email ignores the opt_in consent mode.
	tracked := run(map[string]any{"templateId": templateID, "smtpConfigId": senderID})
	require.True(t, *tracked.ClickTrackingEnabled)
	require.Contains(t, decode(tracked), "/t/open?token=")
	// The per-step switch still applies.
	untracked := run(map[string]any{"templateId": templateID, "smtpConfigId": senderID, "trackEngagement": false})
	require.False(t, *untracked.ClickTrackingEnabled)
	require.NotContains(t, decode(untracked), "/t/open")
	require.NotContains(t, decode(untracked), "/t/click/")
}
