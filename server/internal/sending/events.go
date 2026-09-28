package sending

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/models"
)

type Notification struct {
	Type             string `json:"Type"`
	MessageID        string `json:"MessageId"`
	TopicARN         string `json:"TopicArn"`
	Message          string `json:"Message"`
	Timestamp        string `json:"Timestamp"`
	Subject          string `json:"Subject"`
	Token            string `json:"Token"`
	SubscribeURL     string `json:"SubscribeURL"`
	SignatureVersion string `json:"SignatureVersion"`
	Signature        string `json:"Signature"`
	SigningCertURL   string `json:"SigningCertURL"`
}

func (n Notification) canonical() (string, error) {
	fields := []string{"Message", "MessageId"}
	values := map[string]string{"Message": n.Message, "MessageId": n.MessageID, "Subject": n.Subject, "Timestamp": n.Timestamp, "TopicArn": n.TopicARN, "Type": n.Type, "Token": n.Token, "SubscribeURL": n.SubscribeURL}
	switch n.Type {
	case "Notification":
		if n.Subject != "" {
			fields = append(fields, "Subject")
		}
		fields = append(fields, "Timestamp", "TopicArn", "Type")
	case "SubscriptionConfirmation":
		fields = append(fields, "SubscribeURL", "Timestamp", "Token", "TopicArn", "Type")
	default:
		return "", errors.New("unsupported SNS message type")
	}
	var b strings.Builder
	for _, f := range fields {
		b.WriteString(f + "\n" + values[f] + "\n")
	}
	return b.String(), nil
}

var snsClient = &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

func (s *Service) VerifyNotification(ctx context.Context, n Notification) error {
	if !s.Config.Enabled || n.TopicARN != s.Config.TopicARN || len(n.MessageID) > 128 || n.MessageID == "" {
		return ErrDenied
	}
	stamp, e := time.Parse(time.RFC3339, n.Timestamp)
	if e != nil || stamp.After(s.Now().Add(5*time.Minute)) || stamp.Before(s.Now().Add(-30*24*time.Hour)) {
		return ErrDenied
	}
	u, e := url.Parse(n.SigningCertURL)
	if e != nil || u.Scheme != "https" || u.Host != "sns."+s.Config.Region+".amazonaws.com" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || !regexp.MustCompile(`^/SimpleNotificationService-[a-zA-Z0-9]+\.pem$`).MatchString(u.Path) {
		return ErrDenied
	}
	canonical, e := n.canonical()
	if e != nil {
		return e
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if e != nil {
		return e
	}
	resp, e := snsClient.Do(req)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ErrDenied
	}
	data, e := io.ReadAll(io.LimitReader(resp.Body, 65537))
	if e != nil || len(data) > 65536 {
		return ErrDenied
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return ErrDenied
	}
	cert, e := x509.ParseCertificate(block.Bytes)
	if e != nil || s.Now().Before(cert.NotBefore) || s.Now().After(cert.NotAfter) {
		return ErrDenied
	}
	key, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok || key.N.BitLen() < 2048 {
		return ErrDenied
	}
	return verifySignature(n, canonical, key)
}
func verifySignature(n Notification, canonical string, key *rsa.PublicKey) error {
	signature, e := base64.StdEncoding.DecodeString(n.Signature)
	if e != nil {
		return ErrDenied
	}
	switch n.SignatureVersion {
	case "1":
		digest := sha1.Sum([]byte(canonical))
		e = rsa.VerifyPKCS1v15(key, crypto.SHA1, digest[:], signature)
	case "2":
		digest := sha256.Sum256([]byte(canonical))
		e = rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature)
	default:
		return ErrDenied
	}
	return e
}
func (s *Service) ConfirmSubscription(ctx context.Context, n Notification) error {
	// Use a constructed, pinned AWS URL; never fetch the supplied SubscribeURL.
	u := url.URL{Scheme: "https", Host: "sns." + s.Config.Region + ".amazonaws.com"}
	q := u.Query()
	q.Set("Action", "ConfirmSubscription")
	q.Set("Version", "2010-03-31")
	q.Set("TopicArn", s.Config.TopicARN)
	q.Set("Token", n.Token)
	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if e != nil {
		return e
	}
	resp, e := snsClient.Do(req)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("SNS subscription confirmation failed")
	}
	return nil
}

type Feedback struct {
	EventType string `json:"eventType"`
	Mail      struct {
		MessageID string              `json:"messageId"`
		Source    string              `json:"source"`
		Tags      map[string][]string `json:"tags"`
	} `json:"mail"`
	Bounce struct {
		BounceType string `json:"bounceType"`
		Recipients []struct {
			Email string `json:"emailAddress"`
		} `json:"bouncedRecipients"`
	} `json:"bounce"`
	Complaint struct {
		Recipients []struct {
			Email string `json:"emailAddress"`
		} `json:"complainedRecipients"`
	} `json:"complaint"`
}

