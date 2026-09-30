package receiving

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestReceivingAdminMiddlewareDeniesMemberAndAPIKey(t *testing.T) {
	for _, tc := range []struct {
		role string
		api  bool
	}{{"MEMBER", false}, {"ADMIN", true}} {
		e := echo.New()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
		c.Set("teamID", "team")
		c.Set("userID", "user")
		c.Set("role", tc.role)
		c.Set("isAPIKey", tc.api)
		called := false
		err := admin(func(echo.Context) error { called = true; return nil })(c)
		require.Error(t, err)
		require.False(t, called)
	}
}

func TestReceivingStatusDisabledDoesNotAdvertiseStaleReadiness(t *testing.T) {
	service := receivingTestService(t)
	box := addMailbox(t, service, "status@example.com")
	service.Config.Enabled = false
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
	c.Set("teamID", box.TeamID)
	require.NoError(t, service.status(c))
	var body struct {
		Enabled bool `json:"enabled"`
		Domains []struct {
			ReceivingStatus string `json:"receivingStatus"`
		} `json:"domains"`
		Mailboxes []Mailbox `json:"mailboxes"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.False(t, body.Enabled)
	require.Len(t, body.Domains, 1)
	require.Equal(t, "unconfigured", body.Domains[0].ReceivingStatus)
	require.Len(t, body.Mailboxes, 1)
}
