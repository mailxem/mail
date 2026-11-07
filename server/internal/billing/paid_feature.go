package billing

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrNewsletters is shown when a Free workspace creates or edits a newsletter.
const ErrNewsletters = "Newsletters and newsletter websites come with a paid plan; upgrade in Settings → Billing"

// PaidFeature refuses Free workspaces with 402 and the given message. Like
// PaidOnly, it lets the request through when billing is off or the plan
// can't be looked up.
func PaidFeature(message string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Self-hosted without a license: say how to get one, not how to upgrade.
			if unlicensed() {
				return echo.NewHTTPError(http.StatusPaymentRequired, ErrLicenseFeature)
			}
			team, _ := c.Get("teamID").(string)
			if IsFree(c.Request().Context(), team) {
				return echo.NewHTTPError(http.StatusPaymentRequired, message)
			}
			return next(c)
		}
	}
}
