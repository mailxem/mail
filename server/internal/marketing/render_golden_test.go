package marketing_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kori/internal/affiliates"
	"kori/internal/automation"
	"kori/internal/automation/processors"
	"kori/internal/billing"
	"kori/internal/config"
	"kori/internal/marketing"
	"kori/internal/models"
	"kori/internal/utils"
	cryptoutil "kori/internal/utils/crypto"
)

// Bulk emails used to store their HTML in Body; now they share one
// email_contents row and are rendered when sent or shown. These tests run the
// senders and compare what they render, byte for byte, with what the code
// before stored, copied below. Deploy: all API and worker processes together,
// and create the emails.content_id index CONCURRENTLY first.

// goldenHTML has links that are tracked, skipped (mailto, unsubscribe) and
// already tracked, variables with and without spaces, and a </body>.
const goldenHTML = `<html><body><h1>Hi {{first_name}} {{ last_name }}</h1><p>{{name}} at {{company}} ({{email}})</p>` +
	`<a href="https://example.com/a?x=1&amp;y=2">A</a> <a class='b' href='http://example.com/b'>B</a> <a href="mailto:x@example.com">M</a> ` +
	`<a href="https://example.com/unsubscribe">U</a> <a href="https://t.example.com/t/click/abc">T</a></body></html>`

// No value holds a whole {{variable}}: the old code filled those in random
// map order, so it had no single output to compare with (utils render tests).
var goldenContacts = []models.Contact{
	{Email: "plain@example.com", FirstName: "Asha", LastName: "Rao", Company: "Acme"},
	{Email: "tricky@example.com", FirstName: `<b>"Ravi"</b> & $1 {{`, LastName: "O'Neil", Company: "Acme & <Sons> }}"},
	{Email: "empty@example.com"},
}

// legacyReplaceVariables is utils.ReplaceVariables before content was shared, verbatim.
func legacyReplaceVariables(input string, variables map[string]string, mailId string, cfg *config.Config, trackLinks bool, isMarketing bool) string {
	for variable, value := range variables {
		re := regexp.MustCompile(`{{\s*` + regexp.QuoteMeta(variable) + `(?:\.\w+)*\s*}}`)
		input = re.ReplaceAllStringFunc(input, func(string) string { return value })
	}
	if trackLinks || isMarketing {
		input = utils.ReplaceLinksWithRedirect(input, mailId, cfg, isMarketing, trackLinks)
	}
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// goldenTeam seeds a workspace on the Free plan (badge on) or a paid one.
func goldenTeam(t *testing.T, db *gorm.DB, free bool) string {
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.Affiliate{}, &models.TrackingConsent{}))
	team, owner := uuid.NewString(), uuid.NewString()
	seed(t, db, &models.Team{Base: models.Base{ID: team}, Name: "Team", OwnerUserID: owner})
	seed(t, db, &models.User{Base: models.Base{ID: owner}, Email: owner + "@example.com", Password: "x", Role: models.UserRoleAdmin, TeamID: team})
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
	return team
}

func TestNewsletterEmailsRenderWhatTheyUsedToStore(t *testing.T) {
	for _, free := range []bool{false, true} {
		for _, track := range []bool{true, false} {
			t.Run(fmt.Sprintf("badge=%v/tracked=%v", free, track), func(t *testing.T) {
				db := testDB(t)
				secret := strings.Repeat("s", 32)
				t.Setenv("PUBLIC_API_URL", "https://api.example.com/")
				old := utils.RenderConfig
				t.Cleanup(func() { utils.RenderConfig = old })
				utils.RenderConfig = &config.Config{JWT: config.JWTConfig{Secret: secret}, Server: config.ServerConfig{PublicURL: "https://other.example.com"}}
				team := goldenTeam(t, db, free)
				list, sender, tpl, newsletter := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
				password, err := cryptoutil.Encrypt("fixture-password")
				require.NoError(t, err)
				seed(t, db, &models.SMTPConfig{Base: models.Base{ID: sender}, TeamID: team, FromEmail: "news@example.com", SupportsTLS: true, Password: password})
				seed(t, db, &models.Template{Base: models.Base{ID: tpl}, TeamID: team, CategoryID: uuid.NewString(), Name: "Edition", Subject: "Hello {{first_name}}"})
				for _, c := range goldenContacts {
					c.ID, c.TeamID, c.ListID, c.Status = uuid.NewString(), team, list, models.SubscriberStatusActive
					seed(t, db, &c)
				}
				edition := models.Campaign{Base: models.Base{ID: uuid.NewString()}, NewsletterID: &newsletter, TeamID: team, TemplateID: tpl, ListID: list, SMTPConfigID: sender, Status: models.CampaignStatusScheduled, HTMLBody: goldenHTML, Subject: "News for {{first_name}}", PostalAddress: "42 <Market> Street, Bengaluru", BatchSize: 100, TrackEngagement: &track}
				seed(t, db, &edition)
				s := marketing.New(db, secret, "https://api.example.com/")
				ctx := context.Background()
				require.NoError(t, s.PrepareEdition(ctx, edition.ID))
				badge := affiliates.Badge(ctx, db, team)
				require.Equal(t, free, badge != "")

				// PrepareEdition's body before content was shared, verbatim.
				personalize := func(body string, c models.Contact) string {
					for k, v := range map[string]string{"first_name": c.FirstName, "last_name": c.LastName, "name": strings.TrimSpace(c.FirstName + " " + c.LastName), "email": c.Email, "company": c.Company} {
						body = strings.ReplaceAll(body, "{{"+k+"}}", html.EscapeString(v))
					}
					return body
				}
				legacy := func(c models.Campaign, contact models.Contact, messageID, unsub string, tracking bool) string {
					body := personalize(c.HTMLBody, contact) + `<footer style="padding:24px;text-align:center;color:#71717a;font:12px Arial">` + html.EscapeString(c.PostalAddress) + `<br><a href="` + unsub + `">Unsubscribe</a></footer>` + badge
					if tracking {
						body = utils.TrackEmailHTML(body, messageID, s.PublicURL, s.Secret)
					}
					return base64.StdEncoding.EncodeToString([]byte(body))
				}

				var emails []models.Email
				require.NoError(t, db.Preload("Contact").Find(&emails).Error)
				require.Len(t, emails, len(goldenContacts))
				var contents int64
				require.NoError(t, db.Model(&models.EmailContent{}).Count(&contents).Error)
				require.EqualValues(t, 1, contents, "one content row per edition")
				for _, e := range emails {
					require.Empty(t, e.Body)
					require.NotNil(t, e.ContentID)
					require.Equal(t, track, *e.ClickTrackingEnabled)
					require.Equal(t, personalize(edition.Subject, *e.Contact), e.Subject)
					got, err := utils.EmailHTML(db, &e)
					require.NoError(t, err)
					require.Equal(t, legacy(edition, *e.Contact, e.ID, e.UnsubscribeURL, track), base64.StdEncoding.EncodeToString([]byte(got)), e.To)
				}
			})
		}
	}
}

