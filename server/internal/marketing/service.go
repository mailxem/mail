package marketing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"net/mail"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/models"
	"kori/internal/utils"
)

type Service struct {
	DB                *gorm.DB
	Secret, PublicURL string
}

func New(db *gorm.DB, secret, publicURL string) *Service {
	return &Service{db, secret, strings.TrimRight(publicURL, "/")}
}
func Scope(db *gorm.DB, team string) *gorm.DB {
	return db.Where("team_id = ? AND is_deleted = ?", team, false)
}
func ValidEmail(value string) bool {
	a, e := mail.ParseAddress(value)
	return e == nil && a.Address == value && !strings.ContainsAny(value, "\r\n")
}
func (s *Service) Owns(ctx context.Context, model interface{}, id, team string) bool {
	if uuid.Validate(id) != nil || team == "" {
		return false
	}
	var count int64
	return Scope(s.DB.WithContext(ctx).Model(model), team).Where("id = ?", id).Count(&count).Error == nil && count == 1
}
func NextEdition(at time.Time, cadence, zone string, anchorDay ...int) (*time.Time, error) {
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return nil, fmt.Errorf("choose a valid IANA timezone")
	}
	t := at.In(loc)
	switch cadence {
	case "ONCE":
		return nil, nil
	case "DAILY":
		t = t.AddDate(0, 0, 1)
	case "WEEKLY":
		t = t.AddDate(0, 0, 7)
	case "MONTHLY":
		// Clamp end-of-month instead of skipping February.
		first := time.Date(t.Year(), t.Month()+1, 1, t.Hour(), t.Minute(), t.Second(), 0, loc)
		last := first.AddDate(0, 1, -1).Day()
		day := t.Day()
		if len(anchorDay) > 0 && anchorDay[0] >= 1 && anchorDay[0] <= 31 {
			day = anchorDay[0]
		}
		if day > last {
			day = last
		}
		t = first.AddDate(0, 0, day-1)
	default:
		return nil, fmt.Errorf("invalid cadence")
	}
	u := t.UTC()
	return &u, nil
}
func (s *Service) ValidateNewsletter(ctx context.Context, n *models.Newsletter) error {
	n.Name = strings.TrimSpace(n.Name)
	n.Subject = strings.TrimSpace(n.Subject)
	if len(n.Name) < 2 || len(n.Name) > 120 || len(n.Subject) < 1 || len(n.Subject) > 200 || strings.ContainsAny(n.Subject, "\r\n") {
		return fmt.Errorf("a name and a subject of up to 200 characters are required")
	}
	if !s.Owns(ctx, &models.Template{}, n.TemplateID, n.TeamID) || !s.Owns(ctx, &models.MailingList{}, n.ListID, n.TeamID) || !s.Owns(ctx, &models.SMTPConfig{}, n.SMTPConfigID, n.TeamID) {
		return fmt.Errorf("template, audience and sender must belong to your workspace")
	}
	if _, err := NextEdition(time.Now(), n.Cadence, n.Timezone); err != nil {
		return err
	}
	if n.Status != "DRAFT" && n.Status != "SCHEDULED" && n.Status != "PAUSED" {
		return fmt.Errorf("invalid newsletter status")
	}
	if n.Status == "SCHEDULED" {
		if n.NextSendAt == nil || !n.NextSendAt.After(time.Now()) {
			return fmt.Errorf("choose a future send time")
		}
		if len(strings.TrimSpace(n.PostalAddress)) < 8 || len(n.PostalAddress) > 500 {
			return fmt.Errorf("a sender postal address is required")
		}
		var template models.Template
		if err := s.DB.First(&template, "id = ?", n.TemplateID).Error; err != nil {
			return err
		}
		if template.HTMLBody == "" && template.HtmlFileID == "" {
			return fmt.Errorf("save your template in the template editor before scheduling")
		}
		var sender models.SMTPConfig
		if err := s.DB.First(&sender, "id = ?", n.SMTPConfigID).Error; err != nil {
			return err
		}
		if !sender.SupportsTLS {
			return fmt.Errorf("newsletter sender must support TLS")
		}
	}
	return nil
}
func (s *Service) Token(team, contact string) string {
	mac := hmac.New(sha256.New, []byte(s.Secret))
	mac.Write([]byte("unsubscribe:" + team + ":" + contact))
	return hex.EncodeToString(mac.Sum(nil))
}
func (s *Service) MessageToken(team, contact, message string) string {
	mac := hmac.New(sha256.New, []byte(s.Secret))
	mac.Write([]byte("unsubscribe:" + team + ":" + contact + ":" + message))
	return hex.EncodeToString(mac.Sum(nil))
}
func (s *Service) CheckMessageToken(team, contact, message, token string) bool {
	return len(s.Secret) >= 32 && hmac.Equal([]byte(s.MessageToken(team, contact, message)), []byte(token))
}
func (s *Service) CheckToken(team, contact, token string) bool {
	return len(s.Secret) >= 32 && hmac.Equal([]byte(s.Token(team, contact)), []byte(token))
}

