package sending

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	em "github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kori/internal/api/middleware"
	appdb "kori/internal/db"
	"kori/internal/models"
)

func authenticatedSending(t *testing.T) (*Service, Domain, models.User, string, string) {
	t.Helper()
	s, teamID, domain, _ := setup(t)
	require.NoError(t, s.DB.AutoMigrate(&models.User{}, &models.Team{}, &models.AuthTransaction{}))
	oldDB := appdb.DB
	appdb.DB = s.DB
	t.Cleanup(func() { appdb.DB = oldDB })
	team := models.Team{Base: models.Base{ID: teamID}, Name: "Sending test"}
	user := models.User{Base: models.Base{ID: uuid.NewString()}, TeamID: teamID, Role: models.UserRoleAdmin, Email: "admin@example.test", Password: "unused"}
	require.NoError(t, s.DB.Session(&gorm.Session{SkipHooks: true}).Create(&team).Error)
	require.NoError(t, s.DB.Session(&gorm.Session{SkipHooks: true}).Create(&user).Error)
	secret := "local-sending-auth-regression-test"
	claims := middleware.Claims{UserID: user.ID, TeamID: teamID, Email: user.Email, Scopes: []string{"sending:create"}, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	require.NoError(t, err)
	transaction := models.AuthTransaction{Base: models.Base{ID: uuid.NewString()}, UserID: user.ID, TeamID: teamID, Token: token, Refresh: "unused"}
	require.NoError(t, s.DB.Session(&gorm.Session{SkipHooks: true}).Create(&transaction).Error)
	return s, domain, user, token, secret
}

func TestHTTPDomainCheckWithSessionAndNoBody(t *testing.T) {
	s, domain, _, token, secret := authenticatedSending(t)
	for _, contentType := range []string{"", "application/json"} {
		t.Run(contentType, func(t *testing.T) {
			e := echo.New()
			s.Register(e, secret)
			r := httptest.NewRequest(http.MethodPost, "/api/v1/sending/domains/"+domain.ID+"/check", nil)
			r.Header.Set("Authorization", "Bearer "+token)
			r.Header.Set("Content-Type", contentType)
			w := httptest.NewRecorder()
			e.ServeHTTP(w, r)
			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			var response DomainView
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
			require.Equal(t, domain.ID, response.ID)
			require.True(t, response.Ownership)
			require.True(t, response.Ready)
			require.Greater(t, len(response.Records), 1)
		})
	}
}

func TestHTTPDomainCheckAutomaticallyApprovesAfterDNS(t *testing.T) {
	s, domain, user, token, secret := authenticatedSending(t)
	require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", user.TeamID).Update("approved", false).Error)
	require.NoError(t, s.DB.Model(&Domain{}).Where("id = ?", domain.ID).Updates(map[string]any{"provisioned": false, "ready": false}).Error)
	p := s.Provider.(*fakeProvider)
	p.identity.Verified, p.identity.DKIM = false, "PENDING"
	e := echo.New()
	s.Register(e, secret)
	request := func(method, path string, response any) {
		t.Helper()
		r := httptest.NewRequest(method, path, nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		e.ServeHTTP(w, r)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), response))
	}
	var view DomainView
	var status struct {
		Account Account `json:"account"`
	}
	request(http.MethodPost, "/api/v1/sending/domains/"+domain.ID+"/check", &view)
	require.True(t, view.Provisioned)
	require.False(t, view.Ready)
	require.Len(t, view.Records, 6)
	request(http.MethodGet, "/api/v1/sending", &status)
	require.False(t, status.Account.Approved)
	p.identity.Verified, p.identity.DKIM = true, "SUCCESS"
	request(http.MethodPost, "/api/v1/sending/domains/"+domain.ID+"/check", &view)
	require.True(t, view.Ready)
	request(http.MethodGet, "/api/v1/sending", &status)
	require.True(t, status.Account.Approved)
}

