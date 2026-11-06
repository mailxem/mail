package tasks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
	"kori/internal/events"
	"kori/internal/models"
	"kori/internal/utils/logger"
)

var webhookLog = logger.New("WEBHOOKS")

// webhookEvents maps event-bus names to the event names webhooks subscribe to.
var webhookEvents = map[string]string{
	"email.opened":       "open",
	"email.clicked":      "click",
	"email.bounced":      "bounce",
	"email.complained":   "complaint",
	"email.sent":         "sent",
	"email.delivered":    "delivered",
	"email.unsubscribed": "unsubscribed",
	"email.failed":       "failed",
}

type WebhookEnqueuer interface {
	EnqueueWebhookDeliveryTask(context.Context, WebhookDeliveryTask) error
}

// ListenWebhookEvents queues a delivery to every subscribed webhook when an email event is emitted.
func ListenWebhookEvents(db *gorm.DB, queue WebhookEnqueuer) {
	for name, event := range webhookEvents {
		events.On(name, func(data interface{}) {
			if err := EnqueueWebhooks(context.Background(), db, queue, event, data); err != nil {
				webhookLog.Warn("webhook fan-out for %s failed: %v", event, err)
			}
		})
	}
}

// EnqueueWebhooks queues one delivery per active team webhook subscribed to event.
// data is the event-bus payload: *models.EmailTracking, *models.Email or *models.EmailEvent.
func EnqueueWebhooks(ctx context.Context, db *gorm.DB, queue WebhookEnqueuer, event string, data interface{}) error {
	if queue == nil {
		return nil
	}
	var ev models.EmailEvent
	url := ""
	switch d := data.(type) {
	case *models.EmailTracking:
		var email models.Email
		if err := db.WithContext(ctx).First(&email, "id = ?", d.EmailID).Error; err != nil {
			return err
		}
		ev = models.EmailEvent{TeamID: email.TeamID, EmailID: d.EmailID, CampaignID: d.CampaignID, ContactID: d.ContactID, Address: email.To}
		url = d.URL
	case *models.Email:
		ev = models.EmailEvent{TeamID: d.TeamID, EmailID: d.ID, CampaignID: d.CampaignID, ContactID: d.ContactID, Address: d.To}
	case *models.EmailEvent:
		ev = *d
	default:
		return nil
	}
	var hooks []models.Webhook
	if err := db.WithContext(ctx).Where("team_id = ? AND is_active = ? AND is_deleted = ?", ev.TeamID, true, false).Find(&hooks).Error; err != nil {
		return err
	}
	payload := map[string]interface{}{
		"event":      event,
		"occurredAt": time.Now().UTC().Format(time.RFC3339),
		"teamId":     ev.TeamID,
		"data": map[string]interface{}{
			"emailId":    ev.EmailID,
			"campaignId": ev.CampaignID,
			"contactId":  ev.ContactID,
			"email":      ev.Address,
		},
	}
	if url != "" {
		payload["data"].(map[string]interface{})["url"] = url
	}
	if ev.Source != "" {
		payload["data"].(map[string]interface{})["source"] = ev.Source
	}
	for _, hook := range hooks {
		if !slices.Contains(hook.Events, event) {
			continue
		}
		if err := queue.EnqueueWebhookDeliveryTask(ctx, WebhookDeliveryTask{WebhookID: hook.ID, Event: event, Payload: payload}); err != nil {
			return err
		}
	}
	return nil
}

var errBlockedAddress = errors.New("webhook URL resolves to a private or reserved address")

var sharedAddressSpace = &net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

// refusePrivate runs after DNS resolution, so a public name pointing at an internal address is refused too.
func refusePrivate(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	ip := net.ParseIP(host)
	if err != nil || ip == nil || ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() || sharedAddressSpace.Contains(ip) {
		return errBlockedAddress
	}
	return nil
}

var webhookClient = &http.Client{
	Timeout:       10 * time.Second,
	CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	Transport: &http.Transport{
		DialContext:         (&net.Dialer{Timeout: 5 * time.Second, Control: refusePrivate}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
	},
}

// SignWebhook returns the X-Xem-Signature value: t=<unix>,v1=<hex HMAC-SHA256 of "<t>.<body>">.
func SignWebhook(secret string, body []byte, at time.Time) string {
	t := strconv.FormatInt(at.Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(t + "."))
	mac.Write(body)
	return "t=" + t + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

// HandleWebhookDelivery posts one signed event and records the attempt; errors are retried by asynq.
func (h *TaskHandler) HandleWebhookDelivery(ctx context.Context, t *asynq.Task) error {
	var task WebhookDeliveryTask
	if err := json.Unmarshal(t.Payload(), &task); err != nil {
		return fmt.Errorf("failed to unmarshal webhook task: %w", asynq.SkipRetry)
	}
	var hook models.Webhook
	if err := h.db.WithContext(ctx).First(&hook, "id = ? AND is_active = ? AND is_deleted = ?", task.WebhookID, true, false).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // removed or paused since the event was queued
		}
		return err
	}
	body, err := json.Marshal(task.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", asynq.SkipRetry)
	}
	delivery := models.Delivery{WebhookID: hook.ID, Event: task.Event, Payload: body, Status: "FAILED"}
	err = postWebhook(ctx, hook, body, &delivery)
	if err != nil {
		delivery.Error = err.Error()
	} else {
		delivery.Status = "SUCCESS"
	}
	if dbErr := h.db.WithContext(ctx).Create(&delivery).Error; dbErr != nil {
		h.logger.Warn("failed to record webhook delivery: %v", dbErr)
	}
	if errors.Is(err, errBlockedAddress) {
		return fmt.Errorf("%w: %w", err, asynq.SkipRetry)
	}
	return err
}

func postWebhook(ctx context.Context, hook models.Webhook, body []byte, delivery *models.Delivery) error {
	if !strings.HasPrefix(hook.URL, "https://") && !strings.HasPrefix(hook.URL, "http://") {
		return errors.New("webhook URL must use http or https")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hook.URL, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Xem-Webhooks/1")
	req.Header.Set("X-Xem-Signature", SignWebhook(hook.Secret, body, time.Now()))
	resp, err := webhookClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	delivery.ResponseCode = resp.StatusCode
	delivery.ResponseBody = string(snippet)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("webhook endpoint returned %d", resp.StatusCode)
	}
	return nil
}
