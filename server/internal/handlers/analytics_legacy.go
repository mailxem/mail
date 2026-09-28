package handlers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"kori/internal/analytics"
	"kori/internal/api/middleware"
	"kori/internal/models"
	"time"
)

func (h *TrackingHandler) legacyFilter(c echo.Context) (analytics.Filter, error) {
	q := c.Request().URL.Query()
	for _, pair := range [][2]string{{"startDate", "from"}, {"endDate", "to"}, {"startTime", "from"}, {"endTime", "to"}} {
		if q.Get(pair[0]) != "" && q.Get(pair[1]) == "" {
			q.Set(pair[1], q.Get(pair[0]))
		}
	}
	f, err := analytics.Parse(q, middleware.GetTeamID(c), time.Now())
	if err != nil {
		return f, echo.NewHTTPError(400, err.Error())
	}
	if f.Team == "" {
		return f, echo.NewHTTPError(401, "Sign in to view analytics")
	}
	return f, nil
}

// Legacy detail endpoints retain percentage units, but now divide by accepted
// messages in the same send cohort. New clients use the versioned fraction contract.
func (h *TrackingHandler) legacyDetail(c echo.Context, emailID, campaignID string) (EmailAnalytics, error) {
	f, err := h.legacyFilter(c)
	if err != nil {
		return EmailAnalytics{}, err
	}
	messages := func() *gorm.DB {
		q := h.db.WithContext(c.Request().Context()).Model(&models.Email{}).Where("team_id=? AND is_deleted=false AND test=false AND sent_at>=? AND sent_at<? AND status IN ?", f.Team, f.From, f.AsOf, []string{"SENT", "OPENED", "CLICKED", "BOUNCED"}).Where("COALESCE(cc,'')='' AND COALESCE(bcc,'')=''")
		if emailID != "" {
			q = q.Where("id=?", emailID)
		}
		if campaignID != "" {
			q = q.Where("campaign_id=?", campaignID)
		}
		return q
	}
	var count int64
	if err := messages().Count(&count).Error; err != nil {
		return EmailAnalytics{}, err
	}
	var tracking []models.EmailTracking
	err = h.db.WithContext(c.Request().Context()).Model(&models.EmailTracking{}).Where("is_deleted=false AND timestamp>=? AND timestamp<? AND email_id IN (?)", f.From, f.AsOf, messages().Select("id")).Order("timestamp,id").Limit(100001).Find(&tracking).Error
	if err != nil {
		return EmailAnalytics{}, err
	}
	if len(tracking) > 100000 {
		return EmailAnalytics{}, echo.NewHTTPError(400, "Narrow the date range or use the aggregate report")
	}
	result := processEmailAnalytics(tracking, f.Timezone)
	result.AcceptedMessages = count
	if count > 0 {
		result.ClickRate = float64(result.UniqueClicks) / float64(count) * 100
		result.OpenRate = float64(result.UniqueOpens) / float64(count) * 100
	}
	for i := range result.ClickedLinks {
		if count > 0 {
			result.ClickedLinks[i].ClickRate = float64(result.ClickedLinks[i].UniqueClicks) / float64(count) * 100
		}
	}
	c.Response().Header().Set("Deprecation", "true")
	return result, nil
}
