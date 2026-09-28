package marketing_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"gorm.io/driver/postgres"
	"kori/internal/automation"
	"kori/internal/automation/processors"
	"kori/internal/config"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/handlers"
	"kori/internal/marketing"
	"kori/internal/models"
	cryptoutil "kori/internal/utils/crypto"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	var dialect gorm.Dialector = sqlite.Open("file:" + uuid.NewString() + "?mode=memory&cache=shared")
	if url := os.Getenv("POSTHOOT_TEST_DATABASE_URL"); url != "" {
		admin, err := gorm.Open(postgres.Open(url), &gorm.Config{})
		require.NoError(t, err)
		schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		require.NoError(t, admin.Exec("CREATE SCHEMA "+schema).Error)
		t.Cleanup(func() { admin.Exec("DROP SCHEMA " + schema + " CASCADE"); raw, _ := admin.DB(); raw.Close() })
		dialect = postgres.Open(url + " search_path=" + schema)
	}
	db, err := gorm.Open(dialect, &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Team{}, &models.MailingList{}, &models.Contact{}, &models.Template{}, &models.SMTPConfig{}, &models.Campaign{}, &models.Email{}, &models.Newsletter{}, &models.ContactNote{}, &models.Form{}, &models.FormField{}, &models.FormSubmission{}, &models.FormReceipt{}, &models.FormRevision{}, &models.FormProgress{}, &models.FormAnalyticsEvent{}, &models.FormCompletionEvent{}, &models.SuppressionList{}, &models.File{}, &models.EmailCategory{}, &models.AutomationNodeEdge{}))
	raw, err := db.DB()
	require.NoError(t, err)
	raw.SetMaxOpenConns(1)
	t.Cleanup(func() { raw.Close() })
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	cryptoutil.PrivateKey = key
	cryptoutil.PublicKey = &key.PublicKey
	return db
}
func seed(t *testing.T, db *gorm.DB, value interface{}) {
	t.Helper()
	reflect.ValueOf(value).Elem().FieldByName("CreatedAt").Set(reflect.ValueOf(time.Now().UTC()))
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(value).Error)
}
func TestScheduleCalendarAndDST(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	at := time.Date(2026, 3, 7, 9, 0, 0, 0, loc)
	next, err := marketing.NextEdition(at, "DAILY", loc.String())
	require.NoError(t, err)
	require.Equal(t, 9, next.In(loc).Hour())
	require.Equal(t, 23*time.Hour, next.Sub(at))
	next, err = marketing.NextEdition(time.Date(2026, 1, 31, 9, 0, 0, 0, time.UTC), "MONTHLY", "UTC")
	require.NoError(t, err)
	require.Equal(t, 28, next.Day())
	require.Equal(t, time.February, next.Month())
	_, err = marketing.NextEdition(at, "HOURLY", "UTC")
	require.Error(t, err)
	_, err = marketing.NextEdition(at, "DAILY", "Mars/Olympus")
	require.Error(t, err)
}
func TestNewsletterEditionSnapshotIdempotencyAndSuppression(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	team := uuid.NewString()
	tplID := uuid.NewString()
	listID := uuid.NewString()
	senderID := uuid.NewString()
	category := uuid.NewString()
	seed(t, db, &models.MailingList{Base: models.Base{ID: listID}, TeamID: team, Name: "Audience"})
	seed(t, db, &models.Template{Base: models.Base{ID: tplID}, TeamID: team, CategoryID: category, Name: "Editorial", Subject: "Hello", HTMLBody: "<p>Hello {{first_name}}</p>"})
	password, err := cryptoutil.Encrypt("fixture-password")
	require.NoError(t, err)
	seed(t, db, &models.SMTPConfig{Base: models.Base{ID: senderID}, TeamID: team, FromEmail: "hello@example.com", SupportsTLS: true, Password: password})
	contactID := uuid.NewString()
	seed(t, db, &models.Contact{Base: models.Base{ID: contactID}, TeamID: team, ListID: listID, Email: "active@example.com", FirstName: "<script>", Status: models.SubscriberStatusActive})
	seed(t, db, &models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: listID, Email: "unsubscribed@example.com", Status: models.SubscriberStatusUnsubscribed})
	now := time.Now().UTC()
	due := now.Add(-time.Hour)
	n := models.Newsletter{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Weekly", Subject: "An edition", TemplateID: tplID, ListID: listID, SMTPConfigID: senderID, Status: "SCHEDULED", Cadence: "WEEKLY", Timezone: "UTC", NextSendAt: &due, PostalAddress: "42 Market Street, Bengaluru"}
	seed(t, db, &n)
	service := marketing.New(db, strings.Repeat("s", 32), "https://api.example.com")
	require.NoError(t, service.MaterializeDue(ctx, now))
	require.NoError(t, service.MaterializeDue(ctx, now))
	var editions []models.Campaign
	require.NoError(t, db.Where("newsletter_id = ?", n.ID).Find(&editions).Error)
	require.Len(t, editions, 1)
	require.NoError(t, db.Model(&models.Template{}).Where("id = ?", tplID).Update("html_body", "changed").Error)
	require.Equal(t, "<p>Hello {{first_name}}</p>", editions[0].HTMLBody)
	require.NoError(t, service.PrepareEdition(ctx, editions[0].ID))
	require.NoError(t, service.PrepareEdition(ctx, editions[0].ID))
	var messages []models.Email
	require.NoError(t, db.Find(&messages).Error)
	require.Len(t, messages, 1)
	require.Equal(t, contactID, messages[0].ContactID)
	require.Contains(t, messages[0].UnsubscribeURL, "https://api.example.com/public/unsubscribe/")
	var updated models.Newsletter
	require.NoError(t, db.First(&updated, "id = ?", n.ID).Error)
	require.Equal(t, 1, updated.Editions)
	require.True(t, updated.NextSendAt.After(now))
	require.False(t, service.Owns(ctx, &models.Template{}, tplID, uuid.NewString()))
	require.True(t, service.CheckToken(team, contactID, service.Token(team, contactID)))
	require.False(t, service.CheckToken(team, uuid.NewString(), service.Token(team, contactID)))
}
func TestPublicCaptureConsentRetryAndTenantIsolation(t *testing.T) {
	db := testDB(t)
	team := uuid.NewString()
	listID := uuid.NewString()
	seed(t, db, &models.MailingList{Base: models.Base{ID: listID}, TeamID: team, Name: "Community"})
	handler := handlers.NewMarketingHandler(db, strings.Repeat("s", 32), "https://api.example.com")
	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error { c.Set("teamID", team); return next(c) }
	})
	e.POST("/forms", handler.SaveForm)
	e.GET("/forms", handler.Forms)
	e.POST("/public/:slug", handler.SubmitForm)
	e.PUT("/contacts/:id/stage", handler.UpdateContactStage)
	e.POST("/contacts/:id/notes", handler.Notes)
	call := func(method, path string, payload interface{}) *httptest.ResponseRecorder {
		b, err := json.Marshal(payload)
		require.NoError(t, err)
		req := httptest.NewRequest(method, path, strings.NewReader(string(b)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec
	}
	rec := call("POST", "/forms", map[string]interface{}{"name": "Newsletter signup", "listId": listID, "status": "PUBLISHED", "fields": []map[string]interface{}{{"label": "Email", "type": "EMAIL", "key": "email", "required": true}}})
	require.Equal(t, 200, rec.Code, rec.Body.String())
	var form models.Form
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &form))
	id := uuid.NewString()
	input := map[string]interface{}{"fields": map[string]string{"email": "reader@example.com"}, "consent": false, "requestId": id}
	rec = call("POST", "/public/"+form.Slug, input)
	require.Equal(t, 400, rec.Code)
	input["consent"] = true
	rec = call("POST", "/public/"+form.Slug, input)
	require.Equal(t, 200, rec.Code, rec.Body.String())
	rec = call("POST", "/public/"+form.Slug, input)
	require.Equal(t, 200, rec.Code)
	var total int64
	require.NoError(t, db.Model(&models.FormSubmission{}).Count(&total).Error)
	require.EqualValues(t, 1, total)
	var contact models.Contact
	require.NoError(t, db.Where("email = ?", "reader@example.com").First(&contact).Error)
	require.Equal(t, team, contact.TeamID)
	require.NoError(t, db.Model(&contact).UpdateColumn("status", models.SubscriberStatusUnsubscribed).Error)
	input["requestId"] = uuid.NewString()
	rec = call("POST", "/public/"+form.Slug, input)
	require.Equal(t, 200, rec.Code)
	require.NoError(t, db.First(&contact, "id = ?", contact.ID).Error)
	require.Equal(t, models.SubscriberStatusUnsubscribed, contact.Status)
	foreign := models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: uuid.NewString(), ListID: listID, Email: "other@example.com", Status: models.SubscriberStatusActive}
	seed(t, db, &foreign)
	rec = call("PUT", "/contacts/"+foreign.ID+"/stage", map[string]interface{}{"email": "stolen@example.com", "lifecycleStage": "LEAD"})
	require.Equal(t, 404, rec.Code)
	rec = call("POST", "/contacts/"+foreign.ID+"/notes", map[string]string{"body": "not allowed"})
	require.Equal(t, 404, rec.Code)
	rec = call("POST", "/forms", map[string]interface{}{"name": "Cross-tenant", "listId": uuid.NewString(), "status": "PUBLISHED", "fields": []map[string]interface{}{{"label": "Email", "type": "EMAIL", "key": "email", "required": true}}})
	require.Equal(t, 400, rec.Code)
}

