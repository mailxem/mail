package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"html"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/hygiene"
	"kori/internal/marketing"
	"kori/internal/models"
	"kori/internal/services"
)

var (
	// queueSiteEmail sends a confirmation through the newsletter's own sender.
	queueSiteEmail = services.QueueAPIEmailRequest
	// checkSiteEmail looks up the address's domain, like a form signup.
	checkSiteEmail = func(ctx context.Context, email string) string { return (&hygiene.Checker{}).Reason(ctx, email) }
)

const (
	// signupResend stops a form from mailing one address over and over.
	signupResend = 10 * time.Minute
	// signupTTL is how long a confirmation link works.
	signupTTL = 7 * 24 * time.Hour
	// signupHourlyCap bounds confirmation mail per newsletter.
	// ponytail: per-newsletter cap, since readers all arrive through the site
	// Worker's IPs; per-reader limits need a trusted client IP from the Worker.
	signupHourlyCap = 300
)

// SiteURL is a newsletter website's address: xem.email/{slug} by default.
func SiteURL(slug string) string {
	base := strings.TrimRight(os.Getenv("XEM_SITE_URL"), "/")
	if base == "" {
		base = "https://xem.email"
	}
	return base + "/" + slug
}

func signupHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// The answer is the same whether or not a confirmation went out, so the form
// never reveals who already subscribes.
var checkInbox = map[string]string{"status": "check_inbox"}

// SubscribeNewsletter emails a confirmation link to a reader who signed up on
// the newsletter's website.
func (h *MarketingHandler) SubscribeNewsletter(c echo.Context) error {
	n, err := h.publicNewsletter(c)
	if err != nil {
		return err
	}
	var in struct{ Email string }
	if err := c.Bind(&in); err != nil {
		return bad("Enter a valid email address")
	}
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if !marketing.ValidEmail(email) {
		return bad("Enter a valid email address")
	}
	reason := checkSiteEmail(c.Request().Context(), email)
	if domain, typo := strings.CutPrefix(reason, hygiene.Typo("")); typo {
		return bad("Did you mean " + email[:strings.LastIndexByte(email, '@')+1] + domain + "?")
	} else if reason == hygiene.Disposable {
		return bad("Use an email address you'll keep; disposable addresses can't sign up")
	} else if reason != "" {
		return bad("Enter an email address that can receive mail")
	}
	db := h.service.DB.WithContext(c.Request().Context())
	now := time.Now().UTC()
	var active int64
	if err := db.Model(&models.Contact{}).Where("team_id = ? AND list_id = ? AND email = ? AND status = ? AND is_deleted = ?", n.TeamID, n.ListID, email, models.SubscriberStatusActive, false).Count(&active).Error; err != nil {
		return err
	}
	var recent, hourly int64
	if err := db.Model(&models.NewsletterSignup{}).Where("newsletter_id = ? AND email = ? AND created_at > ?", n.ID, email, now.Add(-signupResend)).Count(&recent).Error; err != nil {
		return err
	}
	if err := db.Model(&models.NewsletterSignup{}).Where("newsletter_id = ? AND created_at > ?", n.ID, now.Add(-time.Hour)).Count(&hourly).Error; err != nil {
		return err
	}
	if hourly >= signupHourlyCap {
		return echo.NewHTTPError(429, "Too many signups right now; try again in a little while")
	}
	if active > 0 || recent > 0 {
		return c.JSON(202, checkInbox)
	}
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return err
	}
	token := hex.EncodeToString(raw)
	signup := models.NewsletterSignup{ID: uuid.NewString(), TeamID: n.TeamID, NewsletterID: n.ID, Email: email, TokenHash: signupHash(token), CreatedAt: now}
	if err := db.Create(&signup).Error; err != nil {
		return err
	}
	link := SiteURL(*n.PublicSlug) + "/confirm?token=" + url.QueryEscape(token)
	name := html.EscapeString(n.Name)
	body := `<div style="font:16px/1.6 Arial,sans-serif;color:#27272a;max-width:520px;margin:0 auto;padding:32px 24px">` +
		`<p>Confirm that you'd like to receive <strong>` + name + `</strong>.</p>` +
		`<p style="margin:28px 0"><a href="` + html.EscapeString(link) + `" style="background:#18181b;color:#fff;padding:12px 20px;border-radius:8px;text-decoration:none">Yes, subscribe me</a></p>` +
		`<p style="color:#71717a;font-size:14px">If you didn't ask for this, ignore this email and you won't hear from us again.</p></div>`
	msg := models.Email{TeamID: n.TeamID, To: email, SMTPConfigID: n.SMTPConfigID, Subject: "Confirm your subscription to " + n.Name, Body: base64.StdEncoding.EncodeToString([]byte(body))}
	if err := queueSiteEmail(&msg, "site-signup:"+signup.ID, ""); err != nil {
		return err
	}
	return c.JSON(202, checkInbox)
}