func TestAutomationEmailsRenderWhatTheyUsedToStore(t *testing.T) {
	for _, category := range []string{"Transactional", "Marketing"} {
		for _, free := range []bool{false, true} {
			for _, track := range []bool{true, false} {
				t.Run(fmt.Sprintf("%s/badge=%v/tracked=%v", category, free, track), func(t *testing.T) {
					db := testDB(t)
					t.Setenv("PUBLIC_API_URL", "https://api.example.com")
					cfg := &config.Config{JWT: config.JWTConfig{Secret: strings.Repeat("x", 32)}, Server: config.ServerConfig{PublicURL: "https://track.example.com"}}
					old := utils.RenderConfig
					t.Cleanup(func() { utils.RenderConfig = old })
					utils.RenderConfig = cfg
					team := goldenTeam(t, db, free)
					sender, categoryID, tpl := uuid.NewString(), uuid.NewString(), uuid.NewString()
					password, err := cryptoutil.Encrypt("fixture")
					require.NoError(t, err)
					seed(t, db, &models.EmailCategory{Base: models.Base{ID: categoryID}, TeamID: team, Name: category})
					seed(t, db, &models.SMTPConfig{Base: models.Base{ID: sender}, TeamID: team, Password: password, FromEmail: "hello@example.com", SupportsTLS: true})
					seed(t, db, &models.Template{Base: models.Base{ID: tpl}, TeamID: team, CategoryID: categoryID, HTMLBody: goldenHTML + `<p>{{ plan }} {{city}}</p>`, Subject: "Hi {{first_name}}"})
					processor := processors.NewEmailProcessor(db, nil, cfg)
					nodeData, _ := json.Marshal(map[string]any{"templateId": tpl, "smtpConfigId": sender, "postalAddress": "9 <Lake> Road, Pune", "variables": map[string]string{"plan": "<Pro> & co"}, "trackEngagement": track})
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
					require.EqualValues(t, 1, contents, "one content row per step")
					var template models.Template
					require.NoError(t, db.Preload("Category").First(&template, "id = ?", tpl).Error)
					badge := affiliates.Badge(context.Background(), db, team)
					require.Equal(t, free, badge != "")
					for _, e := range emails {
						require.Empty(t, e.Body)
						require.Equal(t, track, *e.ClickTrackingEnabled)
						var variables map[string]string
						require.NoError(t, json.Unmarshal(e.Data, &variables))
						// What Process stored before content was shared, verbatim.
						htmlBody := template.HTMLBody
						bodyVariables := make(map[string]string, len(variables))
						for k, v := range variables {
							bodyVariables[k] = html.EscapeString(v)
						}
						isMarketing := template.Category.Name == "Marketing"
						if isMarketing {
							require.NotEmpty(t, e.UnsubscribeURL)
							htmlBody += `<footer style="padding:24px;text-align:center;font:12px Arial">` + html.EscapeString("9 <Lake> Road, Pune") + `<br><a href="` + e.UnsubscribeURL + `">Unsubscribe</a></footer>`
							htmlBody = affiliates.WithBadge(htmlBody, badge)
						}
						want := legacyReplaceVariables(htmlBody, bodyVariables, e.ID, cfg, track, false)
						got, err := utils.EmailHTML(db, &e)
						require.NoError(t, err)
						require.Equal(t, want, base64.StdEncoding.EncodeToString([]byte(got)), e.To)
						require.Equal(t, isMarketing && free, strings.Contains(got, "Sent with"))
					}
				})
			}
		}
	}
}
