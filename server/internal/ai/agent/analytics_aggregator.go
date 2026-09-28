package agent

import (
	"context"
	"fmt"
	"kori/internal/models"
	"time"

	"gorm.io/gorm"
)

// AnalyticsAggregator aggregates email and campaign performance data
type AnalyticsAggregator struct {
	db *gorm.DB
}

// AnalyticsScope defines the scope for analytics
type AnalyticsScope struct {
	TeamID       string
	CampaignID   string
	AutomationID string
	ContactID    string
	StartDate    time.Time
	EndDate      time.Time
}

// EmailAnalytics represents aggregated email metrics
type EmailAnalytics struct {
	TotalSent     int64
	TotalOpened   int64
	TotalClicked  int64
	TotalBounced  int64
	OpenRate      float64
	ClickRate     float64
	BounceRate    float64
	UniqueOpens   int64
	UniqueClicks  int64
}

// CampaignAnalytics represents campaign-specific analytics
type CampaignAnalytics struct {
	CampaignID      string
	CampaignName    string
	TotalContacts   int64
	EmailsSent      int64
	Opens           int64
	Clicks          int64
	Bounces         int64
	OpenRate        float64
	ClickRate       float64
	TopPerformingTime string
	BestDevice      string
}

// ContactInsights represents insights about a specific contact
type ContactInsights struct {
	ContactID           string
	Email               string
	TotalEmailsReceived int64
	EmailsOpened        int64
	EmailsClicked       int64
	EngagementScore     float64
	LastOpenedAt        *time.Time
	LastClickedAt       *time.Time
	PreferredDevice     string
	PreferredTime       string
	TopInterests        []string
}

// NewAnalyticsAggregator creates a new analytics aggregator
func NewAnalyticsAggregator(db *gorm.DB) *AnalyticsAggregator {
	return &AnalyticsAggregator{db: db}
}

// GetEmailAnalytics aggregates email performance metrics
func (a *AnalyticsAggregator) GetEmailAnalytics(ctx context.Context, scope AnalyticsScope) (*EmailAnalytics, error) {
	// Base query for emails
	query := a.db.WithContext(ctx).Model(&models.Email{})

	// Apply scope filters
	if scope.TeamID != "" {
		query = query.Where("team_id = ?", scope.TeamID)
	}
	if scope.CampaignID != "" {
		query = query.Where("campaign_id = ?", scope.CampaignID)
	}
	if !scope.StartDate.IsZero() {
		query = query.Where("created_at >= ?", scope.StartDate)
	}
	if !scope.EndDate.IsZero() {
		query = query.Where("created_at <= ?", scope.EndDate)
	}

	// Count total sent emails
	var totalSent int64
	if err := query.Count(&totalSent).Error; err != nil {
		return nil, fmt.Errorf("failed to count sent emails: %w", err)
	}

	// Get tracking metrics
	trackingQuery := a.db.WithContext(ctx).Model(&models.EmailTracking{}).
		Joins("JOIN emails ON emails.id = email_trackings.email_id")

	if scope.TeamID != "" {
		trackingQuery = trackingQuery.Where("emails.team_id = ?", scope.TeamID)
	}
	if scope.CampaignID != "" {
		trackingQuery = trackingQuery.Where("email_trackings.campaign_id = ?", scope.CampaignID)
	}
	if !scope.StartDate.IsZero() {
		trackingQuery = trackingQuery.Where("email_trackings.timestamp >= ?", scope.StartDate)
	}
	if !scope.EndDate.IsZero() {
		trackingQuery = trackingQuery.Where("email_trackings.timestamp <= ?", scope.EndDate)
	}
	// Each metric starts from the same scope; event predicates must not accumulate
	// across opens, clicks, bounces, and distinct counts.
	trackingQuery = trackingQuery.Session(&gorm.Session{})

	// Count opens
	var totalOpened int64
	if err := trackingQuery.Where("email_trackings.event = ?", "open").Count(&totalOpened).Error; err != nil {
		return nil, fmt.Errorf("failed to count opens: %w", err)
	}

	// Count clicks
	var totalClicked int64
	if err := trackingQuery.Where("email_trackings.event = ?", "click").Count(&totalClicked).Error; err != nil {
		return nil, fmt.Errorf("failed to count clicks: %w", err)
	}

	// Count bounces
	var totalBounced int64
	if err := trackingQuery.Where("email_trackings.event = ?", "bounce").Count(&totalBounced).Error; err != nil {
		return nil, fmt.Errorf("failed to count bounces: %w", err)
	}

	// Count unique opens
	var uniqueOpens int64
	if err := trackingQuery.Where("email_trackings.event = ?", "open").
		Distinct("email_trackings.email_id").Count(&uniqueOpens).Error; err != nil {
		return nil, fmt.Errorf("failed to count unique opens: %w", err)
	}

	// Count unique clicks
	var uniqueClicks int64
	if err := trackingQuery.Where("email_trackings.event = ?", "click").
		Distinct("email_trackings.email_id").Count(&uniqueClicks).Error; err != nil {
		return nil, fmt.Errorf("failed to count unique clicks: %w", err)
	}

	// Calculate rates
	openRate := 0.0
	clickRate := 0.0
	bounceRate := 0.0

	if totalSent > 0 {
		openRate = float64(uniqueOpens) / float64(totalSent) * 100
		clickRate = float64(uniqueClicks) / float64(totalSent) * 100
		bounceRate = float64(totalBounced) / float64(totalSent) * 100
	}

	return &EmailAnalytics{
		TotalSent:    totalSent,
		TotalOpened:  totalOpened,
		TotalClicked: totalClicked,
		TotalBounced: totalBounced,
		OpenRate:     openRate,
		ClickRate:    clickRate,
		BounceRate:   bounceRate,
		UniqueOpens:  uniqueOpens,
		UniqueClicks: uniqueClicks,
	}, nil
}

