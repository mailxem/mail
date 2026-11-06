package sending

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/models"
	"kori/internal/utils/crypto"
)

// Gmail never reports single complaints to SES. Postmaster Tools publishes a
// daily spam rate for the shared DKIM domain and for each Feedback-ID field.
const (
	postmasterScope = "https://www.googleapis.com/auth/postmaster.readonly"
	postmasterAAD   = "xem:postmaster:google"
	// Gmail's published limits: keep under 0.1%, never reach 0.3%.
	gmailSuspendRate = 0.003
	gmailWarnRate    = 0.0008
	gmailDomainRate  = 0.001
)

// Swapped by tests for a fake Postmaster API.
var (
	postmasterAPI      = "https://gmailpostmastertools.googleapis.com/v1"
	postmasterEndpoint = google.Endpoint
)

// Postmaster is the single Google connection for the installation.
type Postmaster struct {
	ID           string `gorm:"primaryKey"`
	Secret       string
	StateHash    string
	Verifier     string
	StateExpires *time.Time
	LastDay      string // newest yyyymmdd already acted on
	ConnectedAt  *time.Time
	UpdatedAt    time.Time
}

func (Postmaster) TableName() string { return "managed_postmaster" }

func postmasterConfig() (*oauth2.Config, string, bool) {
	c := &oauth2.Config{ClientID: os.Getenv("GOOGLE_POSTMASTER_CLIENT_ID"), ClientSecret: os.Getenv("GOOGLE_POSTMASTER_CLIENT_SECRET"), RedirectURL: os.Getenv("GOOGLE_POSTMASTER_REDIRECT_URL"), Endpoint: postmasterEndpoint, Scopes: []string{postmasterScope}}
	domain := strings.ToLower(strings.TrimSpace(os.Getenv("GOOGLE_POSTMASTER_DOMAIN")))
	return c, domain, c.ClientID != "" && c.ClientSecret != "" && c.RedirectURL != "" && validDomain(domain)
}

type trafficStats struct {
	Name                  string  `json:"name"`
	UserReportedSpamRatio float64 `json:"userReportedSpamRatio"`
	SpammyFeedbackLoops   []struct {
		ID        string  `json:"id"`
		SpamRatio float64 `json:"spamRatio"`
	} `json:"spammyFeedbackLoops"`
}

// SyncPostmaster acts on the newest day Google has published, once per day.
// Without configuration or a connection it does nothing.
func (s *Service) SyncPostmaster(ctx context.Context) error {
	cfg, domain, ok := postmasterConfig()
	if !ok {
		return nil
	}
	var p Postmaster
	if err := s.DB.WithContext(ctx).Limit(1).Find(&p, "id = ?", "google").Error; err != nil || p.Secret == "" {
		return err
	}
	raw, err := crypto.OpenSecret(p.Secret, postmasterAAD)
	if err != nil {
		return err
	}
	var token oauth2.Token
	if err = json.Unmarshal(raw, &token); err != nil {
		return err
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Timeout: 30 * time.Second})
	end := s.Now()
	start := end.AddDate(0, 0, -7)
	q := url.Values{}
	for prefix, d := range map[string]time.Time{"startDate": start, "endDate": end} {
		q.Set(prefix+".year", fmt.Sprint(d.Year()))
		q.Set(prefix+".month", fmt.Sprint(int(d.Month())))
		q.Set(prefix+".day", fmt.Sprint(d.Day()))
	}
	res, err := cfg.Client(ctx, &token).Get(postmasterAPI + "/domains/" + url.PathEscape(domain) + "/trafficStats?" + q.Encode())
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("postmaster tools returned %d", res.StatusCode)
	}
	var out struct {
		TrafficStats []trafficStats `json:"trafficStats"`
	}
	if err = json.NewDecoder(res.Body).Decode(&out); err != nil {
		return err
	}
	if len(out.TrafficStats) == 0 {
		return nil
	}
	day := func(t trafficStats) string { return t.Name[strings.LastIndex(t.Name, "/")+1:] }
	sort.Slice(out.TrafficStats, func(i, j int) bool { return day(out.TrafficStats[i]) > day(out.TrafficStats[j]) })
	return s.applyPostmaster(ctx, day(out.TrafficStats[0]), out.TrafficStats[0])
}