func TestMonthlyAnchorReturnsToOriginalDay(t *testing.T) {
	feb, err := marketing.NextEdition(time.Date(2026, 1, 31, 9, 0, 0, 0, time.UTC), "MONTHLY", "UTC", 31)
	require.NoError(t, err)
	march, err := marketing.NextEdition(*feb, "MONTHLY", "UTC", 31)
	require.NoError(t, err)
	require.Equal(t, time.Date(2026, 3, 31, 9, 0, 0, 0, time.UTC), *march)
}

type fileURL string

func (url fileURL) GetSignedURL(context.Context, string, time.Duration) (string, error) {
	return string(url), nil
}
func TestNewsletterReadsExistingEditorTemplate(t *testing.T) {
	db := testDB(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("<p>Existing editor content</p>")) }))
	defer server.Close()
	models.RegisterFileURLGenerator(fileURL(server.URL))
	defer models.RegisterFileURLGenerator(nil)
	teamID := uuid.NewString()
	file := models.File{Base: models.Base{ID: uuid.NewString()}, TeamID: teamID, Path: "template.html"}
	seed(t, db, &file)
	template := models.Template{TeamID: teamID, HtmlFileID: file.ID}
	body, err := marketing.TemplateHTML(db, &template)
	require.NoError(t, err)
	require.Equal(t, "<p>Existing editor content</p>", body)
	template.TeamID = uuid.NewString()
	_, err = marketing.TemplateHTML(db, &template)
	require.Error(t, err)
}
func TestStarterUsesExistingTemplateAndDesign(t *testing.T) {
	for _, starter := range marketing.Starters() {
		raw, err := base64.StdEncoding.DecodeString(starter.DesignJSON)
		require.NoError(t, err)
		var design map[string]interface{}
		require.NoError(t, json.Unmarshal(raw, &design))
		require.NotNil(t, design["body"])
	}
	db := testDB(t)
	handler := handlers.NewMarketingHandler(db, strings.Repeat("x", 32), "https://api.example.com")
	request := httptest.NewRequest("POST", "/starters", strings.NewReader(`{"starterKey":"editorial"}`))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(request, rec)
	teamID := uuid.NewString()
	ctx.Set("teamID", teamID)
	require.NoError(t, handler.ImportTemplateStarter(ctx))
	var template models.Template
	require.NoError(t, db.Where("team_id = ?", teamID).First(&template).Error)
	require.NotEmpty(t, template.DesignJSON)
	require.NotEmpty(t, template.HTMLBody)
}
func TestAutomationEmailRetryUsesOneOutboxRecord(t *testing.T) {
	db := testDB(t)
	teamID, senderID, categoryID, templateID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	password, err := cryptoutil.Encrypt("fixture")
	require.NoError(t, err)
	seed(t, db, &models.EmailCategory{Base: models.Base{ID: categoryID}, TeamID: teamID, Name: "Transactional"})
	seed(t, db, &models.SMTPConfig{Base: models.Base{ID: senderID}, TeamID: teamID, Password: password, FromEmail: "hello@example.com", SupportsTLS: true})
	seed(t, db, &models.Template{Base: models.Base{ID: templateID}, TeamID: teamID, CategoryID: categoryID, HTMLBody: "<p>{{first_name}}</p>", Subject: "Hello {{first_name}}"})
	cfg := &config.Config{}
	cfg.JWT.Secret = strings.Repeat("x", 32)
	processor := processors.NewEmailProcessor(db, nil, cfg)
	data, _ := json.Marshal(map[string]string{"templateId": templateID, "smtpConfigId": senderID})
	node := &models.AutomationNode{Base: models.Base{ID: uuid.NewString()}, Type: models.NodeTypeEmail, Data: data}
	contact := &models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: teamID, Email: "reader@example.com", FirstName: "<b>$1</b>", Status: models.SubscriberStatusActive}
	ctx := &automation.ExecutionContext{TeamID: teamID, ContactID: contact.ID, Contact: contact, AutomationID: uuid.NewString(), Execution: &models.AutomationExecution{Base: models.Base{ID: uuid.NewString()}}}
	first, err := processor.Process(ctx, node)
	require.NoError(t, err)
	second, err := processor.Process(ctx, node)
	require.NoError(t, err)
	require.Equal(t, first.Data["emailId"], second.Data["emailId"])
	var rows []models.Email
	require.NoError(t, db.Find(&rows).Error)
	require.Len(t, rows, 1)
	decoded, err := base64.StdEncoding.DecodeString(rows[0].Body)
	require.NoError(t, err)
	require.Contains(t, string(decoded), "&lt;b&gt;$1&lt;/b&gt;")
	require.Equal(t, models.EmailStatusPending, rows[0].Status)
}