// GetCampaignAnalytics gets detailed analytics for a campaign
func (a *AnalyticsAggregator) GetCampaignAnalytics(ctx context.Context, campaignID string) (*CampaignAnalytics, error) {
	// Get campaign
	var campaign models.Campaign
	if err := a.db.WithContext(ctx).Preload("List").First(&campaign, "id = ?", campaignID).Error; err != nil {
		return nil, fmt.Errorf("failed to load campaign: %w", err)
	}

	// Get email analytics for this campaign
	scope := AnalyticsScope{
		TeamID:     campaign.TeamID,
		CampaignID: campaignID,
	}
	emailAnalytics, err := a.GetEmailAnalytics(ctx, scope)
	if err != nil {
		return nil, err
	}

	// Get top performing time
	var topTime struct {
		Hour  int
		Count int64
	}
	a.db.WithContext(ctx).Model(&models.EmailTracking{}).
		Select("EXTRACT(HOUR FROM timestamp) as hour, COUNT(*) as count").
		Where("campaign_id = ? AND event = ?", campaignID, "open").
		Group("hour").
		Order("count DESC").
		Limit(1).
		Scan(&topTime)

	topPerformingTime := fmt.Sprintf("%02d:00", topTime.Hour)

	// Get best device
	var topDevice struct {
		DeviceType string
		Count      int64
	}
	a.db.WithContext(ctx).Model(&models.EmailTracking{}).
		Select("device_type, COUNT(*) as count").
		Where("campaign_id = ? AND event = ?", campaignID, "open").
		Group("device_type").
		Order("count DESC").
		Limit(1).
		Scan(&topDevice)

	return &CampaignAnalytics{
		CampaignID:        campaign.ID,
		CampaignName:      campaign.Name,
		TotalContacts:     int64(campaign.Processed),
		EmailsSent:        emailAnalytics.TotalSent,
		Opens:             emailAnalytics.UniqueOpens,
		Clicks:            emailAnalytics.UniqueClicks,
		Bounces:           emailAnalytics.TotalBounced,
		OpenRate:          emailAnalytics.OpenRate,
		ClickRate:         emailAnalytics.ClickRate,
		TopPerformingTime: topPerformingTime,
		BestDevice:        topDevice.DeviceType,
	}, nil
}

