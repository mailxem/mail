package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/models"
	"kori/internal/receiving"
	"kori/internal/sending"
)

func TestIMAPHandlerDispatchesManagedMailboxAndDeniesOtherTenant(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, receiving.Migrate(db))
	team := uuid.NewString()
	box := receiving.Mailbox{ID: uuid.NewString(), TeamID: team, DomainID: uuid.NewString(), Address: "inbox@example.com", Active: true, SMTPConfigID: uuid.NewString(), Status: "active", UIDValidity: 7, QuotaBytes: 1024, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	require.NoError(t, db.Create(&box).Error)
	require.NoError(t, db.Create(&receiving.Message{ID: uuid.NewString(), MailboxID: box.ID, TeamID: team, ProviderID: "provider", UID: 1, UIDValidity: 7, Folder: "INBOX", Subject: "hello", RawKey: "mail/provider", RawSize: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}).Error)
	service := receiving.New(db, receiving.Config{}, nil)
	receiving.RegisterDefault(service)
	t.Cleanup(func() { receiving.RegisterDefault(nil) })
	handler := NewIMAPHandler(db)
	request := func(teamID string) (int, error) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/?config_id="+box.ID+"&folder=INBOX", nil)
		c := echo.New().NewContext(req, rec)
		c.Set("teamID", teamID)
		err := handler.GetEmails(c)
		return rec.Code, err
	}
	code, err := request(team)
	require.NoError(t, err)
	require.Equal(t, 200, code)
	_, err = request(uuid.NewString())
	require.Error(t, err)
}

func TestManagedSenderProjectionsRequireReadyMailboxAndAccount(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.IMAPConfig{}, &models.MailConnection{}, &models.CloudflareRelay{}, &models.SMTPConfig{}, &sending.Account{}, &sending.Domain{}))
	require.NoError(t, receiving.Migrate(db))
	team, domainID := uuid.NewString(), uuid.NewString()
	checked := time.Now()
	require.NoError(t, db.Create(&sending.Account{TeamID: team, Approved: true}).Error)
	require.NoError(t, db.Create(&sending.Domain{ID: domainID, TeamID: team, Name: "example.com", Ownership: true, Provisioned: true, Ready: true, CheckedAt: &checked}).Error)
	statuses := []string{"active", "error", "storage_full"}
	ids := map[string]string{}
	for _, status := range statuses {
		senderID := uuid.NewString()
		ids[status] = senderID
		sender := models.SMTPConfig{Base: models.Base{ID: senderID}, TeamID: team, Provider: "MANAGED", Host: "managed.internal", Port: 587, Username: status + "@example.com", FromEmail: status + "@example.com", IsActive: true}
		require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&sender).Error)
		box := receiving.Mailbox{ID: uuid.NewString(), TeamID: team, DomainID: domainID, Address: status + "@example.com", Active: true, SMTPConfigID: senderID, Status: status, UIDValidity: 1, QuotaBytes: 1024}
		require.NoError(t, db.Create(&box).Error)
	}
	call := func(method string) []map[string]any {
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
		c.Set("teamID", team)
		if method == "mailboxes" {
			require.NoError(t, (&MailConnectionsHandler{DB: db}).Mailboxes(c))
		} else {
			require.NoError(t, (&MailConnectionsHandler{DB: db}).Senders(c))
		}
		var rows []map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &rows))
		return rows
	}
	mailboxes := call("mailboxes")
	for _, row := range mailboxes {
		status := row["username"].(string)
		if status == "active@example.com" {
			require.Equal(t, ids["active"], row["smtpConfigId"])
		} else {
			_, exposed := row["smtpConfigId"]
			require.False(t, exposed)
		}
	}
	senders := call("senders")
	require.Len(t, senders, 1)
	require.Equal(t, ids["active"], senders[0]["id"])
	require.NoError(t, db.Model(&sending.Account{}).Where("team_id = ?", team).Update("paused", true).Error)
	for _, row := range call("mailboxes") {
		_, exposed := row["smtpConfigId"]
		require.False(t, exposed)
	}
	require.Empty(t, call("senders"))
}
