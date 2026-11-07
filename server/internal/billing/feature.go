package billing

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrPaidFeature is shown when a Free workspace sets up a paid-plan feature.
const ErrPaidFeature = "Cloudflare mailboxes and the Xem inbox come with a paid plan; upgrade in Settings → Billing"

// PaidOnly refuses Free workspaces, and unlicensed self-hosted installs, with
// 402. Like AllowSend, it lets the request through when the plan can't be
// looked up.
func PaidOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if unlicensed() {
			return echo.NewHTTPError(http.StatusPaymentRequired, ErrLicenseFeature)
		}
		team, _ := c.Get("teamID").(string)
		r := active
		if !Enabled || r == nil || team == "" {
			return next(c)
		}
		if plan, err := r.plan(c.Request().Context(), team); err == nil && plan.free {
			return echo.NewHTTPError(http.StatusPaymentRequired, ErrPaidFeature)
		}
		return next(c)
	}
}

// IsFree reports whether a Cloud workspace is on the Free plan, or this is an
// unlicensed self-hosted install. It is false when the plan can't be looked
// up, so callers never treat a paying or unknown workspace as Free.
func IsFree(ctx context.Context, team string) bool {
	if unlicensed() {
		return true
	}
	r := active
	if !Enabled || r == nil || team == "" {
		return false
	}
	plan, err := r.plan(ctx, team)
	return err == nil && plan.free
}