// GetContactInsights gets detailed insights about a contact
func (a *AnalyticsAggregator) GetContactInsights(ctx context.Context, contactID string) (*ContactInsights, error) {
	// Get contact
	var contact models.Contact
	if err := a.db.WithContext(ctx).First(&contact, "id = ?", contactID).Error; err != nil {
		return nil, fmt.Errorf("failed to load contact: %w", err)
	}

	// Count emails received
	var totalEmailsReceived int64
	a.db.WithContext(ctx).Model(&models.Email{}).
		Where("contact_id = ?", contactID).
		Count(&totalEmailsReceived)

	// Count emails opened
	var emailsOpened int64
	a.db.WithContext(ctx).Model(&models.EmailTracking{}).
		Where("contact_id = ? AND event = ?", contactID, "open").
		Distinct("email_id").
		Count(&emailsOpened)

	// Count emails clicked
	var emailsClicked int64
	a.db.WithContext(ctx).Model(&models.EmailTracking{}).
		Where("contact_id = ? AND event = ?", contactID, "click").
		Distinct("email_id").
		Count(&emailsClicked)

	// Calculate engagement score (0-100)
	engagementScore := 0.0
	if totalEmailsReceived > 0 {
		openWeight := 0.6
		clickWeight := 0.4
		engagementScore = (float64(emailsOpened)/float64(totalEmailsReceived))*openWeight*100 +
			(float64(emailsClicked)/float64(totalEmailsReceived))*clickWeight*100
	}

	// Get last opened time
	var lastOpen models.EmailTracking
	err := a.db.WithContext(ctx).
		Where("contact_id = ? AND event = ?", contactID, "open").
		Order("timestamp DESC").
		First(&lastOpen).Error

	var lastOpenedAt *time.Time
	if err == nil {
		lastOpenedAt = &lastOpen.Timestamp
	}

	// Get last clicked time
	var lastClick models.EmailTracking
	err = a.db.WithContext(ctx).
		Where("contact_id = ? AND event = ?", contactID, "click").
		Order("timestamp DESC").
		First(&lastClick).Error

	var lastClickedAt *time.Time
	if err == nil {
		lastClickedAt = &lastClick.Timestamp
	}

	// Get preferred device
	var prefDevice struct {
		DeviceType string
		Count      int64
	}
	a.db.WithContext(ctx).Model(&models.EmailTracking{}).
		Select("device_type, COUNT(*) as count").
		Where("contact_id = ?", contactID).
		Group("device_type").
		Order("count DESC").
		Limit(1).
		Scan(&prefDevice)

	// Get preferred time (hour of day)
	var prefTime struct {
		Hour  int
		Count int64
	}
	a.db.WithContext(ctx).Model(&models.EmailTracking{}).
		Select("EXTRACT(HOUR FROM timestamp) as hour, COUNT(*) as count").
		Where("contact_id = ? AND event = ?", contactID, "open").
		Group("hour").
		Order("count DESC").
		Limit(1).
		Scan(&prefTime)

	preferredTime := fmt.Sprintf("%02d:00", prefTime.Hour)

	return &ContactInsights{
		ContactID:           contact.ID,
		Email:               contact.Email,
		TotalEmailsReceived: totalEmailsReceived,
		EmailsOpened:        emailsOpened,
		EmailsClicked:       emailsClicked,
		EngagementScore:     engagementScore,
		LastOpenedAt:        lastOpenedAt,
		LastClickedAt:       lastClickedAt,
		PreferredDevice:     prefDevice.DeviceType,
		PreferredTime:       preferredTime,
		TopInterests:        []string{}, // Can be extended based on clicked URLs
	}, nil
}
