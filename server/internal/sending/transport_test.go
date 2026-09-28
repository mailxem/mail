package sending

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net"
	"net/http/httptest"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kori/internal/models"
)

func TestSMTPRequiresTLSAuthAndDurableAcceptance(t *testing.T) {
	s, team, d, provider := setup(t)
	cred, secret, err := s.CreateCredential(context.Background(), team, d.ID, "Integration test")
	require.NoError(t, err)
	// Use a local test certificate, trusted explicitly by this client only.
	fixture := httptest.NewTLSServer(nil)
	cert := fixture.TLS.Certificates[0]
	fixture.Close()
	dir := t.TempDir()
	s.Config.TLSCert, s.Config.TLSKey = filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem")
	s.Config.SMTPHost = "localhost"
	key, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(s.Config.TLSCert, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]}), 0600))
	require.NoError(t, os.WriteFile(s.Config.TLSKey, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}), 0600))
	server, err := newSMTPServer(s)
	require.NoError(t, err)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	t.Cleanup(func() { _ = server.Close(); <-served })
	conn, err := net.DialTimeout("tcp", listener.Addr().String(), time.Second)
	require.NoError(t, err)
	require.NoError(t, conn.SetDeadline(time.Now().Add(15*time.Second)))
	client, err := smtp.NewClient(conn, "localhost")
	require.NoError(t, err)
	defer client.Close()
	advertised, _ := client.Extension("AUTH")
	require.False(t, advertised, "plaintext must not offer authentication")
	require.Error(t, client.Mail("hello@example.com"), "anonymous relay must be denied")
	roots := x509.NewCertPool()
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	require.NoError(t, err)
	roots.AddCert(leaf)
	require.NoError(t, client.StartTLS(&tls.Config{RootCAs: roots, ServerName: "example.com", MinVersion: tls.VersionTLS12}))
	require.NoError(t, client.Auth(smtp.PlainAuth("", cred.ID, secret, "localhost")))
	require.NoError(t, client.Mail("hello@example.com"))
	require.NoError(t, client.Rcpt("reader@example.net"))
	data, err := client.Data()
	require.NoError(t, err)
	_, err = data.Write(input(team, d).Raw)
	require.NoError(t, err)
	require.NoError(t, data.Close(), "250 only after durable commit")
	var m Message
	require.NoError(t, s.DB.First(&m, "credential_id = ?", cred.ID).Error)
	require.Equal(t, "QUEUED", m.Status)
	require.NotEmpty(t, m.Raw)
	require.Zero(t, provider.sends)
	require.NoError(t, s.ProcessOne(context.Background()))
	require.Equal(t, 1, provider.sends)
	// Revoke after AUTH: an existing connection must not retain send permission.
	require.NoError(t, s.DB.Model(&Credential{}).Where("id = ?", cred.ID).Update("revoked_at", s.Now()).Error)
	require.NoError(t, client.Mail("hello@example.com"))
	require.NoError(t, client.Rcpt("reader@example.net"))
	data, err = client.Data()
	require.NoError(t, err)
	_, err = data.Write(input(team, d).Raw)
	require.NoError(t, err)
	require.Error(t, data.Close())
	var count int64
	require.NoError(t, s.DB.Model(&Message{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func apiContext(team, method, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	r := httptest.NewRequest(method, "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	c.Set("teamID", team)
	c.Set("hasAdminAccess", true)
	c.Set("email", "administrator@example.net")
	return c, w
}
func TestHTTPDomainScopeAndOneTimeCredential(t *testing.T) {
	s, team, d, _ := setup(t)
	other := uuid.NewString()
	for _, handler := range []echo.HandlerFunc{s.checkDomain, s.activateSender, s.removeDomain} {
		c, _ := apiContext(other, "POST", `{"from":"hello@example.com"}`)
		c.SetParamNames("id")
		c.SetParamValues(d.ID)
		require.Error(t, admin(handler)(c))
	}
	c, w := apiContext(team, "POST", `{"domainId":"`+d.ID+`","name":"My app"}`)
	require.NoError(t, admin(s.createCredential)(c))
	require.Equal(t, 201, w.Code)
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	var response struct {
		Credential Credential `json:"credential"`
		Password   string     `json:"password"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	require.NotEmpty(t, response.Password)
	require.NotContains(t, w.Body.String(), `"Hash"`)
	require.NotContains(t, w.Body.String(), `"hash"`)
	c, w = apiContext(team, "GET", "")
	require.NoError(t, admin(s.status)(c))
	require.NotContains(t, w.Body.String(), response.Password)
	c, _ = apiContext(other, "DELETE", "")
	c.SetParamNames("id")
	c.SetParamValues(response.Credential.ID)
	require.Error(t, admin(s.revokeCredential)(c))
	_, err := s.Authenticate(context.Background(), response.Credential.ID, response.Password)
	require.NoError(t, err)
}
func TestRecoveryUpdatesSourceEmailAndExpiresQueuedContent(t *testing.T) {
	for _, state := range []string{"SENDING", "QUEUED"} {
		t.Run(state, func(t *testing.T) {
			s, team, d, p := setup(t)
			email := models.Email{Base: models.Base{ID: uuid.NewString()}, TeamID: team, SMTPConfigID: uuid.NewString(), CategoryID: uuid.NewString(), Status: "DRAFT"}
			require.NoError(t, s.DB.Session(&gorm.Session{SkipHooks: true}).Create(&email).Error)
			in := input(team, d)
			in.EmailID = email.ID
			m, err := s.Submit(context.Background(), in)
			require.NoError(t, err)
			require.NoError(t, s.DB.First(&email, "id = ?", email.ID).Error)
			require.EqualValues(t, "QUEUED", email.Status)
			old := s.Now().Add(-8 * 24 * time.Hour)
			require.NoError(t, s.DB.Model(&m).Updates(map[string]any{"status": state, "created_at": old, "updated_at": old}).Error)
			require.NoError(t, s.MaintainMessages(context.Background()))
			require.NoError(t, s.DB.First(&m, "id = ?", m.ID).Error)
			require.NoError(t, s.DB.First(&email, "id = ?", email.ID).Error)
			expected := "FAILED"
			if state == "SENDING" {
				expected = "DELIVERY_UNKNOWN"
			}
			require.Equal(t, expected, m.Status)
			require.EqualValues(t, expected, email.Status)
			require.Empty(t, m.Raw)
			require.ErrorIs(t, s.ProcessOne(context.Background()), gorm.ErrRecordNotFound)
			require.Zero(t, p.sends)
		})
	}
}

func TestAssistantSendingAccessCannotMintCredentials(t *testing.T) {
	for _, item := range []struct {
		path    string
		write   bool
		allowed bool
	}{
		{"/api/v1/sending", true, true},
		{"/api/v1/sending/pause", true, true},
		{"/api/v1/sending/credentials", true, false},
		{"/api/v1/sending", false, false},
	} {
		c := echo.New().NewContext(httptest.NewRequest("POST", item.path, nil), httptest.NewRecorder())
		c.SetPath(item.path)
		c.Set("teamID", uuid.NewString())
		c.Set("isAPIKey", true)
		c.Set("assistantWrite", item.write)
		called := false
		err := admin(func(echo.Context) error { called = true; return nil })(c)
		require.Equal(t, item.allowed, called)
		if !item.allowed {
			require.Error(t, err)
		}
	}
}