func (s *Service) applyPostmaster(ctx context.Context, day string, stats trafficStats) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var p Postmaster
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&p, "id = ?", "google").Error; err != nil {
			return err
		}
		if len(day) != 8 || day <= p.LastDay {
			return nil
		}
		// Each Feedback-ID field is reported on its own: map team, campaign
		// and newsletter IDs to their workspace and keep the worst rate.
		rates := map[string]float64{}
		for _, loop := range stats.SpammyFeedbackLoops {
			if uuid.Validate(loop.ID) != nil {
				continue
			}
			var team string
			for _, table := range []string{"managed_accounts", "campaigns", "newsletters"} {
				column := "id"
				if table == "managed_accounts" {
					column = "team_id"
				}
				if !tx.Migrator().HasTable(table) {
					continue
				}
				var teams []string
				if err := tx.Table(table).Where(column+" = ?", loop.ID).Limit(1).Pluck("team_id", &teams).Error; err != nil {
					return err
				}
				if len(teams) > 0 {
					team = teams[0]
					break
				}
			}
			if team != "" && loop.SpamRatio > rates[team] {
				rates[team] = loop.SpamRatio
			}
		}
		for team, rate := range rates {
			var a Account
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Limit(1).Find(&a, "team_id = ?", team).Error; err != nil {
				return err
			}
			if a.TeamID == "" {
				continue
			}
			reason := fmt.Sprintf("Gmail users marked %.2f%% of your mail as spam on %s", rate*100, day)
			if rate >= gmailSuspendRate && !a.Suspended {
				if err := tx.Model(&a).Update("suspended", true).Error; err != nil {
					return err
				}
				if s.NotifyAlerts {
					if err := models.QueueSendingNotice(tx, team, "gmail-postmaster", day, "SUSPENDED", "Suspended because "+reason, s.Now()); err != nil {
						return err
					}
				}
				if err := tx.Create(&Audit{ID: uuid.NewString(), TeamID: team, Actor: "system:gmail-postmaster", Action: "auto_suspend_gmail", CreatedAt: s.Now()}).Error; err != nil {
					return err
				}
			} else if rate >= gmailWarnRate && s.NotifyAlerts {
				if err := models.QueueAlert(tx, team, "gmail-postmaster", day, "gmail_spam_rate", reason, s.Now()); err != nil {
					return err
				}
			}
		}
		if stats.UserReportedSpamRatio >= gmailDomainRate && s.NotifyAlerts {
			var teams []string
			if err := tx.Model(&models.User{}).Where("role = ? AND is_deleted = false", models.UserRoleSuperAdmin).Distinct().Pluck("team_id", &teams).Error; err != nil {
				return err
			}
			reason := fmt.Sprintf("Gmail users marked %.2f%% of all managed mail as spam on %s", stats.UserReportedSpamRatio*100, day)
			for _, team := range teams {
				if err := models.QueueAlert(tx, team, "gmail-postmaster-domain", day, "gmail_domain_spam_rate", reason, s.Now()); err != nil {
					return err
				}
			}
		}
		return tx.Model(&p).Update("last_day", day).Error
	})
}

func stateHash(state string) string {
	h := sha256.Sum256([]byte(state))
	return hex.EncodeToString(h[:])
}

func superadmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		api, _ := c.Get("isAPIKey").(bool)
		role, _ := c.Get("role").(string)
		if api || role != string(models.UserRoleSuperAdmin) {
			return echo.NewHTTPError(403, "Only a superadmin can connect Google Postmaster Tools")
		}
		return next(c)
	}
}

// postmasterConnect returns Google's consent URL for a superadmin to open.
func (s *Service) postmasterConnect(c echo.Context) error {
	cfg, _, ok := postmasterConfig()
	if !ok {
		return echo.NewHTTPError(404, "Google Postmaster Tools is not configured")
	}
	state, verifier, expires := oauth2.GenerateVerifier(), oauth2.GenerateVerifier(), s.Now().Add(10*time.Minute)
	p := Postmaster{ID: "google", StateHash: stateHash(state), Verifier: verifier, StateExpires: &expires}
	if err := s.DB.WithContext(c.Request().Context()).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"state_hash", "verifier", "state_expires", "updated_at"})}).Create(&p).Error; err != nil {
		return httpError(err)
	}
	return c.JSON(200, map[string]string{"url": cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"), oauth2.S256ChallengeOption(verifier))})
}

// postmasterCallback is Google's redirect target; the single-use state from
// postmasterConnect stands in for the session.
func (s *Service) postmasterCallback(c echo.Context) error {
	cfg, _, ok := postmasterConfig()
	if !ok {
		return echo.NewHTTPError(404, "Google Postmaster Tools is not configured")
	}
	ctx := c.Request().Context()
	var verifier string
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var p Postmaster
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&p, "id = ? AND state_hash = ? AND state_expires > ?", "google", stateHash(c.QueryParam("state")), s.Now()).Error; err != nil {
			return err
		}
		verifier = p.Verifier
		return tx.Model(&p).Updates(map[string]any{"state_hash": "", "verifier": ""}).Error
	})
	if err != nil || c.QueryParam("state") == "" || c.QueryParam("code") == "" {
		return echo.NewHTTPError(400, "This Google authorization link is invalid or expired; start again")
	}
	token, err := cfg.Exchange(context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Timeout: 30 * time.Second}), c.QueryParam("code"), oauth2.VerifierOption(verifier))
	if err != nil || token.RefreshToken == "" {
		return echo.NewHTTPError(400, "Google did not grant offline access; start again")
	}
	raw, _ := json.Marshal(token)
	sealed, err := crypto.SealSecret(raw, postmasterAAD)
	if err != nil {
		return httpError(err)
	}
	now := s.Now()
	if err = s.DB.WithContext(ctx).Model(&Postmaster{ID: "google"}).Updates(map[string]any{"secret": sealed, "connected_at": now}).Error; err != nil {
		return httpError(err)
	}
	return c.String(200, "Google Postmaster Tools is connected. You can close this tab.")
}
