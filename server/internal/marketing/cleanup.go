package marketing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/models"
)

// List cleanup: readers who ignored CleanupIgnored sends over CleanupSpan get one question,
// and become INACTIVE (never deleted or unsubscribed) if they don't answer within CleanupWait.
const (
	CleanupIgnored = 5
	CleanupSpan    = 60 * 24 * time.Hour
	CleanupWait    = 14 * 24 * time.Hour
)

func (s *Service) KeepToken(team, contact string) string {
	mac := hmac.New(sha256.New, []byte(s.Secret))
	mac.Write([]byte("keep:" + team + ":" + contact))
	return hex.EncodeToString(mac.Sum(nil))
}
func (s *Service) CheckKeepToken(team, contact, token string) bool {
	return len(s.Secret) >= 32 && hmac.Equal([]byte(s.KeepToken(team, contact)), []byte(token))
}

// KeepSending marks a reader engaged and restarts their count of ignored mail.
func KeepSending(tx *gorm.DB, contact *models.Contact, now time.Time) error {
	if contact.Status != models.SubscriberStatusActive && contact.Status != models.SubscriberStatusInactive {
		return nil
	}
	if err := tx.Model(&models.Contact{}).Where("id = ?", contact.ID).UpdateColumns(map[string]interface{}{"status": models.SubscriberStatusActive, "reengage_sent_at": nil, "cleanup_from": now}).Error; err != nil {
		return err
	}
	return models.SyncSubscribersCountByID(tx, contact.ListID)
}

// CleanUpLists runs once a day. Every step is a conditional update, so retries are safe.
func (s *Service) CleanUpLists(ctx context.Context, now time.Time) error {
	if len(s.Secret) < 32 || !strings.HasPrefix(s.PublicURL, "https://") {
		return nil // The keep link needs a signing secret and a public URL.
	}
	db := s.DB.WithContext(ctx)
	// contacts returns a fresh query over readers in lists that have cleanup on.
	contacts := func(status models.SubscriberStatus, where string, args ...interface{}) *gorm.DB {
		enabled := db.Model(&models.MailingList{}).Select("id").Where("list_cleanup = ? AND is_deleted = ?", true, false)
		return db.Model(&models.Contact{}).Where("list_id IN (?) AND is_deleted = ? AND status = ?", enabled, false, status).Where(where, args...)
	}
	clicked := func(since string) string {
		return "contacts." + since + " IS NOT NULL AND EXISTS (SELECT 1 FROM email_trackings t WHERE t.contact_id = contacts.id AND t.event = 'click' AND t.timestamp > contacts." + since + ")"
	}
	// changed updates matching readers and refreshes the subscriber counts of their lists.
	changed := func(q func() *gorm.DB, values map[string]interface{}) error {
		var lists []string
		if err := q().Distinct().Pluck("list_id", &lists).Error; err != nil {
			return err
		}
		if err := q().UpdateColumns(values).Error; err != nil {
			return err
		}
		for _, id := range lists {
			if err := models.SyncSubscribersCountByID(db, id); err != nil {
				return err
			}
		}
		return nil
	}
	// A paused reader who clicks an old email is back.
	if err := changed(func() *gorm.DB { return contacts(models.SubscriberStatusInactive, clicked("cleanup_from")) }, map[string]interface{}{"status": models.SubscriberStatusActive, "cleanup_from": now}); err != nil {
		return err
	}
	// Clicking any email while being asked answers the question too.
	if err := contacts(models.SubscriberStatusActive, clicked("reengage_sent_at")).UpdateColumns(map[string]interface{}{"reengage_sent_at": nil, "cleanup_from": now}).Error; err != nil {
		return err
	}
	// No answer in time: pause them, and restart the count in case the owner turns them back on.
	if err := changed(func() *gorm.DB {
		return contacts(models.SubscriberStatusActive, "reengage_sent_at <= ?", now.Add(-CleanupWait))
	}, map[string]interface{}{"status": models.SubscriberStatusInactive, "reengage_sent_at": nil, "cleanup_from": now}); err != nil {
		return err
	}
	// Ask everyone who qualifies. Ignored mail is newsletter and campaign mail since the last restart.
	since := "COALESCE(contacts.cleanup_from, contacts.created_at)"
	ignored := "FROM emails e WHERE e.contact_id = contacts.id AND e.campaign_id IS NOT NULL AND e.status IN ('SENT','OPENED','CLICKED') AND e.created_at > " + since
	for {
		var ids []string
		err := contacts(models.SubscriberStatusActive, "reengage_sent_at IS NULL").
			Where("(SELECT COUNT(*) "+ignored+") >= ? AND (SELECT MIN(e.created_at) "+ignored+") <= ?", CleanupIgnored, now.Add(-CleanupSpan)).
			Where("NOT EXISTS (SELECT 1 FROM email_trackings t WHERE t.contact_id = contacts.id AND t.event IN ('click','open') AND t.timestamp > "+since+")").
			Order("id").Limit(200).Pluck("id", &ids).Error
		if err != nil {
			return err
		}
		for _, id := range ids {
			if err := s.askToKeep(ctx, id, now); err != nil {
				return err
			}
		}
		if len(ids) < 200 {
			return nil
		}
	}
}

