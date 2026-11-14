package sending

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"kori/internal/models"
	"kori/internal/utils/crypto"
)

func TestFeedbackIDNamesSourceTeamAndKind(t *testing.T) {
	s, team, d, _ := setup(t)
	ctx := context.Background()
	in := input(team, d)
	in.Raw = append([]byte("Feedback-ID: forged:"+uuid.NewString()+":campaign:xem\r\n"), in.Raw...)
	in.FeedbackRef, in.FeedbackKind = "campaign-1", "campaign"
	m, err := s.Submit(ctx, in)
	require.NoError(t, err)
	require.Equal(t, 1, strings.Count(string(m.Raw), "Feedback-ID:"), "a sender's own Feedback-ID is replaced")
	require.Contains(t, string(m.Raw), "Feedback-ID: campaign-1:"+team+":campaign:xem\r\n")

	in = input(team, d)
	in.IsTest = true
	m, err = s.Submit(ctx, in)
	require.NoError(t, err)
	require.Contains(t, string(m.Raw), "Feedback-ID: "+d.ID+":"+team+":test:xem\r\n")
}

type fakePostmaster struct {
	stats []map[string]any
	calls int
}

func postmasterSetup(t *testing.T) (*Service, string, *fakePostmaster) {
	t.Helper()
	s, team, _, _ := setup(t)
	s.NotifyAlerts = true
	require.NoError(t, s.DB.AutoMigrate(&models.ServiceNotice{}, &models.ServiceNoticeEvent{}, &models.Alert{}, &models.User{}))
	fake := fakeGoogle(t)
	raw, _ := json.Marshal(oauth2.Token{RefreshToken: "refresh", Expiry: time.Now().Add(-time.Hour)})
	sealed, err := crypto.SealSecret(raw, postmasterAAD)
	require.NoError(t, err)
	require.NoError(t, s.DB.Create(&Postmaster{ID: "google", Secret: sealed}).Error)
	return s, team, fake
}

func fakeGoogle(t *testing.T) *fakePostmaster {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	oldPrivate, oldPublic, oldAPI, oldEndpoint := crypto.PrivateKey, crypto.PublicKey, postmasterAPI, postmasterEndpoint
	t.Cleanup(func() {
		crypto.PrivateKey, crypto.PublicKey, postmasterAPI, postmasterEndpoint = oldPrivate, oldPublic, oldAPI, oldEndpoint
	})
	crypto.PrivateKey, crypto.PublicKey = key, &key.PublicKey
	fake := &fakePostmaster{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, r.ParseForm())
			if r.Form.Get("grant_type") == "authorization_code" {
				require.Equal(t, "good-code", r.Form.Get("code"))
				require.NotEmpty(t, r.Form.Get("code_verifier"))
				fmt.Fprint(w, `{"access_token":"fresh","refresh_token":"refresh","token_type":"Bearer","expires_in":3600}`)
				return
			}
			fmt.Fprint(w, `{"access_token":"fresh","token_type":"Bearer","expires_in":3600}`)
			return
		}
		require.Equal(t, "/domains/mail.xem.test/trafficStats", r.URL.Path)
		require.Equal(t, "Bearer fresh", r.Header.Get("Authorization"))
		fake.calls++
		json.NewEncoder(w).Encode(map[string]any{"trafficStats": fake.stats})
	}))
	t.Cleanup(srv.Close)
	postmasterAPI, postmasterEndpoint = srv.URL, oauth2.Endpoint{TokenURL: srv.URL + "/token"}
	t.Setenv("GOOGLE_POSTMASTER_CLIENT_ID", "client")
	t.Setenv("GOOGLE_POSTMASTER_CLIENT_SECRET", "secret")
	t.Setenv("GOOGLE_POSTMASTER_REDIRECT_URL", "https://api.xem.test/api/v1/sending/postmaster/callback")
	t.Setenv("GOOGLE_POSTMASTER_DOMAIN", "mail.xem.test")
	return fake
}

func loop(id string, ratio float64) map[string]any {
	return map[string]any{"id": id, "spamRatio": ratio}
}

func notices(t *testing.T, s *Service, team, kind string) int64 {
	var n int64
	require.NoError(t, s.DB.Model(&models.ServiceNoticeEvent{}).Where("team_id = ? AND kind = ?", team, kind).Count(&n).Error)
	return n
}