// ConfirmNewsletter adds a reader who followed their confirmation link.
func (h *MarketingHandler) ConfirmNewsletter(c echo.Context) error {
	n, err := h.publicNewsletter(c)
	if err != nil {
		return err
	}
	var in struct{ Token string }
	if err := c.Bind(&in); err != nil || in.Token == "" {
		return bad("This confirmation link is incomplete")
	}
	now := time.Now().UTC()
	err = h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		var s models.NewsletterSignup
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("newsletter_id = ? AND token_hash = ?", n.ID, signupHash(in.Token)).First(&s).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return echo.NewHTTPError(410, "This confirmation link has expired; subscribe again")
			}
			return err
		}
		if s.ConfirmedAt != nil {
			return nil
		}
		if now.Sub(s.CreatedAt) > signupTTL {
			return echo.NewHTTPError(410, "This confirmation link has expired; subscribe again")
		}
		var suppressed int64
		if err := tx.Model(&models.SuppressionList{}).Where("team_id = ? AND email_address = ? AND is_active = ? AND is_deleted = ? AND (expires_at IS NULL OR expires_at > ?)", n.TeamID, s.Email, true, false, now).Count(&suppressed).Error; err != nil {
			return err
		}
		blocked, err := hygiene.Blocked(tx, s.Email)
		if err != nil {
			return err
		}
		if suppressed == 0 && !blocked[s.Email] {
			if err := subscribeConfirmed(tx, n, s.Email, now); err != nil {
				return err
			}
		}
		// Suppressed addresses are never added, and the page doesn't say so.
		return tx.Model(&s).Update("confirmed_at", now).Error
	})
	if err != nil {
		return err
	}
	return c.JSON(200, map[string]string{"status": "subscribed"})
}

// subscribeConfirmed adds or re-activates a confirmed reader on the
// newsletter's list. Bounced and complaining contacts stay as they are.
func subscribeConfirmed(tx *gorm.DB, n models.Newsletter, email string, now time.Time) error {
	var list models.MailingList
	if err := marketing.Scope(tx, n.TeamID).Clauses(clause.Locking{Strength: "UPDATE"}).First(&list, "id = ?", n.ListID).Error; err != nil {
		return err
	}
	var contact models.Contact
	err := marketing.Scope(tx, n.TeamID).Where("email = ? AND list_id = ?", email, list.ID).First(&contact).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		contact = models.Contact{Base: models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now}, TeamID: n.TeamID, ListID: list.ID, Email: email, Status: models.SubscriberStatusActive, LifecycleStage: "LEAD", Metadata: datatypes.JSON(`{"source":"website","consent":true,"doubleOptIn":true}`)}
		if err := tx.Session(&gorm.Session{SkipHooks: true}).Create(&contact).Error; err != nil {
			return err
		}
	case err != nil:
		return err
	case contact.Status == models.SubscriberStatusInactive:
		if err := marketing.KeepSending(tx, &contact, now); err != nil {
			return err
		}
	case contact.Status == models.SubscriberStatusUnsubscribed:
		// Confirming from their own inbox is a fresh, explicit yes.
		if err := tx.Model(&contact).UpdateColumn("status", models.SubscriberStatusActive).Error; err != nil {
			return err
		}
	default:
		return nil
	}
	return models.SyncSubscribersCountByID(tx, list.ID)
}
