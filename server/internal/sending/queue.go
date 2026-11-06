package sending

import (
	"context"
	"encoding/json"
	"fmt"
	busEvents "kori/internal/events"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/models"
)

// EventQueue is the part of the SQS client the feedback consumer uses.
type EventQueue interface {
	ReceiveMessage(context.Context, *sqs.ReceiveMessageInput, ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessageBatch(context.Context, *sqs.DeleteMessageBatchInput, ...func(*sqs.Options)) (*sqs.DeleteMessageBatchOutput, error)
}

func NewEventQueue(ctx context.Context, c Config) (*sqs.Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(c.Region))
	if err != nil {
		return nil, err
	}
	return sqs.NewFromConfig(cfg), nil
}

// ConsumeEvents polls the SES event queue until ctx is cancelled.
func (s *Service) ConsumeEvents(ctx context.Context) {
	for ctx.Err() == nil {
		if err := s.PollEvents(ctx); err != nil && ctx.Err() == nil {
			log.Print("SES event queue poll failed; retrying")
			select {
			case <-ctx.Done():
			case <-time.After(time.Second):
			}
		}
	}
}

// fastStates are applied in one transaction per received batch: they need no
// account lock. Bounce, Complaint and Reject still go through ApplyFeedback.
var fastStates = map[string]string{"Send": "SENT", "Delivery": "DELIVERED", "DeliveryDelay": "DELAYED"}

type queuedEvent struct {
	id     string // SNS MessageId, the same event ID the webhook stores
	raw    string
	f      Feedback
	handle *string
}

// PollEvents receives up to 10 SNS-wrapped SES events and deletes the ones
// whose changes committed. Anything else becomes visible again after the
// visibility timeout and, after the queue's maxReceiveCount, moves to its
// dead-letter queue.
func (s *Service) PollEvents(ctx context.Context) error {
	out, err := s.Events.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{QueueUrl: aws.String(s.Config.EventsQueueURL), MaxNumberOfMessages: 10, WaitTimeSeconds: 20, VisibilityTimeout: 60})
	if err != nil || len(out.Messages) == 0 {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	done := []sqstypes.DeleteMessageBatchRequestEntry{}
	fast, slow := []queuedEvent{}, []queuedEvent{}
	for _, m := range out.Messages {
		// The queue is reachable only with this account's IAM credentials and its
		// policy admits only the feedback topic, so the SNS signature is not checked
		// again here (the HTTPS webhook must). The topic check rejects any other
		// subscription's messages.
		var n Notification
		if json.Unmarshal([]byte(aws.ToString(m.Body)), &n) != nil || n.Type != "Notification" || n.TopicARN != s.Config.TopicARN || n.MessageID == "" || len(n.MessageID) > 128 {
			log.Print("SES event queue message is not a feedback notification; leaving it for the dead-letter queue")
			continue
		}
		e := queuedEvent{id: n.MessageID, raw: n.Message, handle: m.ReceiptHandle}
		if json.Unmarshal([]byte(n.Message), &e.f) == nil && fastStates[e.f.EventType] != "" {
			fast = append(fast, e)
		} else {
			slow = append(slow, e)
		}
	}
	if len(fast) > 0 {
		applied, err := s.applyDeliveryBatch(ctx, fast)
		if err != nil {
			log.Print("SES delivery batch failed; leaving it for retry")
		}
		for i, ok := range applied {
			if ok {
				done = append(done, deleteEntry(len(done), fast[i].handle))
			} else if err == nil {
				log.Print("SES delivery event does not match its message; leaving it for the dead-letter queue")
			}
		}
	}
	// Bounces, complaints and rejects need the account lock; they are rare.
	for _, e := range slow {
		if err := s.ApplyFeedback(ctx, e.id, []byte(e.raw)); err != nil {
			log.Print("SES feedback processing failed; leaving it for retry")
			continue
		}
		done = append(done, deleteEntry(len(done), e.handle))
	}
	if len(done) == 0 {
		return nil
	}
	deleteCtx, deleteCancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer deleteCancel()
	result, err := s.Events.DeleteMessageBatch(deleteCtx, &sqs.DeleteMessageBatchInput{QueueUrl: aws.String(s.Config.EventsQueueURL), Entries: done})
	if err != nil || len(result.Failed) > 0 {
		log.Print("SES feedback committed but its acknowledgement failed; duplicate delivery is expected and ignored")
	}
	return nil
}

func deleteEntry(i int, handle *string) sqstypes.DeleteMessageBatchRequestEntry {
	return sqstypes.DeleteMessageBatchRequestEntry{Id: aws.String(fmt.Sprint(i)), ReceiptHandle: handle}
}

