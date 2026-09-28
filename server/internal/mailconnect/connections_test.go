package mailconnect

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"kori/internal/models"
	"kori/internal/utils/crypto"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestGoogleRefreshPersistsRotationAndDisconnectDeniesReuse(t *testing.T) {
	t.Setenv("GOOGLE_MAIL_CLIENT_ID", "client")
	t.Setenv("GOOGLE_MAIL_CLIENT_SECRET", "client-secret")
	t.Setenv("GOOGLE_MAIL_REDIRECT_URI", "https://xem.example/settings/imap/google/callback")
	oldPrivate, oldPublic, oldTransport := crypto.PrivateKey, crypto.PublicKey, http.DefaultTransport
	t.Cleanup(func() {
		crypto.PrivateKey, crypto.PublicKey, http.DefaultTransport = oldPrivate, oldPublic, oldTransport
	})
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	crypto.PrivateKey, crypto.PublicKey = key, &key.PublicKey
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.MailConnection{}))
	connection := models.MailConnection{Base: models.Base{ID: "connection"}, TeamID: "team", Provider: Google, SMTPConfigID: "sender", Address: "reader@example.com", Active: true}
	require.NoError(t, Seal(&connection, &oauth2.Token{AccessToken: "old-access", RefreshToken: "old-refresh", Expiry: time.Now().Add(-time.Minute)}))
	require.NoError(t, db.Create(&connection).Error)
	calls := 0
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		require.Equal(t, "https://oauth2.googleapis.com/token", r.URL.String())
		require.NoError(t, r.ParseForm())
		require.Equal(t, "old-refresh", r.Form.Get("refresh_token"))
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"access_token":"new-access","refresh_token":"new-refresh","token_type":"Bearer","expires_in":3600}`))}, nil
	})
	token, err := GoogleToken(context.Background(), db, &connection)
	require.NoError(t, err)
	require.Equal(t, "new-access", token)
	var stored models.MailConnection
	require.NoError(t, db.First(&stored, "id = ?", connection.ID).Error)
	var persisted oauth2.Token
	require.NoError(t, Open(&stored, &persisted))
	require.Equal(t, "new-refresh", persisted.RefreshToken)
	raw, err := json.Marshal(stored)
	require.NoError(t, err)
	require.NotContains(t, string(raw), "new-refresh")
	require.NotContains(t, string(raw), "new-access")
	_, err = GoogleToken(context.Background(), db, &connection)
	require.NoError(t, err)
	require.Equal(t, 1, calls)
	require.NoError(t, db.Model(&connection).Updates(map[string]any{"active": false, "secret": ""}).Error)
	_, err = GoogleToken(context.Background(), db, &connection)
	require.Error(t, err)
	_, err = Find(db, "team", "sender", false)
	require.Error(t, err)
	other, err := Find(db, "another-team", "sender", false)
	require.NoError(t, err)
	require.Nil(t, other)
}

func TestGoogleOAuthConfigRequiresSafeRedirect(t *testing.T) {
	t.Setenv("GOOGLE_MAIL_CLIENT_ID", "id")
	t.Setenv("GOOGLE_MAIL_CLIENT_SECRET", "secret")
	for _, raw := range []string{"http://example.com/callback", "https://user:password@example.com/callback", "https://example.com/callback?redirect=elsewhere", ""} {
		t.Setenv("GOOGLE_MAIL_REDIRECT_URI", raw)
		_, err := GoogleConfig()
		require.Error(t, err)
	}
	t.Setenv("GOOGLE_MAIL_REDIRECT_URI", "http://localhost:3131/settings/imap/google/callback")
	_, err := GoogleConfig()
	require.NoError(t, err)
}
