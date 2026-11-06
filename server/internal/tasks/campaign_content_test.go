package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
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
// what is sent and shown is what the old per-email body held.
func TestCampaignStoresContentOnceAndRendersTheOldBody(t *testing.T) {
	pageDatabases(t, func(t *testing.T, database *gorm.DB) {
		if err := database.AutoMigrate(&models.Campaign{}, &models.Template{}, &models.MailingList{}, &models.EmailContent{}, &models.File{}); err != nil {
			t.Fatal(err)
		}
		cfg := &config.Config{JWT: config.JWTConfig{Secret: "campaign-secret"}, Server: config.ServerConfig{PublicURL: "https://api.example.com"}}
		oldDB, oldCfg, oldBatch := db.DB, utils.RenderConfig, utils.ManagedDeliveryBatch
		t.Cleanup(func() { db.DB, utils.RenderConfig, utils.ManagedDeliveryBatch = oldDB, oldCfg, oldBatch })
		db.DB, utils.RenderConfig = database, cfg
		var submitted [][]byte
		utils.ManagedDeliveryBatch = func(emails []*models.Email, raws [][]byte) []error {
			submitted = append(submitted, raws...)
			return make([]error, len(emails))
		}

		const n = 100
		team, list := uuid.NewString(), uuid.NewString()
		quiet := database.Session(&gorm.Session{SkipHooks: true})
		must := func(err error) {
			t.Helper()
			if err != nil {
				t.Fatal(err)
			}
		}
		sender := models.SMTPConfig{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Provider: "MANAGED", Host: "managed", Port: 587, FromEmail: "news@example.com", IsActive: true, MaxSendRate: 50}
		must(quiet.Create(&sender).Error)
		must(quiet.Create(&models.MailingList{Base: models.Base{ID: list}, TeamID: team, Name: "Readers"}).Error)
		tpl := models.Template{Base: models.Base{ID: uuid.NewString()}, TeamID: team, CategoryID: uuid.NewString(), Name: "Weekly", Subject: "News for {{first_name}}", HTMLBody: template50KB}
		must(quiet.Create(&tpl).Error)
		for i := range n {
			must(quiet.Create(&models.Contact{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ListID: list, Email: fmt.Sprintf("reader%d@example.com", i), FirstName: fmt.Sprintf("Reader <%d>", i), Company: "Acme & Co", Status: models.SubscriberStatusActive}).Error)
		}
		campaign := models.Campaign{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Weekly", ListID: list, TemplateID: tpl.ID, SMTPConfigID: sender.ID, Status: models.CampaignStatusScheduled, BatchSize: 50, Data: []byte(`{"company":"Acme & Co"}`)}
		must(quiet.Create(&campaign).Error)

		payload, _ := json.Marshal(CampaignTask{CampaignID: campaign.ID, BatchSize: 50})
		must(NewTaskHandler(database, cfg).HandleCampaignProcess(context.Background(), asynq.NewTask(TaskTypeCampaignProcess, payload)))

		var contents []models.EmailContent
		must(database.Find(&contents).Error)
		if len(contents) != 1 || contents[0].HTML != template50KB || contents[0].Kind != utils.ContentCampaign {
			t.Fatalf("want one campaign content row holding the template, got %d", len(contents))
		}
		var emails []models.Email
		must(database.Where("campaign_id = ?", campaign.ID).Find(&emails).Error)
		if len(emails) != n || len(submitted) != n {
			t.Fatalf("%d emails, %d submitted; want %d", len(emails), len(submitted), n)
		}
		var newBytes, oldBytes, bodies, mime int
		for _, raw := range submitted {
			if !bytes.Contains(raw, []byte("Message-ID: <")) {
				t.Fatal("submitted MIME is missing its headers")
			}
			mime += len(raw)
		}
		for _, e := range emails {
			if e.Body != "" || e.ContentID == nil || *e.ContentID != contents[0].ID {
				t.Fatalf("email %s stored a body or no content", e.ID)
			}
			var variables map[string]string
			must(json.Unmarshal(e.Data, &variables))
			// The body HandleCampaignProcess used to store, same call.
			oldBody := utils.ReplaceVariables(tpl.HTMLBody, variables, e.ID, cfg, true, false)
			html, err := utils.EmailHTML(database, &e)
			must(err)
			if base64.EncodeToBase64(html) != oldBody {
				t.Fatalf("email %s renders differently from the body it used to store", e.ID)
			}
			// Variable-size bytes this email writes: subject, data, and before,
			// the base64 body; now the content ID and a share of the content row.
			small := len(e.Subject) + len(e.Data)
			newBytes += small + len(*e.ContentID) + len(contents[0].HTML)/n
			oldBytes += small + len(oldBody)
			bodies += len(oldBody)
		}
		// Before, managed_messages.raw also kept the whole MIME; now it keeps none.
		oldBytes += mime
		t.Logf("template %d B: per email before %d B (base64 body %d + MIME in managed_messages %d + subject/data), after %d B (subject/data + content id + 1/%d of the content row)",
			len(template50KB), oldBytes/n, bodies/n, mime/n, newBytes/n, n)
	})
}
