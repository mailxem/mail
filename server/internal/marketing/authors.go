package marketing

import (
	"errors"
	"html"
	"regexp"
	"strings"

	"gorm.io/gorm"
	"kori/internal/models"
)

// CampaignAuthor is who signs a campaign: the author it picked, or else the
// workspace's default. nil means no byline.
func CampaignAuthor(db *gorm.DB, teamID string, authorID *string) (*models.Author, error) {
	var a models.Author
	if authorID != nil && *authorID != "" {
		err := Scope(db, teamID).First(&a, "id = ?", *authorID).Error
		if err == nil {
			return &a, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	err := Scope(db, teamID).Where("is_default = ?", true).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// ReplyTo is where readers' replies go: the campaign's own Reply-To, else its
// author's email. "" leaves the sender's own reply address.
func ReplyTo(campaignReplyTo string, a *models.Author) string {
	if r := strings.TrimSpace(campaignReplyTo); r != "" {
		return r
	}
	if a != nil {
		return a.Email
	}
	return ""
}

var bodyTag = regexp.MustCompile(`(?i)<body\b[^>]*>`)

// WithByline puts "By <name>", with the author's photo, at the top of an
// email body, when the author shows their byline in emails.
func WithByline(body string, a *models.Author) string {
	if a == nil || !a.BylineEmail || strings.TrimSpace(a.Name) == "" {
		return body
	}
	photo := ""
	if ValidSiteLogo(a.PhotoURL) && a.PhotoURL != "" {
		photo = `<td style="padding-right:10px"><img src="` + html.EscapeString(a.PhotoURL) + `" width="32" height="32" alt="" style="display:block;width:32px;height:32px;border-radius:50%;border:0"></td>`
	}
	byline := `<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0"><tr><td align="center" style="padding:16px 16px 0"><table role="presentation" cellpadding="0" cellspacing="0" border="0"><tr>` +
		photo + `<td style="font-family:Arial,Helvetica,sans-serif;font-size:14px;line-height:20px;color:#555555">By <strong style="color:#222222">` + html.EscapeString(a.Name) + `</strong></td></tr></table></td></tr></table>`
	if loc := bodyTag.FindStringIndex(body); loc != nil {
		return body[:loc[1]] + byline + body[loc[1]:]
	}
	return byline + body
}

// Byline is what a newsletter website shows under an issue's title; nil when
// the author keeps their byline off the web.
func Byline(a *models.Author) map[string]string {
	if a == nil || !a.BylineWeb || strings.TrimSpace(a.Name) == "" {
		return nil
	}
	return map[string]string{"name": a.Name, "photo": a.PhotoURL}
}