// MaterializeDue commits the edition and next schedule together under a database lock.
// No queue/network work runs in this transaction; the worker discovers committed editions.
func (s *Service) MaterializeDue(ctx context.Context, now time.Time) error {
	if len(s.Secret) < 32 || !strings.HasPrefix(s.PublicURL, "https://") {
		return fmt.Errorf("newsletter delivery requires a 32-character JWT secret and HTTPS PUBLIC_API_URL")
	}
	var ids []string
	if err := s.DB.WithContext(ctx).Model(&models.Newsletter{}).Where("status = ? AND next_send_at <= ? AND is_deleted = ?", "SCHEDULED", now, false).Limit(100).Pluck("id", &ids).Error; err != nil {
		return err
	}
	for _, id := range ids {
		if err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var n models.Newsletter
			err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where("id = ? AND status = ? AND next_send_at <= ?", id, "SCHEDULED", now).First(&n).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			if err != nil {
				return err
			}
			var tpl models.Template
			if err := Scope(tx, n.TeamID).First(&tpl, "id = ?", n.TemplateID).Error; err != nil {
				return err
			}
			body, err := TemplateHTML(tx, &tpl)
			if err != nil {
				return err
			}
			cutoff := now
			campaign := models.Campaign{NewsletterID: &n.ID, Name: n.Name, TeamID: n.TeamID, TemplateID: n.TemplateID, ListID: n.ListID, SMTPConfigID: n.SMTPConfigID, Status: models.CampaignStatusScheduled, Schedule: models.CampaignScheduleOneTime, ScheduledFor: *n.NextSendAt, HTMLBody: body, Subject: n.Subject, PostalAddress: n.PostalAddress, AudienceCutoff: &cutoff, BatchSize: 100, Timezone: n.Timezone}
			if err := tx.Omit(clause.Associations).Create(&campaign).Error; err != nil {
				return err
			}
			if n.MonthDay == 0 {
				loc, err := time.LoadLocation(n.Timezone)
				if err != nil {
					return err
				}
				n.MonthDay = n.NextSendAt.In(loc).Day()
			}
			next, err := NextEdition(*n.NextSendAt, n.Cadence, n.Timezone, n.MonthDay)
			if err != nil {
				return err
			}
			// Skip missed intervals after an outage; never flood the audience with catch-up editions.
			for next != nil && !next.After(now) {
				next, err = NextEdition(*next, n.Cadence, n.Timezone, n.MonthDay)
				if err != nil {
					return err
				}
			}
			status := "SCHEDULED"
			if next == nil {
				status = "COMPLETED"
			}
			return tx.Model(&n).Updates(map[string]interface{}{"month_day": n.MonthDay, "next_send_at": next, "status": status, "editions": gorm.Expr("editions + 1"), "last_sent_at": now}).Error
		}); err != nil {
			return err
		}
	}
	return nil
}
func personalize(body string, c models.Contact) string {
	for k, v := range map[string]string{"first_name": c.FirstName, "last_name": c.LastName, "name": strings.TrimSpace(c.FirstName + " " + c.LastName), "email": c.Email, "company": c.Company} {
		body = strings.ReplaceAll(body, "{{"+k+"}}", html.EscapeString(v))
	}
	return body
}