func TestSessionJSONBodyValidationAndWorkspace(t *testing.T) {
	_, _, user, token, secret := authenticatedSending(t)
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"empty", "", 200},
		{"whitespace", " \n\t", 200},
		{"object", `{"name":"domain"}`, 200},
		{"untrusted team", `{"teamId":"another-workspace","count":9007199254740993}`, 200},
		{"malformed", `{"name":`, 400},
		{"null", `null`, 400},
		{"array", `[]`, 400},
		{"scalar", `"domain"`, 400},
		{"multiple objects", `{} {}`, 400},
		{"trailing garbage", `{} invalid`, 400},
	} {
		for _, method := range []string{http.MethodPost, http.MethodPut} {
			t.Run(method+"/"+tc.name, func(t *testing.T) {
				e := echo.New()
				called := false
				e.Add(method, "/payload", func(c echo.Context) error {
					called = true
					require.Equal(t, user.TeamID, c.Get("teamID"))
					raw, err := io.ReadAll(c.Request().Body)
					require.NoError(t, err)
					require.EqualValues(t, len(raw), c.Request().ContentLength)
					if strings.TrimSpace(tc.body) == "" {
						require.Empty(t, raw)
					} else {
						var payload map[string]json.RawMessage
						require.NoError(t, json.Unmarshal(raw, &payload))
						require.JSONEq(t, `"`+user.TeamID+`"`, string(payload["teamId"]))
						if tc.name == "untrusted team" {
							require.Equal(t, "9007199254740993", string(payload["count"]))
						}
					}
					return c.NoContent(http.StatusOK)
				}, middleware.NewAuthMiddleware(secret).Middleware())
				r := httptest.NewRequest(method, "/payload", strings.NewReader(tc.body))
				r.Header.Set("Authorization", "Bearer "+token)
				r.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				e.ServeHTTP(w, r)
				require.Equal(t, tc.status, w.Code, w.Body.String())
				require.Equal(t, tc.status == http.StatusOK, called)
			})
		}
	}
}

func TestSessionRequestBodyTransport(t *testing.T) {
	_, _, _, token, secret := authenticatedSending(t)
	for _, tc := range []struct {
		name, contentType, body string
		status                  int
	}{
		{"empty unknown length", "application/json", "", http.StatusOK},
		{"multipart unchanged", "multipart/form-data; boundary=test", "--test--\r\n", http.StatusOK},
		{"oversized unknown length", "application/json", `{"name":"too large for limit"}`, http.StatusRequestEntityTooLarge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			e.Use(em.BodyLimit("16B"))
			called := false
			e.POST("/payload", func(c echo.Context) error {
				called = true
				body, err := io.ReadAll(c.Request().Body)
				require.NoError(t, err)
				require.Equal(t, tc.body, string(body))
				return c.NoContent(http.StatusOK)
			}, middleware.NewAuthMiddleware(secret).Middleware())
			r := httptest.NewRequest(http.MethodPost, "/payload", strings.NewReader(tc.body))
			r.ContentLength = -1
			r.Header.Set("Authorization", "Bearer "+token)
			r.Header.Set("Content-Type", tc.contentType)
			w := httptest.NewRecorder()
			e.ServeHTTP(w, r)
			require.Equal(t, tc.status, w.Code, w.Body.String())
			require.Equal(t, tc.status == http.StatusOK, called)
		})
	}
}

func TestBodylessDomainCheckStillRequiresWorkspaceAdmin(t *testing.T) {
	s, domain, user, token, secret := authenticatedSending(t)
	foreign := Domain{ID: uuid.NewString(), TeamID: uuid.NewString(), Name: "foreign.example.test", Token: "foreign-ownership-token"}
	require.NoError(t, s.DB.Create(&foreign).Error)
	for _, tc := range []struct {
		name, auth, domainID string
		role                 models.UserRole
		status               int
	}{
		{"anonymous", "", domain.ID, models.UserRoleAdmin, 401},
		{"invalid token", "Bearer invalid", domain.ID, models.UserRoleAdmin, 401},
		{"non-admin", "Bearer " + token, domain.ID, models.UserRoleMember, 403},
		{"foreign domain", "Bearer " + token, foreign.ID, models.UserRoleAdmin, 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, s.DB.Model(&models.User{}).Where("id = ?", user.ID).UpdateColumn("role", tc.role).Error)
			e := echo.New()
			s.Register(e, secret)
			r := httptest.NewRequest(http.MethodPost, "/api/v1/sending/domains/"+tc.domainID+"/check", nil)
			r.Header.Set("Authorization", tc.auth)
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			e.ServeHTTP(w, r)
			require.Equal(t, tc.status, w.Code, w.Body.String())
		})
	}
}
