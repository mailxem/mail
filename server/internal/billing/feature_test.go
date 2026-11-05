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

func TestPaidOnlyRefusesFreeWorkspaces(t *testing.T) {
	plansByTeam := map[string]string{"free": `{"plan_id":"xem-free","status":"active"}`, "trial": `{"plan_id":"xem-starter","status":"active"}`, "lapsed": `{"plan_id":"xem-growth","status":"cancelled"}`}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, plansByTeam[strings.TrimPrefix(r.URL.Path, "/subscriptions/team/")])
	}))
	defer server.Close()
	Enabled = true
	Use(&Reporter{URL: server.URL, HTTP: server.Client(), Now: time.Now})
	t.Cleanup(func() { Enabled = false; Use(nil); clearPlans() })

	call := func(team string) error {
		c := echo.New().NewContext(httptest.NewRequest(http.MethodPost, "/", nil), httptest.NewRecorder())
		c.Set("teamID", team)
		return PaidOnly(func(echo.Context) error { return nil })(c)
	}
	for _, team := range []string{"free", "lapsed"} {
		var he *echo.HTTPError
		require.ErrorAs(t, call(team), &he, team)
		require.Equal(t, http.StatusPaymentRequired, he.Code, team)
	}
	require.NoError(t, call("trial"))
}