// PrepareEdition builds bounded batches with a stable cursor and a unique delivery key.
func (s *Service) PrepareEdition(ctx context.Context, id string) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var c models.Campaign
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND newsletter_id IS NOT NULL", id).First(&c).Error; err != nil {
			return err
		}
		if c.Status != models.CampaignStatusScheduled && c.Status != models.CampaignStatusSending {
			return nil
		}
		var tpl models.Template
		if err := Scope(tx, c.TeamID).First(&tpl, "id = ?", c.TemplateID).Error; err != nil {
			return err
		}
		var sender models.SMTPConfig
		if err := Scope(tx, c.TeamID).First(&sender, "id = ?", c.SMTPConfigID).Error; err != nil {
			return err
		}
		if !sender.SupportsTLS {
			return fmt.Errorf("sender TLS is required")
		}
		var contacts []models.Contact
		q := Scope(tx, c.TeamID).Where("list_id = ? AND status = ?", c.ListID, models.SubscriberStatusActive)
		if c.AudienceCursor != "" {
			q = q.Where("id > ?", c.AudienceCursor)
		}
		if c.AudienceCutoff != nil {
			q = q.Where("created_at <= ?", *c.AudienceCutoff)
		}
		if err := q.Order("id").Limit(100).Find(&contacts).Error; err != nil {
			return err
		}
		for _, contact := range contacts {
			key := c.ID + ":" + contact.ID
			messageID := uuid.NewString()
			unsub := s.PublicURL + "/public/unsubscribe/" + c.TeamID + "/" + contact.ID + "/" + s.MessageToken(c.TeamID, contact.ID, messageID) + "?message=" + messageID
			body := personalize(c.HTMLBody, contact) + `<footer style="padding:24px;text-align:center;color:#71717a;font:12px Arial">` + html.EscapeString(c.PostalAddress) + `<br><a href="` + unsub + `">Unsubscribe</a></footer>`
			body = utils.TrackEmailHTML(body, messageID, s.PublicURL, s.Secret)
			tracking := true
			email := models.Email{Base: models.Base{ID: messageID}, ClickTrackingEnabled: &tracking, DeliveryKey: &key, UnsubscribeURL: unsub, From: sender.FromEmail, To: contact.Email, Subject: personalize(c.Subject, contact), Body: base64.StdEncoding.EncodeToString([]byte(body)), Status: models.EmailStatusPending, TeamID: c.TeamID, ContactID: contact.ID, TemplateID: c.TemplateID, SMTPConfigID: c.SMTPConfigID, CategoryID: tpl.CategoryID, CampaignID: c.ID}
			if err := tx.Omit(clause.Associations).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "delivery_key"}}, DoNothing: true}).Create(&email).Error; err != nil {
				return err
			}
			c.AudienceCursor = contact.ID
		}
		status := models.CampaignStatusSending
		if len(contacts) < 100 {
			status = models.CampaignStatusCompleted
		} // All messages are prepared; delivery status is tracked on each email.
		return tx.Model(&c).Updates(map[string]interface{}{"audience_cursor": c.AudienceCursor, "status": status, "processed": gorm.Expr("processed + ?", len(contacts))}).Error
	})
}

// TemplateHTML supports templates saved through the existing file-backed editor.
func TemplateHTML(db *gorm.DB, template *models.Template) (string, error) {
	if template.HTMLBody != "" {
		return template.HTMLBody, nil
	}
	if template.HtmlFileID == "" {
		return "", fmt.Errorf("template has no HTML content")
	}
	var file models.File
	if err := Scope(db, template.TeamID).First(&file, "id = ?", template.HtmlFileID).Error; err != nil {
		return "", err
	}
	if file.SignedURL == "" {
		return "", fmt.Errorf("template file is unavailable")
	}
	return utils.GetHTMLFromURL(file.SignedURL)
}
