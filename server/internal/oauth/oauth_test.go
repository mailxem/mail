package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/assistant"
	"kori/internal/models"
)

func call(t *testing.T, h echo.HandlerFunc, method, target, contentType, body string, user *models.User) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	if user != nil {
		c.Set("userID", user.ID)
		c.Set("teamID", user.TeamID)
	}
	if err := h(c); err != nil {
		he := err.(*echo.HTTPError)
		return he.Code, nil
	}
	out := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func TestSignInFlow(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.APIKey{}, &models.Resource{}, &models.ResourcePermission{}, &models.APIKeyPermission{}))
	require.NoError(t, Migrate(db))
	user := models.User{Email: "owner@example.com", TeamID: uuid.NewString(), Role: models.UserRoleAdmin}
	require.NoError(t, db.Create(&user).Error)

	status, _ := call(t, register(db), "POST", "/", "application/json", `{"client_name":"Claude","redirect_uris":["http://evil.example/cb"]}`, nil)
	require.Equal(t, 400, status, "plain http is only for localhost")
	status, reg := call(t, register(db), "POST", "/", "application/json", `{"client_name":"Claude","redirect_uris":["https://claude.ai/api/mcp/auth_callback"]}`, nil)
	require.Equal(t, 201, status)
	clientID := reg["client_id"].(string)
	redirect := "https://claude.ai/api/mcp/auth_callback"

	verifier := strings.Repeat("v", 50)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	authz := func(redirectURI string) (int, map[string]any) {
		b, _ := json.Marshal(map[string]string{"client_id": clientID, "redirect_uri": redirectURI, "code_challenge": challenge, "code_challenge_method": "S256", "state": "xyz"})
		return call(t, authorize(db), "POST", "/", "application/json", string(b), &user)
	}
	status, _ = authz("https://attacker.example/cb")
	require.Equal(t, 400, status, "unregistered redirect URI")
	status, out := authz(redirect)
	require.Equal(t, 200, status)
	back, err := url.Parse(out["redirect"].(string))
	require.NoError(t, err)
	require.Equal(t, "xyz", back.Query().Get("state"))
	code := back.Query().Get("code")

	form := func(v url.Values) (int, map[string]any) {
		return call(t, token(db), "POST", "/", "application/x-www-form-urlencoded", v.Encode(), nil)
	}
	exchange := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "client_id": {clientID}, "redirect_uri": {redirect}, "code_verifier": {"wrong-verifier-wrong-verifier-wrong-verifier"}}
	status, _ = form(exchange)
	require.Equal(t, 400, status, "PKCE verifier must match")
	exchange.Set("code_verifier", verifier)
	status, tok := form(exchange)
	require.Equal(t, 200, status)
	status, _ = form(exchange)
	require.Equal(t, 400, status, "a code works once")

	access := tok["access_token"].(string)
	var key models.APIKey
	require.NoError(t, db.First(&key, "key = ?", assistant.Lookup(access)).Error)
	require.Equal(t, user.ID, key.AssistantUserID)
	require.True(t, assistant.ValidateActor(db, &key))

	refresh := url.Values{"grant_type": {"refresh_token"}, "refresh_token": {tok["refresh_token"].(string)}, "client_id": {clientID}}
	status, next := form(refresh)
	require.Equal(t, 200, status)
	status, _ = form(refresh)
	require.Equal(t, 400, status, "refresh tokens rotate")
	require.NoError(t, db.First(&key, "id = ?", key.ID).Error)
	require.True(t, key.IsDeleted, "the old access key is retired")

	var g Grant
	require.NoError(t, db.First(&g).Error)
	status, _ = call(t, revoke(db), "DELETE", "/", "", "", &user)
	require.Equal(t, 404, status, "needs the grant ID")
	req := httptest.NewRequest("DELETE", "/", nil)
	c := echo.New().NewContext(req, httptest.NewRecorder())
	c.Set("userID", user.ID)
	c.Set("teamID", user.TeamID)
	c.SetParamNames("id")
	c.SetParamValues(g.ID)
	require.NoError(t, revoke(db)(c))
	refresh.Set("refresh_token", next["refresh_token"].(string))
	status, _ = form(refresh)
	require.Equal(t, 400, status, "a disconnected app cannot refresh")
}