// askToKeep claims the contact and queues the question in one transaction, so it is sent once per cycle.
func (s *Service) askToKeep(ctx context.Context, contactID string, now time.Time) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		claim := tx.Model(&models.Contact{}).Where("id = ? AND status = ? AND reengage_sent_at IS NULL", contactID, models.SubscriberStatusActive).UpdateColumn("reengage_sent_at", now)
		if claim.Error != nil || claim.RowsAffected == 0 {
			return claim.Error
		}
		var contact models.Contact
		if err := tx.First(&contact, "id = ?", contactID).Error; err != nil {
			return err
		}
		// Reply from the sender and address the reader last heard from.
		var last models.Email
		if err := tx.Where("contact_id = ? AND campaign_id IS NOT NULL", contactID).Order("created_at DESC").First(&last).Error; err != nil {
			return err
		}
		var campaign models.Campaign
		if err := tx.First(&campaign, "id = ?", last.CampaignID).Error; err != nil {
			return err
		}
		var team models.Team
		if err := tx.First(&team, "id = ?", contact.TeamID).Error; err != nil {
			return err
		}
		messageID := uuid.NewString()
		unsub := s.PublicURL + "/public/unsubscribe/" + contact.TeamID + "/" + contact.ID + "/" + s.MessageToken(contact.TeamID, contact.ID, messageID) + "?message=" + messageID
		keep := s.PublicURL + "/public/keep/" + contact.TeamID + "/" + contact.ID + "/" + s.KeepToken(contact.TeamID, contact.ID)
		name := strings.Join(strings.Fields(team.Name), " ")
		sender := html.EscapeString(name)
		greeting := "Hi there,"
		if contact.FirstName != "" {
			greeting = "Hi " + html.EscapeString(contact.FirstName) + ","
		}
		// Not click-tracked: only the reader's own button press should count as an answer.
		body := `<div style="max-width:520px;margin:0 auto;padding:32px 24px;font:16px/1.5 Arial;color:#18181b"><p>` + greeting + `</p>` +
			`<p>You haven't opened or clicked our recent emails, so we're checking before we send more. If you'd still like to hear from ` + sender + `, press the button below.</p>` +
			`<p style="margin:28px 0"><a href="` + keep + `" style="background:#18181b;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none">Yes, keep sending</a></p>` +
			`<p>If we don't hear back within 14 days, we'll stop sending you these emails. You can unsubscribe at any time.</p></div>` +
			`<footer style="padding:24px;text-align:center;color:#71717a;font:12px Arial">` + html.EscapeString(campaign.PostalAddress) + `<br><a href="` + unsub + `">Unsubscribe</a></footer>`
		key := fmt.Sprintf("reengage:%s:%d", contact.ID, now.Unix())
		tracking := false
		email := models.Email{Base: models.Base{ID: messageID}, ClickTrackingEnabled: &tracking, DeliveryKey: &key, UnsubscribeURL: unsub, From: last.From, To: contact.Email, Subject: "Still want emails from " + name + "?", Body: base64.StdEncoding.EncodeToString([]byte(body)), Status: models.EmailStatusPending, TeamID: contact.TeamID, ContactID: contact.ID, SMTPConfigID: last.SMTPConfigID, CategoryID: last.CategoryID}
		return tx.Omit(clause.Associations).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "delivery_key"}}, DoNothing: true}).Create(&email).Error
	})
}
