package marketing_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"html"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"kori/internal/automation"
	"kori/internal/automation/processors"
	"kori/internal/config"
	"kori/internal/marketing"
	"kori/internal/models"
	"kori/internal/utils"
	cryptoutil "kori/internal/utils/crypto"
)

// goldenHTML has links that are tracked, skipped (mailto, unsubscribe) and
// already tracked, variables with and without spaces, and a </body>.
const goldenHTML = `<html><body><h1>Hi {{first_name}} {{ last_name }}</h1><p>{{name}} at {{company}} ({{email}})</p>` +
	`<a href="https://example.com/a?x=1&amp;y=2">A</a> <a class='b' href='http://example.com/b'>B</a> <a href="mailto:x@example.com">M</a> ` +
	`<a href="https://example.com/unsubscribe">U</a> <a href="https://t.example.com/t/click/abc">T</a></body></html>`

// No value holds a {{variable}}: the old code filled those in random map
// order, so it had no single output to compare with (see utils render tests).
var goldenContacts = []models.Contact{
	{Email: "plain@example.com", FirstName: "Asha", LastName: "Rao", Company: "Acme"},
	{Email: "tricky@example.com", FirstName: `<b>"Ravi"</b> & $1 {{`, LastName: "O'Neil", Company: "Acme & <Sons> }}"},
	{Email: "empty@example.com"},
}

// The newsletter body PrepareEdition stored before content was shared, verbatim.
func legacyNewsletterBody(s *marketing.Service, c models.Campaign, contact models.Contact, messageID, unsub string) string {
	personalize := func(body string, c models.Contact) string {
		for k, v := range map[string]string{"first_name": c.FirstName, "last_name": c.LastName, "name": strings.TrimSpace(c.FirstName + " " + c.LastName), "email": c.Email, "company": c.Company} {
			body = strings.ReplaceAll(body, "{{"+k+"}}", html.EscapeString(v))
		}
		return body
	}
	body := personalize(c.HTMLBody, contact) + `<footer style="padding:24px;text-align:center;color:#71717a;font:12px Arial">` + html.EscapeString(c.PostalAddress) + `<br><a href="` + unsub + `">Unsubscribe</a></footer>`
	body = utils.TrackEmailHTML(body, messageID, s.PublicURL, s.Secret)
	return base64.StdEncoding.EncodeToString([]byte(body))
}

func TestNewsletterEmailsRenderWhatTheyUsedToStore(t *testing.T) {
	db := testDB(t)
	require.NoError(t, db.AutoMigrate(&models.EmailContent{}))
	secret := strings.Repeat("s", 32)
	t.Setenv("PUBLIC_API_URL", "https://api.example.com/")
	old := utils.RenderConfig
	t.Cleanup(func() { utils.RenderConfig = old })
	utils.RenderConfig = &config.Config{JWT: config.JWTConfig{Secret: secret}, Server: config.ServerConfig{PublicURL: "https://other.example.com"}}
	team, list, sender, tpl, newsletter := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	password, err := cryptoutil.Encrypt("fixture-password")
	require.NoError(t, err)
	seed(t, db, &models.SMTPConfig{Base: models.Base{ID: sender}, TeamID: team, FromEmail: "news@example.com", SupportsTLS: true, Password: password})
	seed(t, db, &models.Template{Base: models.Base{ID: tpl}, TeamID: team, CategoryID: uuid.NewString(), Name: "Edition", Subject: "Hello {{first_name}}"})
	for _, c := range goldenContacts {
		c.ID, c.TeamID, c.ListID, c.Status = uuid.NewString(), team, list, models.SubscriberStatusActive
		seed(t, db, &c)
	}
	edition := models.Campaign{Base: models.Base{ID: uuid.NewString()}, NewsletterID: &newsletter, TeamID: team, TemplateID: tpl, ListID: list, SMTPConfigID: sender, Status: models.CampaignStatusScheduled, HTMLBody: goldenHTML, Subject: "News for {{first_name}}", PostalAddress: "42 <Market> Street, Bengaluru", BatchSize: 100}
	seed(t, db, &edition)
	s := marketing.New(db, secret, "https://api.example.com/")
	require.NoError(t, s.PrepareEdition(context.Background(), edition.ID))

	var emails []models.Email
	require.NoError(t, db.Preload("Contact").Find(&emails).Error)
	require.Len(t, emails, len(goldenContacts))
	var contents int64
	require.NoError(t, db.Model(&models.EmailContent{}).Count(&contents).Error)
	require.EqualValues(t, 1, contents)
	for _, e := range emails {
		require.Empty(t, e.Body)
		require.NotNil(t, e.ContentID)
		got, err := utils.EmailHTML(db, &e)
		require.NoError(t, err)
		require.Equal(t, legacyNewsletterBody(s, edition, *e.Contact, e.ID, e.UnsubscribeURL), base64.StdEncoding.EncodeToString([]byte(got)), e.To)
	}
}

