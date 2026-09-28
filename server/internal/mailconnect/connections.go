package mailconnect

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/models"
	"kori/internal/utils/crypto"
)

const Google = "GOOGLE_OAUTH"
const Cloudflare = "CLOUDFLARE"
const MailScope = "https://mail.google.com/"

func GoogleConfig() (*oauth2.Config, error) {
	c := &oauth2.Config{ClientID: os.Getenv("GOOGLE_MAIL_CLIENT_ID"), ClientSecret: os.Getenv("GOOGLE_MAIL_CLIENT_SECRET"), RedirectURL: os.Getenv("GOOGLE_MAIL_REDIRECT_URI"), Endpoint: google.Endpoint, Scopes: []string{MailScope}}
	u, err := url.Parse(c.RedirectURL)
	if c.ClientID == "" || c.ClientSecret == "" || err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1"))) {
		return nil, errors.New("Google mailbox OAuth is not configured")
	}
	return c, nil
}

// No redirects: neither credentials nor mail content may leave provider hosts.
func HTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}
func OAuthContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, HTTPClient())
}
func StateHash(state string) string {
	hash := sha256.Sum256([]byte(state))
	return hex.EncodeToString(hash[:])
}
func aad(c *models.MailConnection) string {
	return "xem:mail:" + c.TeamID + ":" + c.ID + ":" + c.Provider
}
func Seal(c *models.MailConnection, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	c.Secret, err = crypto.SealSecret(raw, aad(c))
	return err
}
func Open(c *models.MailConnection, value any) error {
	raw, err := crypto.OpenSecret(c.Secret, aad(c))
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, value)
}

// Find includes disconnected rows, preventing fallback to generic SMTP/IMAP.
func Find(tx *gorm.DB, team, config string, imap bool) (*models.MailConnection, error) {
	column := "smtp_config_id"
	if imap {
		column = "imap_config_id"
	}
	var c models.MailConnection
	err := tx.Where("team_id = ? AND "+column+" = ?", team, config).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !c.Active || c.Secret == "" {
		return nil, errors.New("mailbox disconnected; connect it again in settings")
	}
	return &c, nil
}

// Refresh under a DB row lock so rotation persists across workers and cannot
// resurrect credentials removed by a concurrent disconnect.
func GoogleToken(ctx context.Context, tx *gorm.DB, connection *models.MailConnection) (string, error) {
	cfg, err := GoogleConfig()
	if err != nil {
		return "", err
	}
	var access string
	err = tx.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var c models.MailConnection
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND team_id = ? AND active = true", connection.ID, connection.TeamID).First(&c).Error; err != nil {
			return err
		}
		var token oauth2.Token
		if err := Open(&c, &token); err != nil {
			return err
		}
		fresh, err := cfg.TokenSource(OAuthContext(ctx), &token).Token()
		if err != nil {
			return errors.New("Google authorization expired or revoked; reconnect the mailbox")
		}
		if fresh.AccessToken != token.AccessToken || fresh.RefreshToken != token.RefreshToken {
			if fresh.RefreshToken == "" {
				fresh.RefreshToken = token.RefreshToken
			}
			if err := Seal(&c, fresh); err != nil {
				return err
			}
			if err := tx.Model(&c).Update("secret", c.Secret).Error; err != nil {
				return err
			}
		}
		access = fresh.AccessToken
		return nil
	})
	return access, err
}

// SASL XOAUTH2 is shared by IMAP and SMTP adapters.
type XOAUTH2 struct{ Username, Token string }

func (x XOAUTH2) Start() (string, []byte, error) {
	return "XOAUTH2", []byte("user=" + x.Username + "\x01auth=Bearer " + x.Token + "\x01\x01"), nil
}
func (x XOAUTH2) Next([]byte) ([]byte, error) {
	return nil, errors.New("Google mailbox authentication failed")
}
