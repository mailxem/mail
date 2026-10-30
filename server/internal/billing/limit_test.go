package billing

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFreePlanIsCappedAndPaidPlansAreNot(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&EmailUsage{}))
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	plansByTeam := map[string]string{"free": `{"plan_id":"xem-free","status":"active"}`, "paid": `{"plan_id":"xem-growth","status":"active"}`, "lapsed": `{"plan_id":"xem-growth","status":"cancelled"}`}
	lookups := 0
	down := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lookups++
		if down {
			w.WriteHeader(502)
			return
		}
		fmt.Fprint(w, plansByTeam[strings.TrimPrefix(r.URL.Path, "/subscriptions/team/")])
	}))
	defer server.Close()
	Enabled = true
	Use(&Reporter{DB: db, URL: server.URL, HTTP: server.Client(), Now: func() time.Time { return now }})
	t.Cleanup(func() { Enabled = false; Use(nil); clearPlans() })

	for _, team := range []string{"free", "paid", "lapsed"} {
		// 2,990 sent earlier this month, 50 last month (not counted).
		require.NoError(t, db.Create(&EmailUsage{TeamID: team, Minute: now.Add(-time.Hour), Emails: 2990}).Error)
		require.NoError(t, db.Create(&EmailUsage{TeamID: team, Minute: now.AddDate(0, -1, 0), Emails: 50}).Error)
	}
	ctx := context.Background()
	require.NoError(t, AllowSend(ctx, "free", 10))
	require.ErrorIs(t, AllowSend(ctx, "free", 11), ErrFreeLimit)
	require.NoError(t, AllowSend(ctx, "paid", 100000))
	require.ErrorIs(t, AllowSend(ctx, "lapsed", 11), ErrFreeLimit)

	// Plans are cached: repeated checks don't call payments again.
	before := lookups
	require.NoError(t, AllowSend(ctx, "paid", 1))
	require.Equal(t, before, lookups)

	// A payments outage never blocks sending.
	down = true
	clearPlans()
	require.NoError(t, AllowSend(ctx, "free", 5000))
}

func clearPlans() { plans.Range(func(k, _ any) bool { plans.Delete(k); return true }) }
