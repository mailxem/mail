package receiving

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"math"
	"net/mail"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/sending"
	"kori/internal/utils"
)

var ErrMailboxQuota = errors.New("managed mailbox storage quota exceeded")
var ErrMailboxInactive = errors.New("managed mailbox is not accepting mail")
var providerIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,256}$`)

type ObjectStore interface {
	Get(context.Context, string) ([]byte, error)
	Copy(context.Context, string, string) error
}

type SESReceipt struct {
	NotificationType string `json:"notificationType"`
	Mail             struct {
		MessageID string    `json:"messageId"`
		Timestamp time.Time `json:"timestamp"`
	} `json:"mail"`
	Receipt struct {
		Recipients  []string `json:"recipients"`
		SpamVerdict struct {
			Status string `json:"status"`
		} `json:"spamVerdict"`
		VirusVerdict struct {
			Status string `json:"status"`
		} `json:"virusVerdict"`
		Action struct {
			Type       string `json:"type"`
			TopicARN   string `json:"topicArn"`
			BucketName string `json:"bucketName"`
			ObjectKey  string `json:"objectKey"`
		} `json:"action"`
	} `json:"receipt"`
}

func (s *Service) ProcessEnvelope(ctx context.Context, envelope sending.Notification, store ObjectStore) error {
	if !s.Config.Enabled {
		return ErrDisabled
	}
	if err := sending.VerifySNSNotification(ctx, envelope, s.Config.TopicARN, s.Config.Region, s.Now()); err != nil {
		return errors.New("untrusted receiving notification")
	}
	if envelope.Type != "Notification" {
		return errors.New("unsupported receiving notification")
	}
	var event SESReceipt
	if json.Unmarshal([]byte(envelope.Message), &event) != nil {
		return errors.New("invalid SES receipt")
	}
	return s.ProcessReceipt(ctx, event, store)
}

func (s *Service) ProcessReceipt(ctx context.Context, event SESReceipt, store ObjectStore) error {
	if !s.Config.Enabled {
		return ErrDisabled
	}
	if store == nil {
		return errors.New("receiving object store unavailable")
	}
	if event.NotificationType != "Received" || !providerIDPattern.MatchString(event.Mail.MessageID) || event.Receipt.Action.Type != "S3" || event.Receipt.Action.BucketName != s.Config.Bucket || event.Receipt.Action.TopicARN != s.Config.TopicARN || event.Receipt.Action.ObjectKey != s.Config.Prefix+event.Mail.MessageID {
		return errors.New("invalid SES receipt routing")
	}
	if event.Receipt.VirusVerdict.Status != "PASS" || !(event.Receipt.SpamVerdict.Status == "PASS" || event.Receipt.SpamVerdict.Status == "FAIL") {
		return s.reject(ctx, event, "quarantined")
	}
	raw, err := store.Get(ctx, event.Receipt.Action.ObjectKey)
	if err != nil {
		return err
	}
	if int64(len(raw)) > s.Config.MaxMessageBytes {
		return s.reject(ctx, event, "message_too_large")
	}
	parsed, err := utils.ParseEmail(bytes.NewReader(raw))
	if err != nil {
		return s.reject(ctx, event, "mime_invalid")
	}
	addresses := map[string]bool{}
	for _, rawAddress := range event.Receipt.Recipients {
		a, e := mail.ParseAddress(rawAddress)
		if e == nil {
			addresses[strings.ToLower(a.Address)] = true
		}
	}
	if len(addresses) == 0 {
		return errors.New("receipt has no valid envelope recipients")
	}
	var boxes []Mailbox
	if err = s.DB.WithContext(ctx).Where("lower(address) IN ?", mapKeys(addresses)).Find(&boxes).Error; err != nil {
		return err
	}
	if len(boxes) == 0 {
		return s.reject(ctx, event, "no_owned_recipient")
	}
	eligible := boxes[:0]
	fullBoxes := []Mailbox{}
	var preflightErr error
	for _, box := range boxes {
		var existing int64
		if e := s.DB.WithContext(ctx).Model(&Message{}).Where("mailbox_id = ? AND provider_id = ?", box.ID, event.Mail.MessageID).Count(&existing).Error; e != nil {
			return e
		}
		if existing > 0 {
			eligible = append(eligible, box)
			continue
		}
		if box.Status == "storage_full" {
			preflightErr = errors.Join(preflightErr, ErrMailboxQuota)
			continue
		}
		if !box.Active {
			preflightErr = errors.Join(preflightErr, ErrMailboxInactive)
			continue
		}
		var d sending.Domain
		var a sending.Account
		if e := s.DB.WithContext(ctx).Where("id = ? AND team_id = ? AND ownership = true", box.DomainID, box.TeamID).First(&d).Error; e != nil {
			if errors.Is(e, gorm.ErrRecordNotFound) {
				preflightErr = errors.Join(preflightErr, ErrMailboxInactive)
				continue
			}
			return e
		}
		if e := s.DB.WithContext(ctx).Where("team_id = ? AND suspended = false", box.TeamID).First(&a).Error; e != nil {
			if errors.Is(e, gorm.ErrRecordNotFound) {
				preflightErr = errors.Join(preflightErr, ErrMailboxInactive)
				continue
			}
			return e
		}
		if box.UsageBytes+int64(len(raw)) <= box.QuotaBytes {
			eligible = append(eligible, box)
		} else if e := s.markStorageFull(ctx, box); e != nil {
			return e
		} else {
			fullBoxes = append(fullBoxes, box)
			preflightErr = errors.Join(preflightErr, ErrMailboxQuota)
		}
	}
	if len(eligible) == 0 {
		for _, box := range fullBoxes {
			if e := s.reconcileDomain(ctx, box.TeamID, box.DomainID); e != nil {
				return e
			}
		}
		if preflightErr != nil {
			return preflightErr
		}
		return ErrMailboxInactive
	}
	durableKey := "mail/" + event.Mail.MessageID
	if err = store.Copy(ctx, event.Receipt.Action.ObjectKey, durableKey); err != nil {
		return err
	}
	folder := "INBOX"
	if event.Receipt.SpamVerdict.Status == "FAIL" {
		folder = "Spam"
	}
	var quotaErr bool
	for _, box := range eligible {
		full, e := s.indexMailbox(ctx, box, event, parsed, durableKey, int64(len(raw)), folder, addresses)
		if e != nil {
			if errors.Is(e, ErrMailboxInactive) {
				preflightErr = errors.Join(preflightErr, e)
				continue
			}
			if errors.Is(e, ErrMailboxQuota) {
				preflightErr = errors.Join(preflightErr, e)
				continue
			}
			return e
		}
		if full {
			quotaErr = true
		}
	}
	for _, box := range fullBoxes {
		if e := s.reconcileDomain(ctx, box.TeamID, box.DomainID); e != nil {
			return e
		}
	}
	if quotaErr {
		preflightErr = errors.Join(preflightErr, ErrMailboxQuota)
	}
	if preflightErr != nil {
		return preflightErr
	}
	return nil
}

func mapKeys(m map[string]bool) []string {
	v := make([]string, 0, len(m))
	for k := range m {
		v = append(v, k)
	}
	return v
}
func (s *Service) reject(ctx context.Context, event SESReceipt, reason string) error {
	normalized := make([]string, 0, len(event.Receipt.Recipients))
	for _, raw := range event.Receipt.Recipients {
		if a, e := mail.ParseAddress(raw); e == nil {
			normalized = append(normalized, strings.ToLower(a.Address))
		}
	}
	sort.Strings(normalized)
	recipients, _ := json.Marshal(normalized)
	sum := sha256.Sum256(recipients)
	recipientHash := hex.EncodeToString(sum[:])
	return s.DB.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "provider_id"}, {Name: "reason"}, {Name: "recipient_set_hash"}}, DoNothing: true}).Create(&Rejection{ID: uuid.NewString(), ProviderID: event.Mail.MessageID, S3Key: event.Receipt.Action.ObjectKey, Reason: reason, RecipientSetHash: recipientHash, RecipientsJSON: recipients, CreatedAt: s.Now()}).Error
}

func attachmentMetadata(items []utils.EmailAttachment) []byte {
	metadata := make([]utils.EmailAttachment, len(items))
	copy(metadata, items)
	for i := range metadata {
		size := len(metadata[i].Data)
		metadata[i].Data = nil
		metadata[i].Size = int64(size)
		metadata[i].AttachmentID = fmt.Sprintf("part:%d", i)
	}
	raw, _ := json.Marshal(metadata)
	return raw
}
func (s *Service) markStorageFull(ctx context.Context, box Mailbox) error {
	return s.DB.WithContext(ctx).Model(&Mailbox{}).Where("id = ? AND team_id = ? AND active = true", box.ID, box.TeamID).Updates(map[string]any{"status": "storage_full", "updated_at": s.Now()}).Error
}
func (s *Service) indexMailbox(ctx context.Context, box Mailbox, event SESReceipt, parsed *utils.ParsedMail, key string, size int64, folder string, envelope map[string]bool) (bool, error) {
	full := false
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked Mailbox
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", box.ID).First(&locked).Error; err != nil {
			return err
		}
		if locked.TeamID != box.TeamID || locked.DomainID != box.DomainID || !envelope[strings.ToLower(locked.Address)] {
			return ErrMailboxInactive
		}
		var existing int64
		if err := tx.Model(&Message{}).Where("mailbox_id = ? AND provider_id = ?", box.ID, event.Mail.MessageID).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return nil
		}
		if locked.Status == "storage_full" {
			return ErrMailboxQuota
		}
		if !locked.Active {
			return ErrMailboxInactive
		}
		var d sending.Domain
		if e := tx.Where("id = ? AND team_id = ? AND ownership = true", locked.DomainID, locked.TeamID).First(&d).Error; e != nil {
			if errors.Is(e, gorm.ErrRecordNotFound) {
				return ErrMailboxInactive
			}
			return e
		}
		var a sending.Account
		if e := tx.Where("team_id = ? AND suspended = false", locked.TeamID).First(&a).Error; e != nil {
			if errors.Is(e, gorm.ErrRecordNotFound) {
				return ErrMailboxInactive
			}
			return e
		}
		if locked.UsageBytes+size > locked.QuotaBytes {
			if err := tx.Model(&locked).Updates(map[string]any{"status": "storage_full", "updated_at": s.Now()}).Error; err != nil {
				return err
			}
			full = true
			return nil
		}
		var maxUID uint32
		if err := tx.Model(&Message{}).Where("mailbox_id = ?", box.ID).Select("COALESCE(MAX(uid),0)").Scan(&maxUID).Error; err != nil {
			return err
		}
		uid := maxUID + 1
		if maxUID == math.MaxUint32 {
			return errors.New("managed mailbox UID space exhausted")
		}
		body := parsed.BodyText
		if body == "" {
			body = parsed.BodyHTML
		}
		snippetRunes := []rune(html.EscapeString(strings.TrimSpace(body)))
		if len(snippetRunes) > 500 {
			snippetRunes = snippetRunes[:500]
		}
		snippet := string(snippetRunes)
		sentAt := parsed.Date
		if sentAt.IsZero() {
			sentAt = event.Mail.Timestamp
		}
		if sentAt.IsZero() {
			sentAt = s.Now()
		}
		message := Message{ID: uuid.NewString(), MailboxID: box.ID, TeamID: box.TeamID, ProviderID: event.Mail.MessageID, UID: uid, UIDValidity: locked.UIDValidity, Folder: folder, FromAddress: formatAddresses(parsed.From), ToAddress: formatAddresses(parsed.To), CcAddress: formatAddresses(parsed.Cc), BccAddress: formatAddresses(parsed.Bcc), ReplyTo: formatAddresses(parsed.ReplyTo), Subject: parsed.Subject, RFCMessageID: parsed.MessageID, SentAt: sentAt, Snippet: snippet, RawKey: key, RawSize: size, AttachmentsJSON: attachmentMetadata(parsed.Attachments), CreatedAt: s.Now(), UpdatedAt: s.Now()}
		if err := tx.Create(&message).Error; err != nil {
			return err
		}
		if err := tx.Model(&locked).Updates(map[string]any{"usage_bytes": locked.UsageBytes + size, "updated_at": s.Now()}).Error; err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "provider_id"}}, DoUpdates: clause.Assignments(map[string]any{"s3_key": key, "size": size, "references": gorm.Expr("managed_receiving_blobs.references + 1")})}).Create(&Blob{ProviderID: event.Mail.MessageID, S3Key: key, Size: size, References: 1, CreatedAt: s.Now()}).Error
	})
	if err != nil {
		return false, err
	}
	if full {
		if err := s.reconcileDomain(ctx, box.TeamID, box.DomainID); err != nil {
			return true, err
		}
	}
	return full, nil
}

func formatAddresses(values []*mail.Address) string {
	parts := make([]string, 0, len(values))
	for _, v := range values {
		parts = append(parts, v.String())
	}
	return strings.Join(parts, ", ")
}
