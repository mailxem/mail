package marketing_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"kori/internal/marketing"
	"kori/internal/models"
)

func TestFormBotsGetThanksButNothingIsSaved(t *testing.T) {
	h := newFormHarness(t)
	f := h.create("optional")
	path := "/public/" + f.Slug
	r := h.call("GET", path, nil)
	require.Equal(t, 200, r.Code)
	var public map[string]any
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), &public))
	require.NotEmpty(t, public["formToken"])
	require.Empty(t, public["turnstileSiteKey"])

	signer := marketing.New(nil, strings.Repeat("s", 32), "")
	submit := func(extra map[string]any) *httptest.ResponseRecorder {
		input := map[string]any{"fields": map[string]string{"email": "reader@example.com", "interest": "personal"}, "version": f.Version, "requestId": uuid.NewString(), "sessionId": uuid.NewString()}
		for k, v := range extra {
			input[k] = v
		}
		return h.call("POST", path, input)
	}
	saved := func() int64 {
		var n int64
		require.NoError(t, h.db.Model(&models.FormSubmission{}).Count(&n).Error)
		return n
	}
	for name, extra := range map[string]map[string]any{
		"honeypot":      {"website": "https://spam.example"},
		"too fast":      {"formToken": signer.FormToken(f.Slug, time.Now())},
		"forged token":  {"formToken": "1700000000000.deadbeef"},
		"missing token": {"formToken": ""},
		"other form":    {"formToken": signer.FormToken("other", time.Now().Add(-time.Minute))},
		"wrong secret":  {"formToken": marketing.New(nil, strings.Repeat("x", 32), "").FormToken(f.Slug, time.Now().Add(-time.Minute))},
	} {
		r := submit(extra)
		require.Equal(t, 200, r.Code, name)
		require.Contains(t, r.Body.String(), `"ok":true`, name)
		require.NotContains(t, r.Body.String(), "submissionId", name)
		require.Zero(t, saved(), name)
	}

	r = submit(map[string]any{"formToken": signer.FormToken(f.Slug, time.Now().Add(-25*time.Hour))})
	require.Equal(t, 400, r.Code)
	require.Contains(t, r.Body.String(), "expired")

	// The token from the page itself works once a person has had time to fill it in.
	time.Sleep(marketing.MinFormFillTime)
	r = submit(map[string]any{"formToken": public["formToken"]})
	require.Equal(t, 200, r.Code, r.Body.String())
	require.Contains(t, r.Body.String(), "submissionId")
	require.EqualValues(t, 1, saved())

	// Copied HTML forms have no render time, so they rely on the honeypot and rate limits.
	values := url.Values{"email": {"html@example.com"}, "interest": {"personal"}, "version": {strconv.Itoa(f.Version)}}
	req := httptest.NewRequest("POST", path, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	h.app.ServeHTTP(rec, req)
	require.Equal(t, 200, rec.Code, rec.Body.String())
	require.EqualValues(t, 2, saved())
}

func TestFormTurnstileIsVerifiedBeforeSaving(t *testing.T) {
	verify := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		ok := r.PostForm.Get("secret") == "secret-key" && r.PostForm.Get("response") == "human"
		_ = json.NewEncoder(w).Encode(map[string]bool{"success": ok})
	}))
	defer verify.Close()
	original := marketing.TurnstileURL
	marketing.TurnstileURL = verify.URL
	defer func() { marketing.TurnstileURL = original }()

	h := newFormHarness(t)
	d := h.definition("optional")
	d["turnstile"] = true
	r := h.call("POST", "/forms", map[string]any{"name": "Guarded", "listId": h.list, "status": "PUBLISHED", "definition": d})
	require.Equal(t, 200, r.Code, r.Body.String())
	var f models.Form
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), &f))
	path := "/public/" + f.Slug
	siteKey := func() any {
		r := h.call("GET", path, nil)
		var public map[string]any
		require.NoError(t, json.Unmarshal(r.Body.Bytes(), &public))
		return public["turnstileSiteKey"]
	}
	submit := func(token string) *httptest.ResponseRecorder {
		return h.call("POST", path, map[string]any{"fields": map[string]string{"email": "reader@example.com", "interest": "personal"}, "version": f.Version, "requestId": uuid.NewString(), "sessionId": uuid.NewString(), "turnstileToken": token})
	}

	// Without server keys the switch does nothing and nobody sees a widget.
	require.Empty(t, siteKey())
	require.Equal(t, 200, submit("").Code)

	t.Setenv("TURNSTILE_SITE_KEY", "site-key")
	t.Setenv("TURNSTILE_SECRET_KEY", "secret-key")
	require.Equal(t, "site-key", siteKey())
	for _, token := range []string{"", "robot"} {
		r := submit(token)
		require.Equal(t, 400, r.Code, r.Body.String())
		require.Contains(t, r.Body.String(), "security check")
	}
	var n int64
	require.NoError(t, h.db.Model(&models.FormSubmission{}).Count(&n).Error)
	require.EqualValues(t, 1, n)
	r = submit("human")
	require.Equal(t, 200, r.Code, r.Body.String())
	require.NoError(t, h.db.Model(&models.FormSubmission{}).Count(&n).Error)
	require.EqualValues(t, 2, n)

	marketing.TurnstileURL = "http://127.0.0.1:1"
	require.Equal(t, 503, submit("human").Code)
}
