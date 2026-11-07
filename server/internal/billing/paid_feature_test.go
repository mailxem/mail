package billing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestPaidFeatureRefusesFreeWorkspacesWithItsMessage(t *testing.T) {
	plansByTeam := map[string]string{"free": `{"plan_id":"xem-free","status":"active"}`, "starter": `{"plan_id":"xem-starter","status":"active"}`}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, plansByTeam[strings.TrimPrefix(r.URL.Path, "/subscriptions/team/")])
	}))
	defer server.Close()
	call := func(team string) error {
		c := echo.New().NewContext(httptest.NewRequest(http.MethodPost, "/", nil), httptest.NewRecorder())
		c.Set("teamID", team)
		return PaidFeature(ErrNewsletters)(func(echo.Context) error { return nil })(c)
	}
	// Self-hosted (billing off): everything is allowed.
	require.NoError(t, call("free"))

	Enabled = true
	Use(&Reporter{URL: server.URL, HTTP: server.Client(), Now: time.Now})
	t.Cleanup(func() { Enabled = false; Use(nil); clearPlans() })
	var he *echo.HTTPError
	require.ErrorAs(t, call("free"), &he)
	require.Equal(t, http.StatusPaymentRequired, he.Code)
	require.Equal(t, ErrNewsletters, he.Message)
	require.NoError(t, call("starter"))
}
