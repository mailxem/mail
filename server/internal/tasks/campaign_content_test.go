package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
	"kori/internal/affiliates"
	"kori/internal/billing"
	"kori/internal/config"
	"kori/internal/db"
	"kori/internal/models"
	"kori/internal/utils"
	"kori/internal/utils/base64"
)

// template50KB is a newsletter-sized template with links on every row.
var template50KB = func() string {
	row := `<tr><td style="padding:12px;font-family:Arial,sans-serif;color:#333">Hello {{first_name}}, here is <a href="https://example.com/article?id=%d">story %d</a> from {{company}}.</td></tr>`
	var b strings.Builder
	b.WriteString(`<html><body><table width="600">`)
	for i := 0; b.Len() < 50*1024; i++ {
		fmt.Fprintf(&b, row, i, i)
	}
	b.WriteString(`</table></body></html>`)
	return b.String()
}()

// A managed campaign stores its HTML once and, per email, only the variables;
// what is sent and shown is byte for byte what the old per-email body held,
// with and without tracking and the Free plan's badge. Deploy: all API and
// worker processes together, and create the emails.content_id index
// CONCURRENTLY first.
func TestCampaignStoresContentOnceAndRendersTheOldBody(t *testing.T) {
	for _, free := range []bool{false, true} {
		for _, track := range []bool{true, false} {
			t.Run(fmt.Sprintf("badge=%v/tracked=%v", free, track), func(t *testing.T) {
				campaignContent(t, free, track)
			})
		}
	}
}