func TestNewsletterUnsubscribeAttributionIsSignedAndIdempotent(t *testing.T) {
	db := testDB(t)
	require.NoError(t, db.AutoMigrate(&models.EmailTracking{}))
	team, list, contactID, messageID, campaignID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	seed(t, db, &models.MailingList{Base: models.Base{ID: list}, TeamID: team, Name: "Newsletter"})
	seed(t, db, &models.Contact{Base: models.Base{ID: contactID}, TeamID: team, ListID: list, Email: "reader@example.com", Status: models.SubscriberStatusActive})
	seed(t, db, &models.Email{Base: models.Base{ID: messageID}, TeamID: team, ContactID: contactID, CampaignID: campaignID, To: "reader@example.com", From: "news@example.com", SMTPConfigID: uuid.NewString(), CategoryID: uuid.NewString(), Status: models.EmailStatusSent})
	secret := strings.Repeat("s", 32)
	service := marketing.New(db, secret, "https://api.example.com")
	handler := handlers.NewMarketingHandler(db, secret, "https://api.example.com")
	e := echo.New()
	e.GET("/unsubscribe/:team/:contact/:token", handler.Unsubscribe)
	e.POST("/unsubscribe/:team/:contact/:token", handler.Unsubscribe)
	target := "/unsubscribe/" + team + "/" + contactID + "/" + service.MessageToken(team, contactID, messageID) + "?message=" + messageID
	call := func(method, path string) int {
		r := httptest.NewRecorder()
		e.ServeHTTP(r, httptest.NewRequest(method, path, nil))
		return r.Code
	}
	require.Equal(t, 200, call("GET", target))
	var count int64
	require.NoError(t, db.Model(&models.EmailTracking{}).Count(&count).Error)
	require.Zero(t, count)
	require.Equal(t, 404, call("POST", strings.Replace(target, "?message="+messageID, "?message="+uuid.NewString(), 1)))
	require.Equal(t, 200, call("POST", target))
	require.Equal(t, 200, call("POST", target))
	require.NoError(t, db.Model(&models.EmailTracking{}).Where("email_id=? AND event='unsubscribe'", messageID).Count(&count).Error)
	require.EqualValues(t, 1, count)
	var contact models.Contact
	require.NoError(t, db.First(&contact, "id=?", contactID).Error)
	require.Equal(t, models.SubscriberStatusUnsubscribed, contact.Status)
}
