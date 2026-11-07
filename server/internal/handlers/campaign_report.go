package handlers

import (
	"kori/internal/marketing"
	"kori/internal/models"

	"github.com/labstack/echo/v4"
)

// campaignDelivery counts a campaign's emails by where they are now.
type campaignDelivery struct {
	Total      int64 `json:"total"`
	Queued     int64 `json:"queued"`
	Sent       int64 `json:"sent"`
	Delivered  int64 `json:"delivered"`
	Bounced    int64 `json:"bounced"`
	Complained int64 `json:"complained"`
	Failed     int64 `json:"failed"`
	Suppressed int64 `json:"suppressed"`
	Unknown    int64 `json:"unknown"`
}

// campaignReport is what the campaign page shows: where each email is, what
// readers did (machine opens and clicks left out), and how far stages got.
type campaignReport struct {
	Delivery     campaignDelivery     `json:"delivery"`
	Unsubscribed int64                `json:"unsubscribed"`
	Engagement   EmailAnalytics       `json:"engagement"`
	Stages       *marketing.StagePlan `json:"stages,omitempty"`
}

// CampaignReport returns one campaign's delivery, engagement and stages.
func (h *MarketingHandler) CampaignReport(c echo.Context) error {
	var campaign models.Campaign
	if err := h.scoped(c).First(&campaign, "id = ?", c.Param("id")).Error; err != nil {
		return missing()
	}
	db := h.service.DB.WithContext(c.Request().Context())
	ids := db.Model(&models.Email{}).Select("id").Where("campaign_id = ? AND is_deleted = ? AND test = ?", campaign.ID, false, false)
	var r campaignReport
	if err := db.Model(&models.Email{}).Where("campaign_id = ? AND is_deleted = ? AND test = ?", campaign.ID, false, false).Select(`COUNT(*) AS total,
		COALESCE(SUM(CASE WHEN status IN ('PENDING','QUEUED','SENDING','ACCEPTED') THEN 1 ELSE 0 END), 0) AS queued,
		COALESCE(SUM(CASE WHEN status IN ('SENT','OPENED','CLICKED','PARTIAL') THEN 1 ELSE 0 END), 0) AS sent,
		COALESCE(SUM(CASE WHEN status = 'DELIVERED' THEN 1 ELSE 0 END), 0) AS delivered,
		COALESCE(SUM(CASE WHEN status = 'BOUNCED' THEN 1 ELSE 0 END), 0) AS bounced,
		COALESCE(SUM(CASE WHEN status = 'COMPLAINED' THEN 1 ELSE 0 END), 0) AS complained,
		COALESCE(SUM(CASE WHEN status = 'FAILED' THEN 1 ELSE 0 END), 0) AS failed,
		COALESCE(SUM(CASE WHEN status = 'SUPPRESSED' THEN 1 ELSE 0 END), 0) AS suppressed,
		COALESCE(SUM(CASE WHEN status = 'DELIVERY_UNKNOWN' THEN 1 ELSE 0 END), 0) AS unknown`).Scan(&r.Delivery).Error; err != nil {
		return err
	}
	// ponytail: loads every tracking row; page or aggregate in SQL if a campaign passes ~100k events.
	var tracking []models.EmailTracking
	if err := db.Where("is_deleted = ? AND email_id IN (?)", false, ids).Order("timestamp, id").Find(&tracking).Error; err != nil {
		return err
	}
	r.Engagement = processEmailAnalytics(tracking, campaign.Timezone)
	var untracked int64
	if err := db.Model(&models.Email{}).Where("campaign_id = ? AND is_deleted = ? AND test = ? AND click_tracking_enabled = ?", campaign.ID, false, false, false).Count(&untracked).Error; err != nil {
		return err
	}
	reached := r.Delivery.Sent + r.Delivery.Delivered + r.Delivery.Bounced + r.Delivery.Complained
	r.Engagement.AcceptedMessages, r.Engagement.Untracked = reached, untracked
	if tracked := reached - untracked; tracked > 0 {
		r.Engagement.OpenRate = float64(r.Engagement.UniqueOpens) / float64(tracked) * 100
		r.Engagement.ClickRate = float64(r.Engagement.UniqueClicks) / float64(tracked) * 100
		for i := range r.Engagement.ClickedLinks {
			r.Engagement.ClickedLinks[i].ClickRate = float64(r.Engagement.ClickedLinks[i].UniqueClicks) / float64(tracked) * 100
		}
	}
	if r.Engagement.UniqueOpens > 0 {
		r.Engagement.ClickToOpenRate = float64(r.Engagement.UniqueClicks) / float64(r.Engagement.UniqueOpens) * 100
	}
	for _, t := range tracking {
		if t.Event == models.EmailTrackingEventUnsubscribe {
			r.Unsubscribed++
		}
	}
	if campaign.SendInStages != nil && *campaign.SendInStages && campaign.Schedule != models.CampaignScheduleRecurring {
		plan, err := marketing.StageProgress(db, &campaign)
		if err != nil {
			return err
		}
		r.Stages = &plan
	}
	return c.JSON(200, r)
}