func campaignContent(t *testing.T, free, track bool) {
	pageDatabases(t, func(t *testing.T, database *gorm.DB) {
		must := func(err error) {
			t.Helper()
			if err != nil {
				t.Fatal(err)
			}
		}
		must(database.AutoMigrate(&models.Campaign{}, &models.Template{}, &models.MailingList{}, &models.EmailContent{}, &models.File{}, &models.Team{}, &models.User{}, &models.Affiliate{}, &models.TrackingConsent{}, &models.Author{}))
		cfg := &config.Config{JWT: config.JWTConfig{Secret: "campaign-secret"}, Server: config.ServerConfig{PublicURL: "https://api.example.com"}}
		oldDB, oldCfg, oldBatch := db.DB, utils.RenderConfig, utils.ManagedDeliveryBatch
		t.Cleanup(func() { db.DB, utils.RenderConfig, utils.ManagedDeliveryBatch = oldDB, oldCfg, oldBatch })
		db.DB, utils.RenderConfig = database, cfg
		var submitted [][]byte
		pages := 0
		utils.ManagedDeliveryBatch = func(emails []*models.Email, raws [][]byte) []error {
			pages++
			submitted = append(submitted, raws...)
			return make([]error, len(emails))
		}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			plan := "xem-starter"
			if free {
				plan = "xem-free"
			}
			fmt.Fprintf(w, `{"plan_id":%q,"status":"active"}`, plan)
		}))
		t.Cleanup(srv.Close)
		billing.Enabled = true
		billing.Use(&billing.Reporter{URL: srv.URL, HTTP: srv.Client(), Now: time.Now})
		t.Cleanup(func() { billing.Enabled = false; billing.Use(nil) })

		const n = 100
		team, list, owner := uuid.NewString(), uuid.NewString(), uuid.NewString()
		quiet := database.Session(&gorm.Session{SkipHooks: true})
		must(quiet.Create(&models.Team{Base: models.Base{ID: team}, Name: "Team", OwnerUserID: owner}).Error)
		must(quiet.Create(&models.User{Base: models.Base{ID: owner}, Email: owner + "@example.com", Password: "x", Role: models.UserRoleAdmin, TeamID: team}).Error)
		sender := models.SMTPConfig{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Provider: "MANAGED", Host: "managed", Port: 587, FromEmail: "news@example.com", IsActive: true, MaxSendRate: 50}
		must(quiet.Create(&sender).Error)
		must(quiet.Create(&models.MailingList{Base: models.Base{ID: list}, TeamID: team, Name: "Readers"}).Error)
		tpl := models.Template{Base: models.Base{ID: uuid.NewString()}, TeamID: team, CategoryID: uuid.NewString(), Name: "Weekly", Subject: "News for {{first_name}}", HTMLBody: template50KB}
		must(quiet.Create(&tpl).Error)
		for i := range n {
			must(quiet.Create(&models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: list, Email: fmt.Sprintf("reader%d@example.com", i), FirstName: fmt.Sprintf("Reader <%d>", i), Company: "Acme & Co", Status: models.SubscriberStatusActive}).Error)
		}
		staged := false
		campaign := models.Campaign{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Weekly", ListID: list, TemplateID: tpl.ID, SMTPConfigID: sender.ID, Status: models.CampaignStatusScheduled, BatchSize: 50, Data: []byte(`{"company":"Acme & Co"}`), SendInStages: &staged, TrackEngagement: &track}
		must(quiet.Create(&campaign).Error)

		payload, _ := json.Marshal(CampaignTask{CampaignID: campaign.ID, BatchSize: 50})
		must(NewTaskHandler(database, cfg).HandleCampaignProcess(context.Background(), asynq.NewTask(TaskTypeCampaignProcess, payload)))

		badge := affiliates.Badge(context.Background(), database, team)
		if free != (badge != "") {
			t.Fatalf("badge %q for free=%v", badge, free)
		}
		var contents []models.EmailContent
		must(database.Find(&contents).Error)
		if len(contents) != 1 || contents[0].HTML != template50KB || contents[0].Kind != utils.ContentCampaign || contents[0].Badge != badge {
			t.Fatalf("want one campaign content row holding the template and badge, got %d", len(contents))
		}
		var emails []models.Email
		must(database.Where("campaign_id = ?", campaign.ID).Find(&emails).Error)
		if len(emails) != n || len(submitted) != n || pages != 1 {
			t.Fatalf("%d emails, %d submitted in %d pages; want %d in one", len(emails), len(submitted), pages, n)
		}
		var newBytes, oldBytes, bodies, mime, perEmail int
		for _, raw := range submitted {
			if !bytes.Contains(raw, []byte("Message-ID: <")) || !bytes.Contains(raw, []byte("List-Unsubscribe-Post: List-Unsubscribe=One-Click")) {
				t.Fatal("submitted MIME is missing its headers")
			}
			mime += len(raw)
		}
		for _, e := range emails {
			if e.Body != "" || e.ContentID == nil || *e.ContentID != contents[0].ID || *e.ClickTrackingEnabled != track {
				t.Fatalf("email %s stored a body, no content or the wrong tracking", e.ID)
			}
			var variables map[string]string
			must(json.Unmarshal(e.Data, &variables))
			// The body HandleCampaignProcess used to store, same call.
			oldBody := utils.ReplaceVariables(affiliates.WithBadge(tpl.HTMLBody, badge), variables, e.ID, cfg, track, true)
			html, err := utils.EmailHTML(database, &e)
			must(err)
			if base64.EncodeToBase64(html) != oldBody {
				t.Fatalf("email %s renders differently from the body it used to store", e.ID)
			}
			// Variable-size bytes this email writes: subject, data, and before,
			// the base64 body; now the content ID and a share of the content row.
			small := len(e.Subject) + len(e.Data)
			newBytes += small + len(*e.ContentID) + (len(contents[0].HTML)+len(contents[0].Subject)+len(contents[0].Badge))/n
			perEmail += small + len(*e.ContentID)
			oldBytes += small + len(oldBody)
			bodies += len(oldBody)
		}
		if track && !free {
			// Before, managed_messages.raw also kept the whole MIME; now it keeps none.
			oldBytes += mime
			t.Logf("template %d B, %d emails: per email before %d B (base64 body %d + MIME in managed_messages %d + subject/data), after %d B (subject/data + content id + 1/%d of the content row; %d B at 1 lakh emails)",
				len(template50KB), n, oldBytes/n, bodies/n, mime/n, newBytes/n, n, perEmail/n+len(contents[0].HTML)/100000)
		}
	})
}