func TestAutomationEmailsRenderWhatTheyUsedToStore(t *testing.T) {
	for _, category := range []string{"Transactional", "Marketing"} {
		t.Run(category, func(t *testing.T) {
			db := testDB(t)
			require.NoError(t, db.AutoMigrate(&models.EmailContent{}))
			t.Setenv("PUBLIC_API_URL", "https://api.example.com")
			cfg := &config.Config{JWT: config.JWTConfig{Secret: strings.Repeat("x", 32)}, Server: config.ServerConfig{PublicURL: "https://track.example.com"}}
			old := utils.RenderConfig
			t.Cleanup(func() { utils.RenderConfig = old })
			utils.RenderConfig = cfg
			team, sender, categoryID, tpl := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
			password, err := cryptoutil.Encrypt("fixture")
			require.NoError(t, err)
			seed(t, db, &models.EmailCategory{Base: models.Base{ID: categoryID}, TeamID: team, Name: category})
			seed(t, db, &models.SMTPConfig{Base: models.Base{ID: sender}, TeamID: team, Password: password, FromEmail: "hello@example.com", SupportsTLS: true})
			seed(t, db, &models.Template{Base: models.Base{ID: tpl}, TeamID: team, CategoryID: categoryID, HTMLBody: goldenHTML + `<p>{{ plan }} {{city}}</p>`, Subject: "Hi {{first_name}}"})
			processor := processors.NewEmailProcessor(db, nil, cfg)
			nodeData, _ := json.Marshal(map[string]any{"templateId": tpl, "smtpConfigId": sender, "postalAddress": "9 <Lake> Road, Pune", "variables": map[string]string{"plan": "<Pro> & co"}})
			node := &models.AutomationNode{Base: models.Base{ID: uuid.NewString()}, Type: models.NodeTypeEmail, Data: nodeData}
			for _, c := range goldenContacts {
				c.ID, c.TeamID, c.Status, c.City = uuid.NewString(), team, models.SubscriberStatusActive, "Pune & <Goa>"
				ctx := &automation.ExecutionContext{TeamID: team, ContactID: c.ID, Contact: &c, AutomationID: uuid.NewString(), Execution: &models.AutomationExecution{Base: models.Base{ID: uuid.NewString()}}, Variables: map[string]any{"city": "<Delhi>"}}
				_, err := processor.Process(ctx, node)
				require.NoError(t, err)
			}
			var emails []models.Email
			require.NoError(t, db.Find(&emails).Error)
			require.Len(t, emails, len(goldenContacts))
			var contents int64
			require.NoError(t, db.Model(&models.EmailContent{}).Count(&contents).Error)
			require.EqualValues(t, 1, contents)
			var template models.Template
			require.NoError(t, db.Preload("Category").First(&template, "id = ?", tpl).Error)
			for _, e := range emails {
				require.Empty(t, e.Body)
				var variables map[string]string
				require.NoError(t, json.Unmarshal(e.Data, &variables))
				// What Process stored before content was shared, verbatim.
				htmlBody := template.HTMLBody
				bodyVariables := make(map[string]string, len(variables))
				for k, v := range variables {
					bodyVariables[k] = html.EscapeString(v)
				}
				if template.Category.Name == "Marketing" {
					require.NotEmpty(t, e.UnsubscribeURL)
					htmlBody += `<footer style="padding:24px;text-align:center;font:12px Arial">` + html.EscapeString("9 <Lake> Road, Pune") + `<br><a href="` + e.UnsubscribeURL + `">Unsubscribe</a></footer>`
				}
				want := utils.ReplaceVariables(htmlBody, bodyVariables, e.ID, cfg, true, false)
				got, err := utils.EmailHTML(db, &e)
				require.NoError(t, err)
				require.Equal(t, want, base64.StdEncoding.EncodeToString([]byte(got)), e.To)
			}
		})
	}
}
