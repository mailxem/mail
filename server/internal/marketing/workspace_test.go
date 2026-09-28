package marketing_test

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"kori/internal/api/validator"
	"kori/internal/handlers"
	"kori/internal/models"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWorkspaceSettingsAndLogsIsolation(t *testing.T) {
	db := testDB(t)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.TeamSettings{}, &models.BrandingSettings{}, &models.APIKey{}, &models.APIKeyUsage{}, &models.Webhook{}, &models.Delivery{}, &models.Domain{}))
	owner, other, user := uuid.NewString(), uuid.NewString(), uuid.NewString()
	seed(t, db, &models.Team{Base: models.Base{ID: owner}, Name: "Workspace"})
	seed(t, db, &models.User{Base: models.Base{ID: user}, TeamID: owner, Email: "owner@example.com", Password: "fixture", FirstName: "Original"})
	foreignUser := models.User{Base: models.Base{ID: uuid.NewString()}, TeamID: other, Email: "other@example.com", Password: "fixture", FirstName: "Other"}
	seed(t, db, &foreignUser)
	h := handlers.NewMarketingHandler(db, "", "")
	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error { c.Set("teamID", owner); c.Set("userID", user); return next(c) }
	})
	e.PUT("/profile", h.SaveProfile)
	e.GET("/branding", h.Branding)
	e.PUT("/branding", h.Branding)
	e.GET("/usage", h.APIKeyUsage)
	e.GET("/usage/:id", h.APIKeyUsage)
	e.PUT("/webhooks/:id/status", h.WebhookStatus)
	e.GET("/webhooks/:id/deliveries", h.WebhookDeliveries)
	e.POST("/domains", h.AddDomain)
	e.POST("/domains/:id/verify", h.VerifyDomain)
	call := func(method, path, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		e.ServeHTTP(w, r)
		return w
	}
	require.Equal(t, 204, call("PUT", "/profile", `{"firstName":"Updated","lastName":"Name","bio":"Email editor","id":"`+foreignUser.ID+`"}`).Code)
	var updated models.User
	require.NoError(t, db.First(&updated, "id = ?", user).Error)
	require.Equal(t, "Updated", updated.FirstName)
	require.NoError(t, db.First(&foreignUser, "id = ?", foreignUser.ID).Error)
	require.Equal(t, "Other", foreignUser.FirstName)
	require.Equal(t, 400, call("PUT", "/profile", `{"firstName":" "}`).Code)
	require.Equal(t, 400, call("PUT", "/branding", `{"dashboardName":"Workspace","logoUrl":"javascript:alert(1)"}`).Code)
	result := call("PUT", "/branding", `{"dashboardName":"Xem workspace","logoUrl":"https://example.com/logo.png"}`)
	require.Equal(t, 200, result.Code, result.Body.String())
	require.Contains(t, call("GET", "/branding", "").Body.String(), "Xem workspace")
	for _, team := range []string{owner, other} {
		key := models.APIKey{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Name: team, Key: uuid.NewString()}
		seed(t, db, &key)
		entry := models.APIKeyUsage{Base: models.Base{ID: uuid.NewString()}, APIKeyID: key.ID, Endpoint: "/fixture/" + team, Method: "GET", Timestamp: time.Now().UTC(), Success: true}
		seed(t, db, &entry)
		if team == other {
			require.Equal(t, 404, call("GET", "/usage/"+entry.ID, "").Code)
			require.Equal(t, 404, call("GET", "/usage?api_key_id="+key.ID, "").Code)
		}
	}
	output := call("GET", "/usage?period=24h", "")
	require.Equal(t, 200, output.Code, output.Body.String())
	require.Contains(t, output.Body.String(), "/fixture/"+owner)
	require.NotContains(t, output.Body.String(), "/fixture/"+other)
	require.Equal(t, 400, call("GET", "/usage?period=unknown", "").Code)
	hook := models.Webhook{Base: models.Base{ID: uuid.NewString()}, TeamID: other, Name: "Private", URL: "https://example.com", Secret: "fixture-secret"}
	seed(t, db, &hook)
	require.Equal(t, 404, call("PUT", "/webhooks/"+hook.ID+"/status", `{"isActive":true}`).Code)
	require.Equal(t, 404, call("GET", "/webhooks/"+hook.ID+"/deliveries", "").Code)
	require.Equal(t, 400, call("POST", "/domains", `{"domain":"http://localhost"}`).Code)
	result = call("POST", "/domains", `{"domain":"EXAMPLE.COM","isVerified":true}`)
	require.Equal(t, 201, result.Code, result.Body.String())
	var domain models.Domain
	require.NoError(t, json.Unmarshal(result.Body.Bytes(), &domain))
	require.Equal(t, "example.com", domain.Domain)
	require.False(t, domain.IsVerified)
	require.Equal(t, owner, domain.TeamID)
	require.True(t, strings.HasPrefix(domain.DNSRecord, "xem-verification="))
	require.Equal(t, 409, call("POST", "/domains", `{"domain":"example.com"}`).Code)
	domain.TeamID = other
	require.NoError(t, db.Model(&domain).Update("team_id", other).Error)
	require.Equal(t, 404, call("POST", "/domains/"+domain.ID+"/verify", "").Code)
}

func TestInvitePermissionFailureRollsBackUserCreation(t *testing.T) {
	db := testDB(t)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.TeamInvite{}))
	invite := models.TeamInvite{Base: models.Base{ID: uuid.NewString()}, TeamID: uuid.NewString(), InviterID: uuid.NewString(), Name: "Reader", Email: "reader@example.com", Role: models.UserRoleMember, Code: "fixture-invite", Status: models.InviteStatusPending, ExpiresAt: time.Now().Add(time.Hour)}
	seed(t, db, &invite)
	e := echo.New()
	e.Validator = validator.NewValidator()
	e.POST("/accept/:code", handlers.NewAuthHandler(db).AcceptInvite)
	r := httptest.NewRequest("POST", "/accept/fixture-invite", strings.NewReader(`{"password":"fixture-password"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)
	// Permission tables are deliberately absent to simulate a transactional failure.
	require.Equal(t, 500, w.Code, w.Body.String())
	var count int64
	require.NoError(t, db.Model(&models.User{}).Where("email = ?", invite.Email).Count(&count).Error)
	require.Zero(t, count)
	require.NoError(t, db.First(&invite, "id = ?", invite.ID).Error)
	require.Equal(t, models.InviteStatusPending, invite.Status)
}
