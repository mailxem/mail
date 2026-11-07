package affiliates

import (
	"context"
	"html"
	"net/url"
	"os"
	"strings"
	"sync"

	"gorm.io/gorm"

	"kori/internal/billing"
	"kori/internal/models"
)

// teamCodes caches each workspace's referral code; codes never change.
var teamCodes sync.Map

// TeamCode returns the referral code for a workspace: its owner's affiliate
// code, created on first use so every Free sender earns from their badge.
func TeamCode(db *gorm.DB, teamID string) (string, error) {
	if code, ok := teamCodes.Load(teamID); ok {
		return code.(string), nil
	}
	var owner string
	if err := db.Model(&models.Team{}).Where("id = ?", teamID).Pluck("owner_user_id", &owner).Error; err != nil {
		return "", err
	}
	if owner == "" {
		// Older workspaces have no owner recorded; use their first administrator.
		if err := db.Model(&models.User{}).Where("team_id = ? AND role IN ?", teamID, []models.UserRole{models.UserRoleAdmin, models.UserRoleSuperAdmin}).
			Order("created_at").Limit(1).Pluck("id", &owner).Error; err != nil || owner == "" {
			return "", err
		}
	}
	a, err := Join(db, owner)
	if err != nil {
		return "", err
	}
	teamCodes.Store(teamID, a.Code)
	return a.Code, nil
}

func siteURL() string {
	if u := strings.TrimRight(os.Getenv("XEM_SITE_URL"), "/"); u != "" {
		return u
	}
	return "https://xem.email"
}

// Badge is the "Sent with Xem" footer for emails from Free workspaces, with
// the sender's referral link. Paid workspaces and licensed self-hosted installs
// get ""; unlicensed self-hosted installs get the badge without a referral. Like
// Mailchimp's badge, it turns every free newsletter into an invitation.
func Badge(ctx context.Context, db *gorm.DB, teamID string) string {
	if !billing.IsFree(ctx, teamID) {
		return ""
	}
	q := url.Values{"utm_source": {"email-badge"}}
	// Self-hosted installs have no Xem Cloud affiliate account to credit.
	if billing.Enabled {
		code, err := TeamCode(db, teamID)
		if err != nil || code == "" {
			return ""
		}
		q.Set("via", code)
	}
	link := siteURL() + "/?" + q.Encode()
	return `<div style="padding:20px 24px;text-align:center;font:12px/1.5 Arial,sans-serif;color:#71717a">` +
		`<a href="` + html.EscapeString(link) + `" style="color:#71717a;text-decoration:none">Sent with <strong style="color:#3f3f46">Xem</strong>, free email marketing</a></div>`
}

// WithBadge adds the badge to an email body: before </body> when there is one.
func WithBadge(body, badge string) string {
	if badge == "" {
		return body
	}
	if i := strings.LastIndex(strings.ToLower(body), "</body>"); i >= 0 {
		return body[:i] + badge + body[i:]
	}
	return body + badge
}
