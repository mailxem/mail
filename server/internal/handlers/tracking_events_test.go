package handlers

import (
	"context"
	"encoding/base64"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/automation"
	"kori/internal/config"
	"kori/internal/events"
	"kori/internal/models"
	"kori/internal/tasks"
)

type recordedAutomations struct {
	mu    sync.Mutex
	tasks []tasks.AutomationExecuteTask
}

func (r *recordedAutomations) EnqueueAutomationTask(_ context.Context, task tasks.AutomationExecuteTask, _ time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks = append(r.tasks, task)
	return nil
}

func (r *recordedAutomations) started(automationID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, task := range r.tasks {
		if task.AutomationID == automationID {
			n++
		}
	}
	return n
}

func TestOpensAndClicksStartAutomations(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	raw.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&models.Contact{}, &models.Campaign{}, &models.Email{}, &models.EmailTracking{}, &models.Automation{}, &models.SMTPConfig{}))
	config.GetConfig().JWT.Secret = "tracking-events-test-secret"

	plain := db.Session(&gorm.Session{SkipHooks: true})
	team, contact := uuid.NewString(), uuid.NewString()
	require.NoError(t, plain.Create(&models.Contact{Base: models.Base{ID: contact}, Email: "ada@example.net", TeamID: team}).Error)
	email := models.Email{Base: models.Base{ID: uuid.NewString()}, TeamID: team, ContactID: contact, From: "news@example.com", To: "ada@example.net", Subject: "Hi", Body: "x", Status: models.EmailStatusSent, SMTPConfigID: uuid.NewString(), CategoryID: uuid.NewString()}
	require.NoError(t, plain.Create(&email).Error)
	opened := models.Automation{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Opened", TriggerEvent: "email.opened", IsActive: true}
	clicked := models.Automation{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: "Clicked", TriggerEvent: "email.clicked", IsActive: true}
	require.NoError(t, plain.Create(&opened).Error)
	require.NoError(t, plain.Create(&clicked).Error)

	queue := &recordedAutomations{}
	require.NoError(t, automation.NewTriggerManager(db, queue).Initialize(context.Background()))

	h := NewTrackingHandler(db)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"mailId": email.ID}).SignedString([]byte(config.GetConfig().JWT.Secret))
	require.NoError(t, err)
	e := echo.New()
	for range 3 {
		require.NoError(t, h.HandleEmailOpen(e.NewContext(httptest.NewRequest("GET", "/t/open?token="+token, nil), httptest.NewRecorder())))
	}
	link := base64.StdEncoding.EncodeToString([]byte("https://example.com/pricing"))
	for range 2 {
		require.NoError(t, h.HandleEmailClick(e.NewContext(httptest.NewRequest("GET", "/t/click/"+link+"?token="+token, nil), httptest.NewRecorder())))
	}

	require.Eventually(t, func() bool { return queue.started(clicked.ID) == 2 }, 2*time.Second, 10*time.Millisecond, "every click starts the click automation")
	require.Eventually(t, func() bool { return queue.started(opened.ID) == 1 }, 2*time.Second, 10*time.Millisecond, "the first open starts the open automation")
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 1, queue.started(opened.ID), "repeat opens do not start it again")
	var opens int64
	require.NoError(t, db.Model(&models.EmailTracking{}).Where("event = ?", "open").Count(&opens).Error)
	require.EqualValues(t, 3, opens, "every open is still recorded")

	unsubscribed := make(chan *models.EmailEvent, 4)
	events.On("email.unsubscribed", func(data interface{}) {
		if ev, ok := data.(*models.EmailEvent); ok && ev.TeamID == team {
			unsubscribed <- ev
		}
	})
	for range 2 { // a mail client's RFC 8058 one-click POST, replayed
		req := httptest.NewRequest("POST", "/t/unsubscribe?token="+token, strings.NewReader("List-Unsubscribe=One-Click"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		require.NoError(t, h.HandleEmailUnsubscribe(e.NewContext(req, httptest.NewRecorder())))
	}
	select {
	case ev := <-unsubscribed:
		require.Equal(t, "one_click", ev.Source)
		require.Equal(t, email.ID, ev.EmailID)
		require.Equal(t, "ada@example.net", ev.Address)
	case <-time.After(2 * time.Second):
		t.Fatal("unsubscribe was not emitted")
	}
	select {
	case <-unsubscribed:
		t.Fatal("an already-unsubscribed contact was reported again")
	case <-time.After(100 * time.Millisecond):
	}
}
