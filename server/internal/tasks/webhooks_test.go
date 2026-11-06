package tasks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/lib/pq"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/models"
	ulogger "kori/internal/utils/logger"
)

type queuedWebhooks []WebhookDeliveryTask

func (q *queuedWebhooks) EnqueueWebhookDeliveryTask(_ context.Context, task WebhookDeliveryTask) error {
	*q = append(*q, task)
	return nil
}

func webhookDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := db.DB()
	raw.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&models.Webhook{}, &models.Delivery{}, &models.Email{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func addWebhook(t *testing.T, db *gorm.DB, team, url string, active bool, events ...string) models.Webhook {
	t.Helper()
	hook := models.Webhook{Base: models.Base{ID: uuid.NewString()}, Name: "CRM", URL: url, Events: pq.StringArray(events), IsActive: active, Secret: "whsec-0123456789abcdef", TeamID: team}
	if err := db.Create(&hook).Error; err != nil {
		t.Fatal(err)
	}
	return hook
}

func TestEmailEventsQueueSubscribedActiveWebhooks(t *testing.T) {
	db := webhookDB(t)
	team := uuid.NewString()
	opens := addWebhook(t, db, team, "https://hooks.example.com/a", true, "open", "click", "delivered", "unsubscribed")
	addWebhook(t, db, team, "https://hooks.example.com/paused", false, "open", "click", "delivered", "unsubscribed")
	addWebhook(t, db, team, "https://hooks.example.com/bounces", true, "bounce")
	addWebhook(t, db, uuid.NewString(), "https://hooks.example.com/other-team", true, "open")
	email := models.Email{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ContactID: uuid.NewString(), CampaignID: uuid.NewString(), From: "news@example.com", To: "ada@example.net", Subject: "Hi", Body: "x", Status: "SENT", SMTPConfigID: uuid.NewString(), CategoryID: uuid.NewString()}
	if err := db.Session(&gorm.Session{SkipHooks: true}).Create(&email).Error; err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	if err := EnqueueWebhooks(ctx, db, nil, "open", &models.EmailTracking{EmailID: email.ID}); err != nil {
		t.Fatalf("a missing queue is skipped silently, got %v", err)
	}

	var q queuedWebhooks
	click := &models.EmailTracking{EmailID: email.ID, CampaignID: email.CampaignID, ContactID: email.ContactID, URL: "https://example.com/pricing"}
	for _, step := range []struct {
		event string
		data  interface{}
	}{
		{"click", click},
		{"delivered", &models.EmailEvent{TeamID: team, EmailID: email.ID, Address: "ada@example.net"}},
		{"unsubscribed", &models.EmailEvent{TeamID: team, ContactID: email.ContactID, Address: "ada@example.net", Source: "one_click"}},
	} {
		if err := EnqueueWebhooks(ctx, db, &q, step.event, step.data); err != nil {
			t.Fatal(err)
		}
	}
	if len(q) != 3 {
		t.Fatalf("want one delivery per event to the active subscribed webhook, got %d: %+v", len(q), q)
	}
	for _, task := range q {
		if task.WebhookID != opens.ID {
			t.Fatalf("delivery queued for the wrong webhook: %+v", task)
		}
	}
	data := q[0].Payload["data"].(map[string]interface{})
	if q[0].Payload["event"] != "click" || q[0].Payload["teamId"] != team || data["url"] != click.URL || data["email"] != "ada@example.net" || data["campaignId"] != email.CampaignID || data["contactId"] != email.ContactID {
		t.Fatalf("click payload: %+v", q[0].Payload)
	}
	if q[1].Payload["event"] != "delivered" {
		t.Fatalf("delivered payload: %+v", q[1].Payload)
	}
	if got := q[2].Payload["data"].(map[string]interface{})["source"]; got != "one_click" {
		t.Fatalf("unsubscribe source: %v", got)
	}
}

func verify(secret, header string, body []byte, now time.Time) bool {
	var ts, sig string
	for _, part := range strings.Split(header, ",") {
		k, v, _ := strings.Cut(part, "=")
		if k == "t" {
			ts = v
		} else if k == "v1" {
			sig = v
		}
	}
	unix, err := strconv.ParseInt(ts, 10, 64)
	if err != nil || now.Sub(time.Unix(unix, 0)) > 5*time.Minute {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "." + string(body)))
	want, _ := hex.DecodeString(sig)
	return hmac.Equal(mac.Sum(nil), want)
}

func deliveryTask(t *testing.T, hook models.Webhook) *asynq.Task {
	raw, err := json.Marshal(WebhookDeliveryTask{WebhookID: hook.ID, Event: "open", Payload: map[string]interface{}{"event": "open", "teamId": hook.TeamID, "data": map[string]interface{}{"emailId": "e1"}}})
	if err != nil {
		t.Fatal(err)
	}
	return asynq.NewTask(TaskTypeWebhookDelivery, raw)
}

func TestWebhookDeliveryIsSignedRecordedAndRetried(t *testing.T) {
	db := webhookDB(t)
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !verify("whsec-0123456789abcdef", r.Header.Get("X-Xem-Signature"), body, time.Now()) || r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if calls.Add(1) == 1 {
			http.Error(w, "busy", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("thanks"))
	}))
	defer srv.Close()
	old := webhookClient
	webhookClient = srv.Client() // the test endpoint is on loopback, which the real client refuses
	defer func() { webhookClient = old }()

	h := &TaskHandler{db: db, logger: ulogger.New("test")}
	hook := addWebhook(t, db, uuid.NewString(), srv.URL, true, "open")
	if err := h.HandleWebhookDelivery(context.Background(), deliveryTask(t, hook)); err == nil || errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("a 500 must be retried, got %v", err)
	}
	if err := h.HandleWebhookDelivery(context.Background(), deliveryTask(t, hook)); err != nil {
		t.Fatalf("retry should succeed: %v", err)
	}
	var rows []models.Delivery
	db.Order("created_at").Find(&rows)
	if len(rows) != 2 || rows[0].Status != "FAILED" || rows[0].ResponseCode != 500 || rows[1].Status != "SUCCESS" || rows[1].ResponseCode != 200 || rows[1].ResponseBody != "thanks" || rows[1].Event != "open" {
		t.Fatalf("attempts not recorded: %+v", rows)
	}

	paused := addWebhook(t, db, uuid.NewString(), srv.URL, false, "open")
	if err := h.HandleWebhookDelivery(context.Background(), deliveryTask(t, paused)); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("an inactive webhook was called")
	}
}

func TestWebhookDeliveryRefusesInternalAddresses(t *testing.T) {
	db := webhookDB(t)
	var hit atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hit.Store(true) }))
	defer srv.Close()
	h := &TaskHandler{db: db, logger: ulogger.New("test")}
	for _, url := range []string{srv.URL, "http://10.1.2.3/hook", "http://169.254.169.254/latest/meta-data", "http://[::1]:9/hook", "http://localhost:9/hook"} {
		hook := addWebhook(t, db, uuid.NewString(), url, true, "open")
		err := h.HandleWebhookDelivery(context.Background(), deliveryTask(t, hook))
		if !errors.Is(err, errBlockedAddress) || !errors.Is(err, asynq.SkipRetry) {
			t.Fatalf("%s: want a refused, non-retried delivery, got %v", url, err)
		}
	}
	if hit.Load() {
		t.Fatal("a loopback endpoint was reached")
	}
	var failed int64
	db.Model(&models.Delivery{}).Where("status = ?", "FAILED").Count(&failed)
	if failed != 5 {
		t.Fatalf("refused attempts should be recorded, got %d", failed)
	}
}
