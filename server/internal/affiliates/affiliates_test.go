package affiliates

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

	"kori/internal/billing"
	"kori/internal/models"
)

func setup(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Team{}, &models.User{}, &models.Affiliate{}, &models.AffiliateClick{},
		&models.AffiliateReferral{}, &models.AffiliateCommission{}, &models.AffiliatePayout{}))
	return db
}

func addUser(t *testing.T, db *gorm.DB, id, team string, created time.Time) {
	require.NoError(t, db.Exec("INSERT INTO teams (id, name, created_at, updated_at) VALUES (?, ?, ?, ?) ON CONFLICT DO NOTHING", team, "Team "+team, created, created).Error)
	require.NoError(t, db.Exec("INSERT INTO users (id, email, password, role, team_id, created_at, updated_at) VALUES (?, ?, 'x', 'admin', ?, ?, ?)", id, id+"@example.com", team, created, created).Error)
}

func TestCommissionIsThirtyPercentRoundedDown(t *testing.T) {
	require.Equal(t, int64(270), Commission(900))
	require.Equal(t, int64(2), Commission(9))
}

func TestAttributionOnlyForNewWorkspacesAndNeverSelf(t *testing.T) {
	db := setup(t)
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	addUser(t, db, "aff", "aff-team", now.AddDate(0, -3, 0))
	a, err := Join(db, "aff")
	require.NoError(t, err)
	again, err := Join(db, "aff")
	require.NoError(t, err)
	require.Equal(t, a.Code, again.Code, "joining twice keeps the same link")

	addUser(t, db, "new", "new-team", now.Add(-time.Hour))
	addUser(t, db, "old", "old-team", now.AddDate(0, 0, -3))
	addUser(t, db, "mate", "aff-team", now.Add(-time.Hour))

	require.NoError(t, Attribute(db, "new-team", "new", " "+a.Code+" ", now))
	require.NoError(t, Attribute(db, "old-team", "old", a.Code, now))    // too old
	require.NoError(t, Attribute(db, "aff-team", "aff", a.Code, now))    // themselves
	require.NoError(t, Attribute(db, "aff-team", "mate", a.Code, now))   // their own workspace
	require.NoError(t, Attribute(db, "new-team", "new", "unknown", now)) // unknown code
	require.NoError(t, Attribute(db, "new-team", "new", a.Code, now))    // repeat

	var refs []models.AffiliateReferral
	require.NoError(t, db.Find(&refs).Error)
	require.Len(t, refs, 1)
	require.Equal(t, "new-team", refs[0].TeamID)
}

