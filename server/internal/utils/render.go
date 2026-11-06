package utils

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"maps"
	"os"
	"slices"
	"strings"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/affiliates"
	"kori/internal/config"
	"kori/internal/db"
	"kori/internal/models"
	"kori/internal/utils/base64"
)

// Kinds of shared email content, by sender. Change the version when a
// sender's output changes, so new sends never reuse rows rendered the old way.
//
// Deploy notes: every API and worker process must run a build that knows these
// kinds before any of them stores one, or an old binary sends an empty body.
// Create the emails.content_id index CONCURRENTLY before deploying
// (AutoMigrate would build idx_emails_content_id while blocking writes).
const (
	ContentCampaign   = "campaign/1"
	ContentNewsletter = "newsletter/1"
	ContentAutomation = "automation/1"
)

// RenderConfig is the config bulk emails are rendered with. main sets it to
// the one their senders get; without it the environment is read.
var RenderConfig *config.Config

func renderConfig() *config.Config {
	if RenderConfig != nil {
		return RenderConfig
	}
	return config.GetConfig()
}

// StoreEmailContent saves what a bulk send's emails share, once, and returns
// its ID: a hash of the team and everything rendering depends on.
func StoreEmailContent(tx *gorm.DB, c models.EmailContent) (string, error) {
	sum := sha256.Sum256([]byte(strings.Join([]string{c.TeamID, c.Kind, c.HTML, c.Subject, c.PostalAddress, c.Badge}, "\x00")))
	c.ID = hex.EncodeToString(sum[:])
	return c.ID, tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&c).Error
}

// Personalize fills {{key}} with the HTML-escaped value, the way newsletters
// do. Keys go in order so the result is the same every time.
func Personalize(body string, values map[string]string) string {
	for _, k := range slices.Sorted(maps.Keys(values)) {
		body = strings.ReplaceAll(body, "{{"+k+"}}", html.EscapeString(values[k]))
	}
	return body
}

// RenderEmail renders shared content for one email: the HTML its sender used
// to store in Body, from the email's ID, data, tracking flag and unsubscribe URL.
func RenderEmail(email *models.Email, content *models.EmailContent, cfg *config.Config) (string, error) {
	data := map[string]string{}
	if len(email.Data) > 0 {
		if err := json.Unmarshal(email.Data, &data); err != nil {
			return "", err
		}
	}
	tracked := email.ClickTrackingEnabled == nil || *email.ClickTrackingEnabled
	switch content.Kind {
	case ContentCampaign:
		// Campaigns go to a list, so they are always marketing.
		return fillVariables(affiliates.WithBadge(content.HTML, content.Badge), data, email.ID, cfg, tracked, true), nil
	case ContentNewsletter:
		body := Personalize(content.HTML, data) + `<footer style="padding:24px;text-align:center;color:#71717a;font:12px Arial">` + html.EscapeString(content.PostalAddress) + `<br><a href="` + email.UnsubscribeURL + `">Unsubscribe</a></footer>` + content.Badge
		if tracked {
			// marketing.Service tracks with PUBLIC_API_URL, not PUBLIC_URL.
			body = TrackEmailHTML(body, email.ID, strings.TrimRight(os.Getenv("PUBLIC_API_URL"), "/"), cfg.JWT.Secret)
		}
		return body, nil
	case ContentAutomation:
		body := content.HTML
		if content.PostalAddress != "" { // marketing templates only
			body += `<footer style="padding:24px;text-align:center;font:12px Arial">` + html.EscapeString(content.PostalAddress) + `<br><a href="` + email.UnsubscribeURL + `">Unsubscribe</a></footer>`
			body = affiliates.WithBadge(body, content.Badge)
		}
		for k, v := range data {
			data[k] = html.EscapeString(v)
		}
		return fillVariables(body, data, email.ID, cfg, tracked, false), nil
	}
	return "", fmt.Errorf("unknown email content kind %q", content.Kind)
}

// EmailHTML returns an email's HTML: its stored Body or, for a bulk send that
// stored none, its shared content rendered for it. A stored body always wins.
func EmailHTML(tx *gorm.DB, email *models.Email) (string, error) {
	if email.Body != "" || email.ContentID == nil {
		return base64.DecodeFromBase64(email.Body)
	}
	content, err := loadContent(tx, *email.ContentID, email.TeamID)
	if err != nil {
		return "", err
	}
	return RenderEmail(email, content, renderConfig())
}

// Content rows never change (the ID is their hash), so they are cached.
var contents = struct {
	sync.Mutex
	rows map[string]*models.EmailContent
}{rows: map[string]*models.EmailContent{}}

func loadContent(tx *gorm.DB, id, team string) (*models.EmailContent, error) {
	contents.Lock()
	c := contents.rows[id]
	contents.Unlock()
	if c != nil && c.TeamID == team {
		return c, nil
	}
	c = &models.EmailContent{}
	if err := tx.First(c, "id = ? AND team_id = ?", id, team).Error; err != nil {
		return nil, err
	}
	contents.Lock()
	if len(contents.rows) >= 64 {
		clear(contents.rows) // ponytail: crude bound; an LRU if many bulk sends run at once
	}
	contents.rows[id] = c
	contents.Unlock()
	return c, nil
}

// RenderRaw rebuilds the MIME message SendEmail handed to managed sending, so
// the managed outbox doesn't keep a copy (sending.RenderRaw).
func RenderRaw(ctx context.Context, emailID string) ([]byte, error) {
	var email models.Email
	if err := db.GetDB().WithContext(ctx).First(&email, "id = ?", emailID).Error; err != nil {
		return nil, err
	}
	m, _, err := buildMessage(&email)
	if err != nil {
		return nil, err
	}
	var raw bytes.Buffer
	_, err = m.WriteTo(&raw)
	return raw.Bytes(), err
}
