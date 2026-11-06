package utils

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"kori/internal/affiliates"
	"kori/internal/config"
	"kori/internal/models"
	"kori/internal/utils/base64"
)

// legacyReplaceVariables is ReplaceVariables as it was before bulk content
// was shared, verbatim (map order).
func legacyReplaceVariables(input string, variables map[string]string, mailId string, cfg *config.Config, trackLinks bool, isMarketing bool) string {
	for variable, value := range variables {
		re := regexp.MustCompile(`{{\s*` + regexp.QuoteMeta(variable) + `(?:\.\w+)*\s*}}`)
		input = re.ReplaceAllStringFunc(input, func(string) string { return value })
	}

	if trackLinks || isMarketing {
		input = ReplaceLinksWithRedirect(input, mailId, cfg, isMarketing, trackLinks)
	}

	return base64.EncodeToBase64(input)
}

const campaignHTML = `<html><body><h1>Hi {{first_name}} {{ last_name }}</h1><p>{{name}} / {{full_name}} at {{company}}, {{city}} {{ custom.field }}</p>` +
	`<a href="https://example.com/a?x=1&amp;y=2">A</a> <a class='b' href='http://example.com/b'>B</a> <a href="mailto:x@example.com">M</a> ` +
	`<a href="https://example.com/unsubscribe">U</a> <a href="https://t.example.com/t/click/abc">T</a></body></html>`

const testBadge = `<div style="padding:20px 24px;text-align:center"><a href="https://xem.email/?utm_source=email-badge&amp;via=abc">Sent with <strong>Xem</strong></a></div>`

// What HandleCampaignProcess stored in Body for each contact, by tracking and
// badge, is what RenderEmail gives from the shared content.
func TestCampaignEmailsRenderWhatTheyUsedToStore(t *testing.T) {
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "campaign-secret"}, Server: config.ServerConfig{PublicURL: "https://api.example.com"}}
	big := strings.Repeat(`<p style="color:#333">Filler with a <a href="https://example.com/x">link</a> {{first_name}}</p>`, 600)
	for _, html := range []string{campaignHTML, strings.TrimSuffix(campaignHTML, "</body></html>"), big} {
		for _, badge := range []string{"", testBadge} {
			for _, tracked := range []bool{true, false} {
				// The variables HandleCampaignProcess builds and stores in Data.
				variables := map[string]string{"email": "r@example.com", "first_name": `<b>"Ravi"</b> & $1`, "last_name": "O'Neil", "name": "Ravi O'Neil", "full_name": "Ravi O'Neil", "company": "Acme & <Sons>", "city": "", "custom": "from campaign data"}
				data, _ := json.Marshal(variables)
				email := &models.Email{Base: models.Base{ID: uuid.NewString()}, ClickTrackingEnabled: &tracked, Data: data}
				got, err := RenderEmail(email, &models.EmailContent{Kind: ContentCampaign, HTML: html, Badge: badge}, cfg)
				if err != nil {
					t.Fatal(err)
				}
				want := legacyReplaceVariables(affiliates.WithBadge(html, badge), variables, email.ID, cfg, tracked, true)
				if base64.EncodeToBase64(got) != want {
					t.Fatalf("badge=%v tracked=%v: rendered campaign HTML differs from what was stored", badge != "", tracked)
				}
				if tracked != strings.Contains(got, "/t/open?token=") || !strings.Contains(got, "Unsubscribe from this list") || (badge != "") != strings.Contains(got, "Sent with") {
					t.Fatalf("badge=%v tracked=%v: wrong footer, badge or tracking", badge != "", tracked)
				}
			}
		}
	}
}

func TestRenderingIsStableWhenValuesHoldVariables(t *testing.T) {
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "s"}}
	data, _ := json.Marshal(map[string]string{"company": "{{email}}", "email": "{{first_name}}", "first_name": "x", "a": "{{company}}"})
	email := &models.Email{Base: models.Base{ID: uuid.NewString()}, Data: data}
	for _, kind := range []string{ContentCampaign, ContentNewsletter, ContentAutomation} {
		content := &models.EmailContent{Kind: kind, HTML: "{{a}} {{company}} {{email}} {{first_name}}"}
		first, err := RenderEmail(email, content, cfg)
		if err != nil {
			t.Fatal(err)
		}
		for range 50 {
			if again, _ := RenderEmail(email, content, cfg); again != first {
				t.Fatalf("%s rendered %q, then %q", kind, first, again)
			}
		}
	}
}

func TestStoredBodyWinsOverContent(t *testing.T) {
	id := "missing"
	email := &models.Email{Body: base64.EncodeToBase64("<p>stored</p>"), ContentID: &id}
	if got, err := EmailHTML(nil, email); err != nil || got != "<p>stored</p>" {
		t.Fatalf("got %q, %v", got, err)
	}
}
