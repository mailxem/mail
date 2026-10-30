package billing

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	FreePlanID = "xem-free"
	// FreeEmails is the Free plan's monthly cap, matching the payments seed.
	FreeEmails = 3000
)

// ErrFreeLimit is shown on the email and in the sending logs.
var ErrFreeLimit = errors.New("the Free plan's 3,000 emails for this month are used up; upgrade in Settings → Billing to keep sending")

var (
	active  *Reporter
	plans   sync.Map // team ID → cachedPlan
	planTTL = 5 * time.Minute
)

type cachedPlan struct {
	free bool
	at   time.Time
}

// Use makes r the reporter that AllowSend asks about plans.
func Use(r *Reporter) { active = r }

// AllowSend refuses a send that would take a Free workspace past its monthly
// cap. Paid plans are never capped: Dodo bills emails past the allowance. A
// payments outage never blocks sending; the emails are still counted.
func AllowSend(ctx context.Context, team string, n int) error {
	r := active
	if !Enabled || r == nil || team == "" || n <= 0 {
		return nil
	}
	free, err := r.onFreePlan(ctx, team)
	if err != nil || !free {
		return nil
	}
	now := r.Now().UTC()
	month := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	var used int64
	if err := r.DB.WithContext(ctx).Model(&EmailUsage{}).Select("COALESCE(SUM(emails), 0)").Where("team_id = ? AND minute >= ?", team, month).Scan(&used).Error; err != nil {
		return nil
	}
	if used+int64(n) > FreeEmails {
		return ErrFreeLimit
	}
	return nil
}

func (r *Reporter) onFreePlan(ctx context.Context, team string) (bool, error) {
	if v, ok := plans.Load(team); ok && r.Now().Sub(v.(cachedPlan).at) < planTTL {
		return v.(cachedPlan).free, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.URL+"/subscriptions/team/"+url.PathEscape(team), nil)
	if err != nil {
		return false, err
	}
	res, err := r.HTTP.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return false, errors.New(res.Status)
	}
	var sub struct {
		PlanID string `json:"plan_id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(res.Body).Decode(&sub); err != nil {
		return false, err
	}
	// A paid plan counts only while it is in good standing.
	paid := sub.PlanID != "" && sub.PlanID != FreePlanID && (sub.Status == "active" || sub.Status == "pending" || sub.Status == "TRIAL")
	plans.Store(team, cachedPlan{free: !paid, at: r.Now()})
	return !paid, nil
}
