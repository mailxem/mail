// Package oauth lets MCP clients such as Claude, ChatGPT and Grok connect to a
// workspace by signing in, instead of pasting an API key. It implements the
// authorization code flow with PKCE (S256 only), dynamic client registration
// and rotating refresh tokens. Access tokens are short-lived assistant keys,
// bound to the user who approved, so the API's own permission checks apply.
package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	em "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/api/middleware"
	"kori/internal/assistant"
	"kori/internal/models"
)

const (
	codeTTL    = 10 * time.Minute
	accessTTL  = time.Hour
	refreshTTL = 30 * 24 * time.Hour
)

type Client struct {
	ID           string    `gorm:"primaryKey" json:"client_id"`
	Name         string    `json:"client_name"`
	RedirectURIs string    `json:"-"` // newline-separated
	CreatedAt    time.Time `json:"-"`
}

func (Client) TableName() string { return "oauth_clients" }

// Grant is one app a user connected. Its code is single use; its refresh
// token rotates on every use and only its hash is stored.
type Grant struct {
	ID            string     `gorm:"primaryKey;type:uuid" json:"id"`
	ClientID      string     `gorm:"index;not null" json:"-"`
	ClientName    string     `json:"clientName"`
	TeamID        string     `gorm:"index;type:uuid;not null" json:"-"`
	UserID        string     `gorm:"index;type:uuid;not null" json:"-"`
	RedirectURI   string     `json:"-"`
	Challenge     string     `json:"-"`
	CodeHash      string     `gorm:"index" json:"-"`
	CodeExpiresAt time.Time  `json:"-"`
	RefreshHash   string     `gorm:"index" json:"-"`
	RefreshUntil  time.Time  `json:"-"`
	AccessKeyID   string     `json:"-"`
	LastUsedAt    *time.Time `json:"lastUsedAt"`
	RevokedAt     *time.Time `json:"-"`
	CreatedAt     time.Time  `json:"createdAt"`
}

func (Grant) TableName() string { return "oauth_grants" }

func Migrate(db *gorm.DB) error { return db.AutoMigrate(&Client{}, &Grant{}) }

func secret() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
func hash(s string) string { sum := sha256.Sum256([]byte(s)); return hex.EncodeToString(sum[:]) }

// validRedirect accepts https URIs and http loopback ones (for local clients
// such as Claude Code), never fragments or credentials.
func validRedirect(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Fragment != "" || u.User != nil || u.Host == "" {
		return false
	}
	if u.Scheme == "https" {
		return true
	}
	h := u.Hostname()
	return u.Scheme == "http" && (h == "localhost" || h == "127.0.0.1" || h == "::1")
}

func (c Client) allows(redirect string) bool {
	return slices.Contains(strings.Split(c.RedirectURIs, "\n"), redirect)
}

// oauthError is RFC 6749's error body.
func oauthError(c echo.Context, status int, code, description string) error {
	c.Response().Header().Set("Cache-Control", "no-store")
	return c.JSON(status, map[string]string{"error": code, "error_description": description})
}

func Register(e *echo.Echo, db *gorm.DB, jwtSecret string) {
	limit := func(r float64, burst int) echo.MiddlewareFunc {
		return em.RateLimiterWithConfig(em.RateLimiterConfig{Store: em.NewRateLimiterMemoryStoreWithConfig(em.RateLimiterMemoryStoreConfig{Rate: rate.Limit(r), Burst: burst})})
	}
	session := middleware.NewAuthMiddleware(jwtSecret).SessionMiddleware()
	e.POST("/api/v1/oauth/register", register(db), em.BodyLimit("16K"), limit(1, 20))
	e.GET("/api/v1/oauth/clients/:id", client(db), limit(5, 30))
	e.POST("/api/v1/oauth/authorize", authorize(db), em.BodyLimit("8K"), session)
	e.POST("/api/v1/oauth/token", token(db), em.BodyLimit("8K"), limit(10, 50))
	e.GET("/api/v1/oauth/grants", grants(db), session)
	e.DELETE("/api/v1/oauth/grants/:id", revoke(db), session)
}