func TestCommissionsUnlockAfterSixtyDaysAndPayOutAboveMinimum(t *testing.T) {
	db := setup(t)
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	addUser(t, db, "aff", "aff-team", start)
	addUser(t, db, "cust", "cust-team", start)
	a, err := Join(db, "aff")
	require.NoError(t, err)
	require.NoError(t, Attribute(db, "cust-team", "cust", a.Code, start))

	// Someone nobody referred earns no one anything.
	require.NoError(t, RecordPayment(db, Payment{ID: "stranger", TeamID: "other", Base: 9900, Currency: "usd", PaidAt: start}))

	// Five monthly $99 payments, one webhook delivered twice.
	for i := range 5 {
		p := Payment{ID: "pay-" + string(rune('a'+i)), TeamID: "cust-team", Base: 9900, Currency: "usd", PaidAt: start.AddDate(0, i, 0)}
		require.NoError(t, RecordPayment(db, p))
		require.NoError(t, RecordPayment(db, p))
	}
	var n int64
	db.Model(&models.AffiliateCommission{}).Count(&n)
	require.Equal(t, int64(5), n)

	// Day 70: only the first payment ($29.70) has unlocked, under the $100 minimum.
	day70 := start.Add(70 * 24 * time.Hour)
	s, err := Summarize(db, a, day70)
	require.NoError(t, err)
	require.Equal(t, int64(2970), s.Totals["unlocked"])
	require.Equal(t, int64(4*2970), s.Totals["pending"])
	require.Equal(t, "Customer 1", s.Commissions[0].Customer)
	require.Equal(t, 1, s.Customers)
	_, err = Pay(db, a.ID, "USD", day70)
	require.ErrorIs(t, err, ErrBelowMinimum)

	// Month 7: all five unlocked, 5 x $29.70 = $148.50.
	month7 := start.AddDate(0, 7, 0)
	due, err := DuePayouts(db, month7)
	require.NoError(t, err)
	require.Equal(t, []Due{{AffiliateID: a.ID, Email: "aff@example.com", Currency: "USD", Amount: 14850}}, due)
	payout, err := Pay(db, a.ID, "usd", month7)
	require.NoError(t, err)
	require.Equal(t, int64(14850), payout.Amount)

	// A paid commission refunded later comes off the next payout.
	require.NoError(t, Reverse(db, "pay-a", month7))
	require.NoError(t, RecordPayment(db, Payment{ID: "big", TeamID: "cust-team", Base: 50000, Currency: "USD", PaidAt: month7}))
	later := month7.Add(Hold)
	due, err = DuePayouts(db, later)
	require.NoError(t, err)
	require.Equal(t, int64(15000-2970), due[0].Amount)
	payout, err = Pay(db, a.ID, "USD", later)
	require.NoError(t, err)
	require.Equal(t, int64(12030), payout.Amount)

	due, err = DuePayouts(db, later)
	require.NoError(t, err)
	require.Empty(t, due, "nothing is paid twice")
	s, err = Summarize(db, a, later)
	require.NoError(t, err)
	require.Equal(t, int64(2970), s.Totals["reversed"])
}

func TestRefundBeforePayoutEarnsNothing(t *testing.T) {
	db := setup(t)
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	addUser(t, db, "aff", "aff-team", start)
	addUser(t, db, "cust", "cust-team", start)
	a, _ := Join(db, "aff")
	require.NoError(t, Attribute(db, "cust-team", "cust", a.Code, start))
	require.NoError(t, RecordPayment(db, Payment{ID: "p", TeamID: "cust-team", Base: 100000, Currency: "USD", PaidAt: start}))
	require.NoError(t, Reverse(db, "p", start.Add(time.Hour)))
	due, err := DuePayouts(db, start.Add(Hold*2))
	require.NoError(t, err)
	require.Empty(t, due)
}

func TestBadgeOnlyForFreeWorkspacesAndCarriesTheOwnersCode(t *testing.T) {
	db := setup(t)
	now := time.Now()
	addUser(t, db, "owner", "free-team", now)
	addUser(t, db, "payer", "paid-team", now)
	require.NoError(t, db.Exec("UPDATE teams SET owner_user_id = 'owner' WHERE id = 'free-team'").Error)

	plans := map[string]string{"free-team": `{"plan_id":"xem-free","status":"active"}`, "paid-team": `{"plan_id":"xem-starter","status":"active"}`}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, plans[strings.TrimPrefix(r.URL.Path, "/subscriptions/team/")])
	}))
	defer srv.Close()
	billing.Enabled = true
	billing.Use(&billing.Reporter{URL: srv.URL, HTTP: srv.Client(), Now: time.Now})
	t.Cleanup(func() { billing.Enabled = false; billing.Use(nil) })

	badge := Badge(context.Background(), db, "free-team")
	a, err := find(db, "user_id = ?", "owner")
	require.NoError(t, err, "the owner becomes an affiliate on first use")
	require.Contains(t, badge, "https://xem.email/?utm_source=email-badge&amp;via="+a.Code)
	require.Contains(t, badge, "Sent with")
	require.Empty(t, Badge(context.Background(), db, "paid-team"))

	require.Equal(t, "<p>Hi</p>"+badge+"</BODY>", WithBadge("<p>Hi</p></BODY>", badge))
	require.Equal(t, "<p>Hi</p>", WithBadge("<p>Hi</p>", ""))

	billing.Enabled = false
	require.Empty(t, Badge(context.Background(), db, "free-team"), "self-hosted installs never get a badge")
}
