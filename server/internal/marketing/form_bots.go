package marketing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// A person needs a moment to read and fill a form; scripts submit instantly.
const (
	MinFormFillTime   = 2 * time.Second
	FormTokenLifetime = 24 * time.Hour
)

// TurnstileURL is Cloudflare's siteverify endpoint; tests point it elsewhere.
var TurnstileURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

// FormToken signs the moment a form was shown, so a submission can prove when it started.
func (s *Service) FormToken(slug string, at time.Time) string {
	issued := strconv.FormatInt(at.UnixMilli(), 10)
	mac := hmac.New(sha256.New, []byte(s.Secret))
	mac.Write([]byte("form:" + slug + ":" + issued))
	return issued + "." + hex.EncodeToString(mac.Sum(nil))
}

// FormTokenAge reports how long ago token was issued for slug; ok is false for a missing or forged token.
func (s *Service) FormTokenAge(slug, token string, now time.Time) (time.Duration, bool) {
	issued, _, found := strings.Cut(token, ".")
	ms, err := strconv.ParseInt(issued, 10, 64)
	if !found || err != nil || len(s.Secret) < 32 || !hmac.Equal([]byte(s.FormToken(slug, time.UnixMilli(ms))), []byte(token)) {
		return 0, false
	}
	return now.Sub(time.UnixMilli(ms)), true
}

// TurnstileSiteKey is the public widget key, or empty when Turnstile isn't fully configured.
func TurnstileSiteKey() string {
	if os.Getenv("TURNSTILE_SECRET_KEY") == "" {
		return ""
	}
	return os.Getenv("TURNSTILE_SITE_KEY")
}

// VerifyTurnstile asks Cloudflare whether a widget response came from a person.
func VerifyTurnstile(ctx context.Context, response, ip string) (bool, error) {
	if response == "" || len(response) > 2048 {
		return false, nil
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	form := url.Values{"secret": {os.Getenv("TURNSTILE_SECRET_KEY")}, "response": {response}, "remoteip": {ip}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TurnstileURL, strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	var out struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return false, err
	}
	return out.Success, nil
}
