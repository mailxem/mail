package utils

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
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

	if trackLinks {
		input = ReplaceLinksWithRedirect(input, mailId, cfg, isMarketing)
	}

	return base64.EncodeToBase64(input)
}

const campaignHTML = `<html><body><h1>Hi {{first_name}} {{ last_name }}</h1><p>{{name}} / {{full_name}} at {{company}}, {{city}} {{ custom.field }}</p>` +
	`<a href="https://example.com/a?x=1&amp;y=2">A</a> <a class='b' href='http://example.com/b'>B</a> <a href="mailto:x@example.com">M</a> ` +
	`<a href="https://example.com/unsubscribe">U</a> <a href="https://t.example.com/t/click/abc">T</a></body></html>`

func TestCampaignEmailsRenderWhatTheyUsedToStore(t *testing.T) {
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "campaign-secret"}, Server: config.ServerConfig{PublicURL: "https://api.example.com"}}
	big := strings.Repeat(`<p style="color:#333">Filler with a <a href="https://example.com/x">link</a> {{first_name}}</p>`, 600)
	for _, html := range []string{campaignHTML, campaignHTML[:len(campaignHTML)-len("</body></html>")], big} {
		for _, marketing := range []bool{false, true} {
			// The variables HandleCampaignProcess builds and stores in Data.
			variables := map[string]string{"email": "r@example.com", "first_name": `<b>"Ravi"</b> & $1`, "last_name": "O'Neil", "name": "Ravi O'Neil", "full_name": "Ravi O'Neil", "company": "Acme & <Sons>", "city": "", "custom": "from campaign data"}
			data, _ := json.Marshal(variables)
			kind := ContentCampaign
			if marketing {
				kind = ContentCampaignMarketing
			}
			tracked := true
			email := &models.Email{Base: models.Base{ID: uuid.NewString()}, ClickTrackingEnabled: &tracked, Data: data}
			got, err := RenderEmail(email, &models.EmailContent{Kind: kind, HTML: html}, cfg)
			if err != nil {
				t.Fatal(err)
			}
			if want := legacyReplaceVariables(html, variables, email.ID, cfg, true, marketing); base64.EncodeToBase64(got) != want {
				t.Fatalf("marketing=%v: rendered campaign HTML differs from what was stored", marketing)
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
