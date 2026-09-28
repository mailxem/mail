package handlers

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"kori/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMailOAuthStateTenantBindingExpiryAndReplay(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.MailOAuthState{}))
	row := models.MailOAuthState{Hash: "state", TeamID: "team", UserID: "user", Verifier: "verifier", ExpiresAt: time.Now().Add(time.Minute)}
	require.NoError(t, db.Create(&row).Error)
	_, err = consumeMailState(db, "state", "other-team", "user")
	require.Error(t, err)
	_, err = consumeMailState(db, "state", "team", "other-user")
	require.Error(t, err)
	state, err := consumeMailState(db, "state", "team", "user")
	require.NoError(t, err)
	require.Equal(t, "verifier", state.Verifier)
	_, err = consumeMailState(db, "state", "team", "user")
	require.Error(t, err)
	row.Hash = "expired"
	row.ExpiresAt = time.Now().Add(-time.Minute)
	require.NoError(t, db.Create(&row).Error)
	_, err = consumeMailState(db, "expired", "team", "user")
	require.Error(t, err)
}

func TestMailConnectionSecretsNeverSerialize(t *testing.T) {
	raw, err := json.Marshal(models.MailConnection{Secret: "oauth-secret", TeamID: "private-team"})
	require.NoError(t, err)
	require.NotContains(t, string(raw), "oauth-secret")
	require.NotContains(t, string(raw), "private-team")
	require.NotContains(t, string(raw), "secret")
}

func TestOnlyWorkspaceAdministratorsManageMailCredentials(t *testing.T) {
	for _, test := range []struct {
		role         string
		key, allowed bool
	}{{string(models.UserRoleAdmin), false, true}, {string(models.UserRoleMember), false, false}, {string(models.UserRoleAdmin), true, false}, {"", false, false}} {
		c := echo.New().NewContext(httptest.NewRequest(http.MethodPost, "/", nil), httptest.NewRecorder())
		c.Set("userID", "user")
		c.Set("teamID", "team")
		c.Set("role", test.role)
		c.Set("isAPIKey", test.key)
		called := false
		err := MailConnectionAdmin(func(echo.Context) error { called = true; return nil })(c)
		require.Equal(t, test.allowed, called)
		if test.allowed {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}

func TestIMAPPaginationAndSearch(t *testing.T) {
	for _, test := range []struct {
		query         string
		limit, offset int
		valid         bool
	}{{"", 20, 0, true}, {"?page=2&limit=10", 10, 20, true}, {"?page=2&limit=10&offset=3", 10, 3, true}, {"?limit=0", 0, 0, false}, {"?limit=101", 0, 0, false}, {"?offset=-1", 0, 0, false}, {"?page=99999999999999", 0, 0, false}} {
		c := echo.New().NewContext(httptest.NewRequest("GET", "/"+test.query, nil), httptest.NewRecorder())
		p, err := parseMailPagination(c)
		if test.valid {
			require.NoError(t, err)
			require.Equal(t, test.limit, p.Limit)
			require.Equal(t, test.offset, p.Offset)
		} else {
			require.Error(t, err)
		}
	}
	c := echo.New().NewContext(httptest.NewRequest("GET", "/?q=receipt", nil), httptest.NewRecorder())
	criteria, err := mailCriteria(c)
	require.NoError(t, err)
	require.Equal(t, []string{"receipt"}, criteria.Text)
	require.Empty(t, criteria.Header)
	require.Empty(t, criteria.Body)
}