func TestPostmasterSuspendsAtGmailLimitOncePerDay(t *testing.T) {
	s, team, fake := postmasterSetup(t)
	ctx := context.Background()
	fake.stats = []map[string]any{
		{"name": "domains/mail.xem.test/trafficStats/20261103", "spammyFeedbackLoops": []any{}},
		{"name": "domains/mail.xem.test/trafficStats/20261104", "spammyFeedbackLoops": []any{loop(team, 0.004), loop(uuid.NewString(), 0.5), loop("not-a-uuid", 0.5)}},
	}
	require.NoError(t, s.SyncPostmaster(ctx))
	a, err := s.Account(ctx, team)
	require.NoError(t, err)
	require.True(t, a.Suspended)
	var audits int64
	require.NoError(t, s.DB.Model(&Audit{}).Where("team_id = ? AND action = ?", team, "auto_suspend_gmail").Count(&audits).Error)
	require.EqualValues(t, 1, audits)
	require.EqualValues(t, 1, notices(t, s, team, "sending_suspended"))

	// The same day is never acted on twice, even after an operator lifts the suspension.
	require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Update("suspended", false).Error)
	require.NoError(t, s.SyncPostmaster(ctx))
	a, _ = s.Account(ctx, team)
	require.False(t, a.Suspended)
	require.Equal(t, 2, fake.calls)
	var unknown int64
	require.NoError(t, s.DB.Model(&Audit{}).Where("team_id <> ?", team).Count(&unknown).Error)
	require.Zero(t, unknown, "IDs that are not Xem workspaces are ignored")
}

func TestPostmasterWarnsOwnerAndAdminsBelowSuspension(t *testing.T) {
	s, team, fake := postmasterSetup(t)
	ctx := context.Background()
	campaign := models.Campaign{TeamID: team, TemplateID: uuid.NewString(), ListID: uuid.NewString(), SMTPConfigID: uuid.NewString()}
	campaign.ID = uuid.NewString()
	require.NoError(t, s.DB.Create(&campaign).Error)
	admin := models.User{TeamID: uuid.NewString(), Email: "ops@xem.test", Role: models.UserRoleSuperAdmin}
	admin.ID = uuid.NewString()
	require.NoError(t, s.DB.Create(&admin).Error)
	fake.stats = []map[string]any{{"name": "domains/mail.xem.test/trafficStats/20261105", "userReportedSpamRatio": 0.001, "spammyFeedbackLoops": []any{loop(campaign.ID, 0.001)}}}
	require.NoError(t, s.SyncPostmaster(ctx))
	a, _ := s.Account(ctx, team)
	require.False(t, a.Suspended)
	require.EqualValues(t, 1, notices(t, s, team, "gmail_spam_rate"))
	require.EqualValues(t, 1, notices(t, s, admin.TeamID, "gmail_domain_spam_rate"))

	fake.stats[0]["spammyFeedbackLoops"] = []any{loop(team, 0.0005)}
	fake.stats[0]["name"] = "domains/mail.xem.test/trafficStats/20261106"
	fake.stats[0]["userReportedSpamRatio"] = 0.0009
	require.NoError(t, s.SyncPostmaster(ctx))
	require.EqualValues(t, 1, notices(t, s, team, "gmail_spam_rate"), "below 0.08% nobody is alerted")
	require.EqualValues(t, 1, notices(t, s, admin.TeamID, "gmail_domain_spam_rate"))
}

func TestPostmasterDoesNothingUntilConfiguredAndConnected(t *testing.T) {
	s, _, fake := postmasterSetup(t)
	ctx := context.Background()
	t.Setenv("GOOGLE_POSTMASTER_CLIENT_ID", "")
	require.NoError(t, s.SyncPostmaster(ctx))
	require.Zero(t, fake.calls)

	t.Setenv("GOOGLE_POSTMASTER_CLIENT_ID", "client")
	require.NoError(t, s.DB.Where("id = ?", "google").Delete(&Postmaster{}).Error)
	require.NoError(t, s.SyncPostmaster(ctx))
	require.Zero(t, fake.calls)
}

func TestPostmasterConnectIsSuperadminOnlyAndStateIsSingleUse(t *testing.T) {
	s, _, user, token, secret := authenticatedSending(t)
	fakeGoogle(t)
	e := echo.New()
	s.Register(e, secret)
	call := func(method, target string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, target, nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		e.ServeHTTP(w, r)
		return w
	}
	require.Equal(t, http.StatusForbidden, call(http.MethodPost, "/api/v1/sending/postmaster/connect").Code)
	require.NoError(t, s.DB.Model(&user).Update("role", models.UserRoleSuperAdmin).Error)
	w := call(http.MethodPost, "/api/v1/sending/postmaster/connect")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var out struct{ URL string }
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out))
	u, err := url.Parse(out.URL)
	require.NoError(t, err)
	require.Equal(t, postmasterScope, u.Query().Get("scope"))
	state := u.Query().Get("state")

	require.Equal(t, http.StatusBadRequest, call(http.MethodGet, "/api/v1/sending/postmaster/callback?code=good-code&state=wrong").Code)
	require.Equal(t, http.StatusOK, call(http.MethodGet, "/api/v1/sending/postmaster/callback?code=good-code&state="+url.QueryEscape(state)).Code)
	var p Postmaster
	require.NoError(t, s.DB.First(&p, "id = ?", "google").Error)
	require.NotEmpty(t, p.Secret)
	require.NotContains(t, p.Secret, "refresh")
	require.Equal(t, http.StatusBadRequest, call(http.MethodGet, "/api/v1/sending/postmaster/callback?code=good-code&state="+url.QueryEscape(state)).Code)
}