func (s *Service) ApplyFeedback(ctx context.Context, eventID string, raw []byte) error {
	var f Feedback
	if e := json.Unmarshal(raw, &f); e != nil {
		return e
	}
	ids := f.Mail.Tags["xem_message"]
	configs := f.Mail.Tags["ses:configuration-set"]
	if len(ids) != 1 || len(configs) != 1 || f.Mail.MessageID == "" {
		return errors.New("event has no Xem routing metadata")
	}
	kinds := map[string]string{"Send": "SENT", "Delivery": "DELIVERED", "Bounce": "BOUNCED", "Complaint": "COMPLAINED", "Reject": "FAILED", "DeliveryDelay": "DELAYED"}
	state, ok := kinds[f.EventType]
	if !ok {
		return nil
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var account Account
		reputationEvent := f.EventType == "Complaint" || (f.EventType == "Bounce" && f.Bounce.BounceType == "Permanent")
		if reputationEvent {
			var routed Message
			if e := tx.Select("team_id").First(&routed, "id = ?", ids[0]).Error; e != nil {
				return e
			}
			// Account-before-message ordering matches submission; serialize feedback
			// counts with other events and with approval/suspension decisions.
			if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&account, "team_id = ?", routed.TeamID).Error; e != nil {
				return e
			}
		}
		var m Message
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&m, "id = ?", ids[0]).Error; e != nil {
			return e
		}
		source, e := address(f.Mail.Source)
		if e != nil || source != m.From || configs[0] != tenant(m.TeamID) || (m.ProviderID != "" && m.ProviderID != f.Mail.MessageID) || m.Status == "QUEUED" {
			return ErrDenied
		}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&Event{ID: eventID, TeamID: m.TeamID, MessageID: m.ID, Kind: f.EventType, Detail: "SES reported " + f.EventType, CreatedAt: s.Now()})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		recipients := map[string]bool{}
		for _, r := range strings.Split(m.Recipients, ",") {
			recipients[r] = true
		}
		blocked := []string{}
		if f.EventType == "Bounce" && f.Bounce.BounceType == "Permanent" {
			for _, r := range f.Bounce.Recipients {
				blocked = append(blocked, r.Email)
			}
		}
		if f.EventType == "Complaint" {
			for _, r := range f.Complaint.Recipients {
				blocked = append(blocked, r.Email)
			}
		}
		if reputationEvent && len(blocked) == 0 {
			return errors.New("feedback has no affected recipients")
		}
		for _, r := range blocked {
			r = strings.ToLower(r)
			if !recipients[r] {
				return errors.New("feedback recipient does not belong to message")
			}
			suppression := Suppression{TeamID: m.TeamID, Address: r, Reason: f.EventType, CreatedAt: s.Now()}
			if e := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&suppression).Error; e != nil {
				return e
			}
			contactStatus := "BOUNCED"
			if f.EventType == "Complaint" {
				contactStatus = "COMPLAINED"
			}
			var contacts []models.Contact
			if e := tx.Where("team_id = ? AND LOWER(email) = ?", m.TeamID, r).Find(&contacts).Error; e != nil {
				return e
			}
			for i := range contacts {
				if e := tx.Model(&contacts[i]).Update("status", contactStatus).Error; e != nil {
					return e
				}
			}
		}
		if len(blocked) > 0 && !account.Suspended {
			action := ""
			if f.EventType == "Complaint" {
				action = "auto_suspend_complaint"
			} else {
				var hardBounces int64
				if e := tx.Model(&Suppression{}).Where("team_id = ? AND reason = ? AND created_at >= ?", m.TeamID, "Bounce", s.Now().Add(-24*time.Hour)).Count(&hardBounces).Error; e != nil {
					return e
				}
				if hardBounces >= 5 {
					action = "auto_suspend_bounces"
				}
			}
			if action != "" {
				if e := tx.Model(&account).Update("suspended", true).Error; e != nil {
					return e
				}
				if e := tx.Create(&Audit{ID: uuid.NewString(), TeamID: m.TeamID, Actor: "system:ses-feedback", Action: action, CreatedAt: s.Now()}).Error; e != nil {
					return e
				}
			}
		}
		rank := map[string]int{"QUEUED": 0, "SENDING": 0, "DELIVERY_UNKNOWN": 0, "SENT": 1, "DELAYED": 2, "DELIVERED": 3, "FAILED": 4, "BOUNCED": 5, "COMPLAINED": 6}
		updates := map[string]any{"provider_id": f.Mail.MessageID}
		if rank[state] >= rank[m.Status] {
			updates["status"] = state
			updates["detail"] = fmt.Sprintf("SES reported %s.", f.EventType)
		}
		if e := tx.Model(&m).Updates(updates).Error; e != nil {
			return e
		}
		if m.EmailID != "" && rank[state] >= rank[m.Status] {
			emailState := state
			if state == "DELIVERED" || state == "DELAYED" {
				emailState = "SENT"
			}
			return tx.Model(&models.Email{}).Where("id = ? AND team_id = ?", m.EmailID, m.TeamID).Updates(map[string]any{"status": emailState, "sent_at": s.Now()}).Error
		}
		return nil
	})
}