// register is RFC 7591 dynamic client registration for public clients.
func register(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var in struct {
			Name         string   `json:"client_name"`
			RedirectURIs []string `json:"redirect_uris"`
		}
		if json.NewDecoder(c.Request().Body).Decode(&in) != nil || len(in.RedirectURIs) == 0 || len(in.RedirectURIs) > 10 {
			return oauthError(c, 400, "invalid_client_metadata", "redirect_uris is required")
		}
		for _, r := range in.RedirectURIs {
			if len(r) > 2000 || !validRedirect(r) {
				return oauthError(c, 400, "invalid_redirect_uri", "Redirect URIs must use https, or http on localhost")
			}
		}
		name := strings.TrimSpace(in.Name)
		if name == "" {
			name = "An AI app"
		}
		if len([]rune(name)) > 80 {
			name = string([]rune(name)[:80])
		}
		cl := Client{ID: "xem_" + secret()[:24], Name: name, RedirectURIs: strings.Join(in.RedirectURIs, "\n")}
		if err := db.WithContext(c.Request().Context()).Create(&cl).Error; err != nil {
			return oauthError(c, 500, "server_error", "Registration is unavailable")
		}
		c.Response().Header().Set("Cache-Control", "no-store")
		return c.JSON(201, map[string]any{"client_id": cl.ID, "client_name": cl.Name, "redirect_uris": in.RedirectURIs, "token_endpoint_auth_method": "none", "grant_types": []string{"authorization_code", "refresh_token"}, "response_types": []string{"code"}})
	}
}

// client tells the consent page which app is asking, after checking its
// redirect URI, so the page never sends a code to an unregistered address.
func client(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var cl Client
		if db.WithContext(c.Request().Context()).First(&cl, "id = ?", c.Param("id")).Error != nil || !cl.allows(c.QueryParam("redirect_uri")) {
			return echo.NewHTTPError(404, "This app isn't registered, or its return address doesn't match")
		}
		u, _ := url.Parse(c.QueryParam("redirect_uri"))
		return c.JSON(200, map[string]string{"name": cl.Name, "host": u.Host})
	}
}

func authorize(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var in struct {
			ClientID    string `json:"client_id"`
			RedirectURI string `json:"redirect_uri"`
			Challenge   string `json:"code_challenge"`
			Method      string `json:"code_challenge_method"`
			State       string `json:"state"`
		}
		if c.Bind(&in) != nil || in.Method != "S256" || len(in.Challenge) < 43 || len(in.Challenge) > 128 || len(in.State) > 1000 {
			return echo.NewHTTPError(400, "The app sent an incomplete sign-in request (PKCE S256 is required)")
		}
		ctx := c.Request().Context()
		var cl Client
		if db.WithContext(ctx).First(&cl, "id = ?", in.ClientID).Error != nil || !cl.allows(in.RedirectURI) {
			return echo.NewHTTPError(400, "This app isn't registered, or its return address doesn't match")
		}
		var user models.User
		if db.WithContext(ctx).First(&user, "id = ? AND team_id = ? AND is_deleted = false", c.Get("userID"), c.Get("teamID")).Error != nil {
			return echo.NewHTTPError(401, "Sign in required")
		}
		code := secret()
		g := Grant{ID: uuid.NewString(), ClientID: cl.ID, ClientName: cl.Name, TeamID: user.TeamID, UserID: user.ID, RedirectURI: in.RedirectURI, Challenge: in.Challenge, CodeHash: hash(code), CodeExpiresAt: time.Now().UTC().Add(codeTTL)}
		if err := db.WithContext(ctx).Create(&g).Error; err != nil {
			return echo.NewHTTPError(503, "Sign-in is unavailable")
		}
		u, _ := url.Parse(in.RedirectURI)
		q := u.Query()
		q.Set("code", code)
		if in.State != "" {
			q.Set("state", in.State)
		}
		u.RawQuery = q.Encode()
		c.Response().Header().Set("Cache-Control", "no-store")
		return c.JSON(200, map[string]string{"redirect": u.String()})
	}
}

var errGrant = errors.New("invalid_grant")

