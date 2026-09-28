package sending

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/models"
)

type fakeProvider struct {
	mu         sync.Mutex
	sends      int
	err        error
	identity   Identity
	provisions int
}

func (p *fakeProvider) Provision(context.Context, string, string) error    { p.provisions++; return nil }
func (p *fakeProvider) Identity(context.Context, string) (Identity, error) { return p.identity, nil }
func (p *fakeProvider) Send(context.Context, string, string, string, []string, []byte, string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sends++
	return "ses-message", p.err
}

type fakeDNS map[string][]string

func (d fakeDNS) LookupTXT(_ context.Context, name string) ([]string, error) { return d[name], nil }
func setup(t *testing.T) (*Service, string, Domain, *fakeProvider) {
	t.Helper()
	var dialect gorm.Dialector = sqlite.Open("file:" + uuid.NewString() + "?mode=memory&cache=shared")
	if dsn := os.Getenv("POSTHOOT_TEST_DATABASE_URL"); dsn != "" {
		admin, e := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		require.NoError(t, e)
		schema := "sending_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		require.NoError(t, admin.Exec("CREATE SCHEMA "+schema).Error)
		t.Cleanup(func() { admin.Exec("DROP SCHEMA " + schema + " CASCADE"); raw, _ := admin.DB(); raw.Close() })
		dialect = postgres.Open(dsn + " search_path=" + schema)
	}
	db, e := gorm.Open(dialect, &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, e)
	require.NoError(t, Migrate(db))
	require.NoError(t, db.AutoMigrate(&models.Contact{}, &models.SuppressionList{}, &models.Email{}, &models.SMTPConfig{}, &models.Campaign{}))
	raw, e := db.DB()
	require.NoError(t, e)
	if os.Getenv("POSTHOOT_TEST_DATABASE_URL") == "" {
		raw.SetMaxOpenConns(1)
	}
	t.Cleanup(func() { raw.Close() })
	p := &fakeProvider{identity: Identity{Verified: true, Status: "SUCCESS", DKIM: "SUCCESS", MAILFROM: "SUCCESS", Tokens: []string{"one", "two", "three"}}}
	s := New(db, Config{Enabled: true, SMTPEnabled: true, MaxBytes: 1024 * 1024, DailyLimit: 10, MonthlyLimit: 100, BudgetMicros: 100000, CostMicros: 1000, QueueLimit: 50}, p)
	team := uuid.NewString()
	_, e = s.Account(context.Background(), team)
	require.NoError(t, e)
	require.NoError(t, db.Model(&Account{}).Where("team_id = ?", team).Update("approved", true).Error)
	d, e := s.AddDomain(context.Background(), team, "example.com")
	require.NoError(t, e)
	s.DNS = fakeDNS{"_xem.example.com": {d.Token}, "_dmarc.example.com": {"v=DMARC1; p=none"}}
	d, e = s.RefreshDomain(context.Background(), team, d.ID)
	require.NoError(t, e)
	require.True(t, d.Ready)
	return s, team, d, p
}
func input(team string, d Domain) Submission {
	return Submission{TeamID: team, DomainID: d.ID, Key: uuid.NewString(), From: "hello@example.com", Recipients: []string{"reader@example.net"}, Raw: []byte("From: hello@example.com\r\nTo: reader@example.net\r\nSubject: Hello\r\nContent-Type: text/plain\r\n\r\nHello!")}
}
func TestSubmissionIsolationAndIdempotency(t *testing.T) {
	s, team, d, _ := setup(t)
	ctx := context.Background()
	in := input(team, d)
	first, e := s.Submit(ctx, in)
	require.NoError(t, e)
	second, e := s.Submit(ctx, in)
	require.NoError(t, e)
	require.Equal(t, first.ID, second.ID)
	a, e := s.Account(ctx, team)
	require.NoError(t, e)
	require.EqualValues(t, 1, a.DailyUsed)
	in.Raw = append(in.Raw, '!')
	_, e = s.Submit(ctx, in)
	require.Error(t, e)
	in = input(uuid.NewString(), d)
	_, e = s.Submit(ctx, in)
	require.ErrorIs(t, e, ErrDenied)
	in = input(team, d)
	in.From = "victim@other.com"
	in.Raw = []byte("From: victim@other.com\r\n\r\nhello")
	_, e = s.Submit(ctx, in)
	require.Error(t, e)
}
func TestConcurrentReservationsDoNotExceedQuota(t *testing.T) {
	s, team, d, _ := setup(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	var mu sync.Mutex
	success := 0
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, e := s.Submit(ctx, input(team, d))
			if e == nil {
				mu.Lock()
				success++
				mu.Unlock()
			} else {
				require.ErrorIs(t, e, ErrLimit)
			}
		}()
	}
	wg.Wait()
	require.Equal(t, 10, success)
	a, e := s.Account(ctx, team)
	require.NoError(t, e)
	require.EqualValues(t, 10, a.DailyUsed)
	require.EqualValues(t, 10000, a.BudgetUsedMicros)
}
func TestBudgetAndQueueCaps(t *testing.T) {
	s, team, d, _ := setup(t)
	ctx := context.Background()
	require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Update("monthly_budget_micros", 1000).Error)
	_, e := s.Submit(ctx, input(team, d))
	require.NoError(t, e)
	_, e = s.Submit(ctx, input(team, d))
	require.ErrorIs(t, e, ErrLimit)
	require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Update("monthly_budget_micros", 100000).Error)
	s.Config.QueueLimit = 1
	_, e = s.Submit(ctx, input(team, d))
	require.ErrorIs(t, e, ErrLimit)
}
func TestCredentialExpiryRevocationAndDispatchRecheck(t *testing.T) {
	s, team, d, p := setup(t)
	ctx := context.Background()
	cred, password, e := s.CreateCredential(ctx, team, d.ID, "WordPress")
	require.NoError(t, e)
	require.NotEqual(t, password, cred.Hash)
	encoded, e := json.Marshal(cred)
	require.NoError(t, e)
	require.NotContains(t, string(encoded), cred.Hash)
	_, e = s.Authenticate(ctx, cred.ID, password)
	require.NoError(t, e)
	_, e = s.Authenticate(ctx, cred.ID, "wrong")
	require.ErrorIs(t, e, ErrDenied)
	in := input(team, d)
	in.CredentialID = cred.ID
	_, e = s.Submit(ctx, in)
	require.NoError(t, e)
	require.NoError(t, s.DB.Model(&Credential{}).Where("id = ?", cred.ID).Update("revoked_at", s.Now()).Error)
	_, e = s.Authenticate(ctx, cred.ID, password)
	require.ErrorIs(t, e, ErrDenied)
	require.NoError(t, s.ProcessOne(ctx))
	require.Zero(t, p.sends)
	_, e = s.Submit(ctx, input(team, d))
	require.NoError(t, e)
	require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Update("paused", true).Error)
	require.ErrorIs(t, s.ProcessOne(ctx), gorm.ErrRecordNotFound)
	require.Zero(t, p.sends)
	require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Update("paused", false).Error)
	require.NoError(t, s.ProcessOne(ctx))
	require.Equal(t, 1, p.sends)
}
func TestSuppressionAtSubmissionAndDispatch(t *testing.T) {
	s, team, d, p := setup(t)
	ctx := context.Background()
	in := input(team, d)
	_, e := s.Submit(ctx, in)
	require.NoError(t, e)
	require.NoError(t, s.DB.Create(&Suppression{TeamID: team, Address: in.Recipients[0], Reason: "Complaint"}).Error)
	_, e = s.Submit(ctx, input(team, d))
	require.ErrorIs(t, e, ErrSuppressed)
	require.NoError(t, s.ProcessOne(ctx))
	require.Zero(t, p.sends)
	// Other workspaces' suppression must not leak into this workspace.
	blocked, e := Suppressed(s.DB, uuid.NewString(), in.Recipients[0], true)
	require.NoError(t, e)
	require.False(t, blocked)
}
func TestAmbiguousDeliveryNeverRetries(t *testing.T) {
	s, team, d, p := setup(t)
	p.err = errors.New("connection reset after request")
	ctx := context.Background()
	m, e := s.Submit(ctx, input(team, d))
	require.NoError(t, e)
	require.NoError(t, s.ProcessOne(ctx))
	require.ErrorIs(t, s.ProcessOne(ctx), gorm.ErrRecordNotFound)
	require.Equal(t, 1, p.sends)
	require.NoError(t, s.DB.First(&m, "id = ?", m.ID).Error)
	require.Equal(t, "DELIVERY_UNKNOWN", m.Status)
}
func feedback(m Message, kind string) []byte {
	return []byte(fmt.Sprintf(`{"eventType":%q,"mail":{"messageId":"ses-message","source":%q,"tags":{"xem_message":[%q],"ses:configuration-set":[%q]}},"bounce":{"bounceType":"Permanent","bouncedRecipients":[{"emailAddress":"reader@example.net"}]}}`, kind, m.From, m.ID, tenant(m.TeamID)))
}
func TestFeedbackIdempotencyOrderingAndOwnership(t *testing.T) {
	s, team, d, _ := setup(t)
	ctx := context.Background()
	m, e := s.Submit(ctx, input(team, d))
	require.NoError(t, e)
	require.NoError(t, s.ProcessOne(ctx))
	require.NoError(t, s.ApplyFeedback(ctx, "bounce", feedback(m, "Bounce")))
	require.NoError(t, s.ApplyFeedback(ctx, "bounce", feedback(m, "Bounce")))
	require.NoError(t, s.ApplyFeedback(ctx, "late-delivery", feedback(m, "Delivery")))
	require.NoError(t, s.DB.First(&m, "id = ?", m.ID).Error)
	require.Equal(t, "BOUNCED", m.Status)
	var count int64
	require.NoError(t, s.DB.Model(&Suppression{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
	bad := strings.Replace(string(feedback(m, "Complaint")), tenant(team), tenant(uuid.NewString()), 1)
	require.ErrorIs(t, s.ApplyFeedback(ctx, "wrong-tenant", []byte(bad)), ErrDenied)
}
func TestRawSanitizationAndMarketingHeaders(t *testing.T) {
	raw := []byte("From: hello@example.com\r\nBcc: hidden@example.net\r\nX-SES-TENANT: victim\r\nX-SES-CONFIGURATION-SET: victim\r\nSubject: Hi\r\n\r\nbody")
	clean, _, _, e := NormalizeRaw(raw, "hello@example.com")
	require.NoError(t, e)
	require.NotContains(t, string(clean), "X-SES")
	require.NotContains(t, string(clean), "Bcc:")
	_, _, _, e = NormalizeRaw([]byte("From: a@example.com\r\nFrom: b@example.com\r\n\r\nbody"), "a@example.com")
	require.Error(t, e)
	require.Error(t, validateMarketing(raw))
	require.NoError(t, validateMarketing([]byte("List-Unsubscribe: <https://example.com/unsubscribe>\r\nList-Unsubscribe-Post: List-Unsubscribe=One-Click\r\n\r\nbody")))
}
func TestDomainFailsClosedAndDoesNotResurrectRevocation(t *testing.T) {
	s, team, d, _ := setup(t)
	ctx := context.Background()
	s.DNS = fakeDNS{}
	fresh, e := s.RefreshDomain(ctx, team, d.ID)
	require.NoError(t, e)
	require.False(t, fresh.Ready)
	require.NoError(t, s.DB.Model(&Domain{}).Where("id = ?", d.ID).Update("token", "new-token").Error)
	require.ErrorIs(t, s.persistDomain(ctx, &d), ErrDenied)
	require.False(t, validDMARC([]string{"v=DMARC1; p=none", "v=DMARC1; p=reject"}))
	require.False(t, validDMARC([]string{"v=DMARC1; p=invalid"}))
}
func TestSNSRejectsUntrustedCertificatesBeforeNetwork(t *testing.T) {
	s := New(nil, Config{Enabled: true, Region: "us-east-1", TopicARN: "arn:aws:sns:us-east-1:123456789012:events"}, nil)
	for _, cert := range []string{"http://sns.us-east-1.amazonaws.com/a.pem", "https://sns.us-east-1.amazonaws.com.evil.test/SimpleNotificationService-test.pem", "https://127.0.0.1/a.pem", "https://sns.us-east-1.amazonaws.com:444/SimpleNotificationService-test.pem"} {
		n := Notification{Type: "Notification", TopicARN: s.Config.TopicARN, MessageID: "id", Timestamp: s.Now().Format(time.RFC3339), SigningCertURL: cert}
		require.ErrorIs(t, s.VerifyNotification(context.Background(), n), ErrDenied)
	}
}
func TestSNSSignatureAndTamperDetection(t *testing.T) {
	key, e := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, e)
	n := Notification{Type: "Notification", Message: "{}", MessageID: "id", TopicARN: "topic", Timestamp: "now", SignatureVersion: "2"}
	canonical, e := n.canonical()
	require.NoError(t, e)
	digest := sha256.Sum256([]byte(canonical))
	sig, e := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	require.NoError(t, e)
	n.Signature = base64.StdEncoding.EncodeToString(sig)
	require.NoError(t, verifySignature(n, canonical, &key.PublicKey))
	require.Error(t, verifySignature(n, canonical+"tampered", &key.PublicKey))
}
func TestHTTPDeniesMembersAndAPICredentials(t *testing.T) {
	e := echo.New()
	for _, api := range []bool{false, true} {
		c := e.NewContext(httptest.NewRequest("GET", "/", nil), httptest.NewRecorder())
		c.Set("isAPIKey", api)
		c.Set("hasAdminAccess", api)
		c.Set("teamID", uuid.NewString())
		err := admin(func(echo.Context) error { t.Fatal("unauthorized handler invoked"); return nil })(c)
		require.Error(t, err)
	}
}
func TestCredentialExpires(t *testing.T) {
	s, team, d, _ := setup(t)
	ctx := context.Background()
	cred, password, e := s.CreateCredential(ctx, team, d.ID, "App")
	require.NoError(t, e)
	now := s.Now()
	s.Now = func() time.Time { return now.Add(91 * 24 * time.Hour) }
	_, e = s.Authenticate(ctx, cred.ID, password)
	require.ErrorIs(t, e, ErrDenied)
}