// applyDeliveryBatch applies Send, Delivery and DeliveryDelay events in one
// transaction with the same checks and writes ApplyFeedback makes for each, in
// order. applied[i] reports whether event i is recorded (or already was); an
// event that fails the ownership checks is skipped without affecting the rest.
// On error nothing is written.
func (s *Service) applyDeliveryBatch(ctx context.Context, events []queuedEvent) (applied []bool, err error) {
	// Bus events to emit after commit, as ApplyFeedback does: one per newly
	// recorded SES event, never for test messages.
	type pendingEmit struct {
		name  string
		event *models.EmailEvent
	}
	var emits []pendingEmit
	ids := []string{}
	eventIDs := []string{}
	for _, e := range events {
		if tag := e.f.Mail.Tags["xem_message"]; len(tag) == 1 && uuid.Validate(tag[0]) == nil {
			ids = append(ids, tag[0])
		}
		eventIDs = append(eventIDs, e.id)
	}
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		applied = make([]bool, len(events))
		if len(ids) == 0 {
			return nil
		}
		// One locked read in ID order so concurrent batches cannot deadlock each other.
		var rows []Message
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id", "team_id", "domain_id", "email_id", "from", "recipients", "status", "detail", "provider_id", "is_test").Where("id IN ?", ids).Order("id").Find(&rows).Error; err != nil {
			return err
		}
		messages := map[string]*Message{}
		for i := range rows {
			messages[rows[i].ID] = &rows[i]
		}
		var stored []string
		if err := tx.Model(&Event{}).Where("id IN ?", eventIDs).Pluck("id", &stored).Error; err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, id := range stored {
			seen[id] = true
		}
		now := s.Now()
		inserts := []Event{}
		touched, dirty := []*Message{}, map[string]bool{}
		emails := map[string][]string{}
		emits = nil                 // a retried transaction starts again
		notices := []func() error{} // rare: tests and delays, written after the bulk updates
		for i, e := range events {
			f := e.f
			tags, configs := f.Mail.Tags["xem_message"], f.Mail.Tags["ses:configuration-set"]
			if len(tags) != 1 || len(configs) != 1 || f.Mail.MessageID == "" {
				continue
			}
			m := messages[tags[0]]
			if m == nil {
				continue
			}
			// The same ownership checks as ApplyFeedback.
			source, err := address(f.Mail.Source)
			if err != nil || source != m.From || configs[0] != tenant(m.TeamID) || (m.ProviderID != "" && m.ProviderID != f.Mail.MessageID) || m.Status == "QUEUED" {
				continue
			}
			applied[i] = true
			if seen[e.id] {
				continue
			}
			seen[e.id] = true
			inserts = append(inserts, Event{ID: e.id, TeamID: m.TeamID, MessageID: m.ID, Kind: f.EventType, Detail: "SES reported " + f.EventType, CreatedAt: now})
			if name := map[string]string{"Send": "email.sent", "Delivery": "email.delivered"}[f.EventType]; name != "" && !m.IsTest {
				emits = append(emits, pendingEmit{name, &models.EmailEvent{TeamID: m.TeamID, EmailID: m.EmailID, Address: m.Recipients}})
			}
			if !dirty[m.ID] {
				dirty[m.ID] = true
				touched = append(touched, m)
			}
			m.ProviderID = f.Mail.MessageID
			state := fastStates[f.EventType]
			if statusRank[state] < statusRank[m.Status] {
				continue
			}
			m.Status, m.Detail = state, fmt.Sprintf("SES reported %s.", f.EventType)
			if m.IsTest {
				kind := map[string]string{"SENT": "test_accepted", "DELIVERED": "test_delivered", "DELAYED": "test_attention"}[state]
				team, domain := m.TeamID, m.DomainID
				notices = append(notices, func() error { return QueueMilestone(tx, team, domain, kind) })
			}
			if s.NotifyAlerts && !m.IsTest {
				team, id, reason := m.TeamID, m.ID, feedbackReason(f)
				notices = append(notices, func() error { return models.QueueSendingNotice(tx, team, "managed", id, state, reason, now) })
			}
			if m.EmailID != "" {
				emails[m.TeamID] = append(emails[m.TeamID], m.EmailID)
			}
		}
		if len(inserts) == 0 {
			return nil
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&inserts).Error; err != nil {
			return err
		}
		// One UPDATE for every touched message; status and detail are the values the
		// rank rule left in memory, so they never move down.
		var sql strings.Builder
		args := []any{}
		for _, column := range []string{"status", "detail", "provider_id"} {
			sql.WriteString(column + " = CASE id")
			for _, m := range touched {
				value := map[string]string{"status": m.Status, "detail": m.Detail, "provider_id": m.ProviderID}[column]
				sql.WriteString(" WHEN ? THEN ?")
				args = append(args, m.ID, value)
			}
			sql.WriteString(" END, ")
		}
		touchedIDs := []string{}
		for _, m := range touched {
			touchedIDs = append(touchedIDs, m.ID)
		}
		args = append(args, now, touchedIDs)
		if err := tx.Exec("UPDATE managed_messages SET "+sql.String()+"updated_at = ? WHERE id IN ?", args...).Error; err != nil {
			return err
		}
		// Campaign and contact for the emitted events, in one read.
		emitted := []string{}
		for _, p := range emits {
			if p.event.EmailID != "" {
				emitted = append(emitted, p.event.EmailID)
			}
		}
		if len(emitted) > 0 {
			var sources []models.Email
			if err := tx.Select("id", "campaign_id", "contact_id").Where("id IN ?", emitted).Find(&sources).Error; err != nil {
				return err
			}
			byID := map[string]models.Email{}
			for _, e := range sources {
				byID[e.ID] = e
			}
			for _, p := range emits {
				if e, ok := byID[p.event.EmailID]; ok {
					p.event.CampaignID, p.event.ContactID = e.CampaignID, e.ContactID
				}
			}
		}
		for team, emailIDs := range emails {
			if err := tx.Model(&models.Email{}).Where("id IN ? AND team_id = ?", emailIDs, team).Updates(map[string]any{"status": "SENT", "sent_at": now}).Error; err != nil {
				return err
			}
		}
		for _, queue := range notices {
			if err := queue(); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return make([]bool, len(events)), err
	}
	for _, p := range emits {
		busEvents.Emit(p.name, p.event)
	}
	return applied, nil
}
