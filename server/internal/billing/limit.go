package billing

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
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
	id      string
	free    bool
	hasPaid bool
	at      time.Time
}

// includedEmails is each paid plan's monthly allowance, matching the payments
// seed. A yearly plan's pool is twelve of these.
var includedEmails = map[string]int64{"xem-starter": 20000, "xem-growth": 100000, "xem-scale": 500000}

// FreeDailyEmails is the Free plan's daily cap, as on the pricing page. A
// paid plan may send 30% of its monthly emails in a day.
const FreeDailyEmails = 100

// PaidCeiling bounds a paid workspace's managed sending at this many times its
// monthly allowance once it has paid. Usage past the allowance is billed; the
// ceiling only limits what an account that stops paying can cost in SES.
const PaidCeiling = 2

// ManagedLimits is a workspace's managed-sending limits for its plan:
//   - Free: 3,000 a month, at most 100 a day.
//   - Paid: 30% of the plan's monthly emails a day. A month allows the plan's
//     included emails before its first payment (a trial, or a checkout not
//     yet charged), since overage cannot be billed yet, and twice them after.
//
// ok is false when billing is off or the plan is unknown, and the account's
// own limits then apply.
func ManagedLimits(ctx context.Context, team string) (daily, monthly int64, ok bool) {
	r := active
	if !Enabled || r == nil || team == "" {
		return 0, 0, false
	}
	plan, err := r.plan(ctx, team)
	if err != nil {
		return 0, 0, false
	}
	if plan.free {
		return FreeDailyEmails, FreeEmails, true
	}
	included := includedEmails[strings.TrimSuffix(plan.id, "-yearly")]
	if included == 0 {
		included = includedEmails["xem-starter"] // plans sold before per-email pricing
	}
	monthly = included
	if plan.hasPaid {
		monthly *= PaidCeiling
	}
	return included * 3 / 10, monthly, true
}

// Use makes r the reporter that AllowSend asks about plans.
func Use(r *Reporter) { active = r }

// AllowSend refuses a send that would take a Free workspace past its monthly
// cap. Paid plans are never capped: Dodo bills emails past the allowance.
// During a payments outage the last known plan applies; a workspace never
// seen before may send, and its emails are still counted.
func AllowSend(ctx context.Context, team string, n int) error {
	r := active
	if !Enabled || r == nil || team == "" || n <= 0 {
		return nil
	}
	plan, err := r.plan(ctx, team)
	if err != nil || !plan.free {
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

func (r *Reporter) plan(ctx context.Context, team string) (cachedPlan, error) {
	cached, known := plans.Load(team)
	if known && r.Now().Sub(cached.(cachedPlan).at) < planTTL {
		return cached.(cachedPlan), nil
	}
	plan, err := r.lookupPlan(ctx, team)
	if err != nil {
		if known {
			return cached.(cachedPlan), nil
		}
		return cachedPlan{}, err
	}
	plan.at = r.Now()
	plans.Store(team, plan)
	return plan, nil
}

func (r *Reporter) lookupPlan(ctx context.Context, team string) (cachedPlan, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.URL+"/subscriptions/team/"+url.PathEscape(team), nil)
	if err != nil {
		return cachedPlan{}, err
	}
	res, err := r.HTTP.Do(req)
	if err != nil {
		return cachedPlan{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return cachedPlan{}, errors.New(res.Status)
	}
	var sub struct {
		PlanID  string `json:"plan_id"`
		Status  string `json:"status"`
		HasPaid bool   `json:"has_paid"`
	}
	if err := json.NewDecoder(res.Body).Decode(&sub); err != nil {
		return cachedPlan{}, err
	}
	// A paid plan counts only once paid and while in good standing; an
	// unfinished checkout stays pending and keeps the Free cap.
	paid := sub.PlanID != "" && sub.PlanID != FreePlanID && sub.Status == "active"
	return cachedPlan{id: sub.PlanID, free: !paid, hasPaid: sub.HasPaid}, nil
}