func token(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		r := c.Request()
		if err := r.ParseForm(); err != nil {
			return oauthError(c, 400, "invalid_request", "Send a form-encoded body")
		}
		f := r.PostForm
		ctx := r.Context()
		var g Grant
		var refreshing bool
		err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			lock := tx.Clauses(clause.Locking{Strength: "UPDATE"})
			now := time.Now().UTC()
			switch f.Get("grant_type") {
			case "authorization_code":
				if f.Get("code") == "" || lock.First(&g, "code_hash = ? AND client_id = ?", hash(f.Get("code")), f.Get("client_id")).Error != nil {
					return errGrant
				}
				sum := sha256.Sum256([]byte(f.Get("code_verifier")))
				if g.RevokedAt != nil || now.After(g.CodeExpiresAt) || g.RedirectURI != f.Get("redirect_uri") || base64.RawURLEncoding.EncodeToString(sum[:]) != g.Challenge {
					return errGrant
				}
			case "refresh_token":
				refreshing = true
				if f.Get("refresh_token") == "" || lock.First(&g, "refresh_hash = ? AND client_id = ?", hash(f.Get("refresh_token")), f.Get("client_id")).Error != nil {
					return errGrant
				}
				if g.RevokedAt != nil || now.After(g.RefreshUntil) {
					return errGrant
				}
			default:
				return errors.New("unsupported_grant_type")
			}
			var user models.User
			if tx.First(&user, "id = ? AND team_id = ? AND is_deleted = false", g.UserID, g.TeamID).Error != nil {
				return errGrant
			}
			access, _, keyID, err := assistant.IssueKey(tx, user, g.ClientName+" (connected app)", accessTTL)
			if err != nil {
				return err
			}
			if g.AccessKeyID != "" {
				if err := tx.Model(&models.APIKey{}).Where("id = ?", g.AccessKeyID).Update("is_deleted", true).Error; err != nil {
					return err
				}
			}
			refresh := secret()
			// The code is spent: clearing its hash makes a replay fail.
			if err := tx.Model(&g).Updates(map[string]any{"code_hash": "", "refresh_hash": hash(refresh), "refresh_until": now.Add(refreshTTL), "access_key_id": keyID, "last_used_at": now}).Error; err != nil {
				return err
			}
			c.Set("oauth.result", map[string]any{"access_token": access, "token_type": "Bearer", "expires_in": int(accessTTL.Seconds()), "refresh_token": refresh})
			return nil
		})
		if err != nil {
			if err.Error() == "unsupported_grant_type" {
				return oauthError(c, 400, "unsupported_grant_type", "Use authorization_code or refresh_token")
			}
			if errors.Is(err, errGrant) {
				msg := "The code is invalid, expired or already used"
				if refreshing {
					msg = "The refresh token is invalid, expired or revoked"
				}
				return oauthError(c, 400, "invalid_grant", msg)
			}
			return oauthError(c, 500, "server_error", "Sign-in is unavailable")
		}
		c.Response().Header().Set("Cache-Control", "no-store")
		c.Response().Header().Set("Pragma", "no-cache")
		return c.JSON(http.StatusOK, c.Get("oauth.result"))
	}
}

// grants lists the apps the signed-in user connected.
func grants(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		list := []Grant{}
		if err := db.WithContext(c.Request().Context()).Where("user_id = ? AND team_id = ? AND revoked_at IS NULL AND refresh_hash <> '' AND refresh_until > ?", c.Get("userID"), c.Get("teamID"), time.Now().UTC()).Order("created_at DESC").Limit(50).Find(&list).Error; err != nil {
			return echo.NewHTTPError(503, "Connected apps are unavailable")
		}
		return c.JSON(200, list)
	}
}

// revoke disconnects an app at once: its refresh token stops working and
// its current access key is deleted.
func revoke(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := db.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
			var g Grant
			if err := tx.First(&g, "id = ? AND user_id = ? AND team_id = ?", c.Param("id"), c.Get("userID"), c.Get("teamID")).Error; err != nil {
				return err
			}
			if g.AccessKeyID != "" {
				if err := tx.Model(&models.APIKey{}).Where("id = ?", g.AccessKeyID).Update("is_deleted", true).Error; err != nil {
					return err
				}
			}
			return tx.Model(&g).Updates(map[string]any{"revoked_at": time.Now().UTC(), "refresh_hash": ""}).Error
		})
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(404, "App not found")
		}
		if err != nil {
			return echo.NewHTTPError(503, "Could not disconnect the app")
		}
		return c.NoContent(204)
	}
}
