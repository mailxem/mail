package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"kori/internal/analytics"
	"kori/internal/config"
	"kori/internal/models"
	"kori/internal/utils"
	"kori/internal/utils/logger"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/mssola/user_agent"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

var trackingLog = logger.New("TRACKING_HANDLER")

// 🔍 TrackingHandler handles email tracking events
type TrackingHandler struct {
	db *gorm.DB
}

// 🆕 NewTrackingHandler creates a new tracking handler
func NewTrackingHandler(db *gorm.DB) *TrackingHandler {
	return &TrackingHandler{db: db}
}

// 📊 createTrackingEntry creates a new tracking entry with device and location info
// @Summary Create a new tracking entry
// @Description Create a new tracking entry with device and location info
// @Accept json
// @Produce json
// @Param emailID path string true "Email ID"
// @Param event path string true "Event"
// @Param url path string true "URL"
// @Success 200 {object} models.EmailTracking "Tracking entry created successfully"
// @Failure 400 {object} map[string]string "Validation error or email not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/t [post]
func (h *TrackingHandler) createTrackingEntry(c echo.Context, emailID string, event models.EmailTrackingEvent, url string) (*models.EmailTracking, error) {
	// Get email details
	var email models.Email
	if err := h.db.Preload("Campaign").First(&email, "id = ?", emailID).Error; err != nil {
		return nil, err
	}

	// Parse User-Agent
	ua := user_agent.New(c.Request().UserAgent())
	deviceType := "other"
	if ua.Mobile() {
		deviceType = "mobile"
	} else if !ua.Mobile() {
		// If not mobile, assume desktop for now
		deviceType = "desktop"
	}

	// Get browser and OS info
	browser, version := ua.Browser()

	// Create tracking entry
	tracking := &models.EmailTracking{
		EmailID:    emailID,
		Event:      event,
		Timestamp:  time.Now(),
		CampaignID: email.CampaignID,
		ContactID:  email.ContactID,
		IPAddress:  utils.GetIPAddress(c.Request()),
		UserAgent:  c.Request().UserAgent(),
		DeviceType: deviceType,
		Browser:    browser + " " + version,
		OS:         ua.OS(),
		URL:        url,
	}

	// Get geolocation data
	geoData, err := utils.GetGeolocationData(tracking.IPAddress)
	if err == nil {
		tracking.Country = geoData.Country
		tracking.City = geoData.City
		tracking.Region = geoData.Region
	}

	// Save to database
	if err := h.db.Create(tracking).Error; err != nil {
		return nil, err
	}

	return tracking, nil
}

// 🖱️ HandleClick handles click tracking
// @Summary Handle click tracking
// @Description Handle click tracking
// @Accept json
// @Produce json
// @Param token query string true "Token"
// @Success 200 {object} models.EmailTracking "Tracking entry created successfully"
// @Failure 400 {object} map[string]string "Validation error or token missing"
// @Failure 401 {object} map[string]string "Invalid token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/t/click [get]
func (h *TrackingHandler) HandleEmailClick(c echo.Context) error {
	// Extract token from query params
	token := c.QueryParam("token")
	if token == "" {
		return c.String(http.StatusBadRequest, "Missing token")
	}

	// Parse JWT token
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetConfig().JWT.Secret), nil
	})

	if err != nil {
		return c.String(http.StatusUnauthorized, "Invalid token")
	}

	// Extract email ID from claims
	emailID, ok := claims["mailId"].(string)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid token claims")
	}

	// Get the original URL from the path
	originalURL := strings.TrimPrefix(c.Request().URL.Path, "/t/click/")

	decodedURL, err := base64.StdEncoding.DecodeString(originalURL)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid URL")
	}

	// Create tracking entry
	_, err = h.createTrackingEntry(c, emailID, models.EmailTrackingEventClick, string(decodedURL))
	if err != nil {
		trackingLog.Error("Failed to create click tracking entry", err)
		// Still redirect even if tracking fails
	}

	// Redirect to original URL
	return c.Redirect(http.StatusFound, string(decodedURL))
}

// 👁️ HandleOpen handles open tracking
// @Summary Handle open tracking
// @Description Handle open tracking
// @Accept json
// @Produce json
// @Param token query string true "Token"
// @Success 200 {object} models.EmailTracking "Tracking entry created successfully"
// @Failure 400 {object} map[string]string "Validation error or token missing"
// @Failure 401 {object} map[string]string "Invalid token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/t/open [get]
func (h *TrackingHandler) HandleEmailOpen(c echo.Context) error {
	// Extract token from query params
	token := c.QueryParam("token")
	if token == "" {
		return c.String(http.StatusBadRequest, "Missing token")
	}

	// Parse JWT token
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetConfig().JWT.Secret), nil
	})

	if err != nil {
		return c.String(http.StatusUnauthorized, "Invalid token")
	}

	// Extract email ID from claims
	emailID, ok := claims["mailId"].(string)
	if !ok {
		return c.String(http.StatusBadRequest, "Invalid token claims")
	}

	// Create tracking entry
	_, err = h.createTrackingEntry(c, emailID, models.EmailTrackingEventOpen, "")
	if err != nil {
		trackingLog.Error("Failed to create open tracking entry", err)
	}

	// Return a 1x1 transparent GIF
	c.Response().Header().Set("Content-Type", "image/gif")
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	return c.Blob(http.StatusOK, "image/gif", utils.TransparentGIF())
}

// 📈 GetEmailAnalytics returns analytics for a specific email
// @Summary Get email analytics
// @Description Get email analytics
// @Accept json
// @Produce json
// @Param emailId query string true "Email ID"
// @Success 200 {object} EmailAnalytics "Email analytics"
// @Failure 400 {object} map[string]string "Validation error or email not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/analytics/email [get]
func (h *TrackingHandler) GetEmailAnalytics(c echo.Context) error {
	id := c.QueryParam("emailId")
	if id == "" {
		return echo.NewHTTPError(400, "Missing emailId")
	}
	result, err := h.legacyDetail(c, id, "")
	if err != nil {
		return err
	}
	return c.JSON(200, result)
}

// 📊 GetCampaignAnalytics returns analytics for a campaign
// @Summary Get campaign analytics
// @Description Get campaign analytics
// @Accept json
// @Produce json
// @Param campaignId query string true "Campaign ID"
// @Success 200 {object} EmailAnalytics "Campaign analytics"
// @Failure 400 {object} map[string]string "Validation error or campaign not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/analytics/campaign [get]
func (h *TrackingHandler) GetCampaignAnalytics(c echo.Context) error {
	id := c.QueryParam("campaignId")
	if id == "" {
		return echo.NewHTTPError(400, "Missing campaignId")
	}
	result, err := h.legacyDetail(c, "", id)
	if err != nil {
		return err
	}
	return c.JSON(200, result)
}

// Analytics response structures
// 📊 EmailAnalytics represents email analytics data
// @Description Email analytics data
type EmailAnalytics struct {
	AcceptedMessages int64   `json:"acceptedMessages"`
	ClickToOpenRate  float64 `json:"clickToOpenRate"`
	// 📊 Basic Metrics
	OpenCount      int     `json:"openCount"`
	ClickCount     int     `json:"clickCount"`
	UniqueOpens    int     `json:"uniqueOpens"`
	UniqueClicks   int     `json:"uniqueClicks"`
	ClickRate      float64 `json:"clickRate"`
	OpenRate       float64 `json:"openRate"`
	BounceCount    int     `json:"bounceCount"`
	ComplaintCount int     `json:"complaintCount"`

	// 📱 Device & Browser Analytics
	DeviceBreakdown  map[string]int `json:"deviceBreakdown"`
	BrowserBreakdown map[string]int `json:"browserBreakdown"`
	OSBreakdown      map[string]int `json:"osBreakdown"`

	// 🌍 Geographic Analytics
	GeoBreakdown    map[string]int `json:"geoBreakdown"`
	CityBreakdown   map[string]int `json:"cityBreakdown"`
	RegionBreakdown map[string]int `json:"regionBreakdown"`

	// 🔗 Link Analytics
	ClickedLinks []LinkAnalytics `json:"clickedLinks"`

	// ⏰ Time-based Analytics
	TimelineData       []TimelineDataPoint `json:"timelineData"`
	HourlyBreakdown    map[int]int         `json:"hourlyBreakdown"`    // Hour (0-23) -> count
	DayOfWeekBreakdown map[string]int      `json:"dayOfWeekBreakdown"` // Day name -> count

	// 🎯 Engagement Metrics
	EngagementScore float64       `json:"engagementScore"`
	FirstOpenTime   time.Duration `json:"firstOpenTime"`   // Time to first open
	AverageReadTime time.Duration `json:"averageReadTime"` // Average time between open and click

	// 📈 Comparative Metrics
	IndustryAvgOpenRate  float64 `json:"industryAvgOpenRate,omitempty"`
	IndustryAvgClickRate float64 `json:"industryAvgClickRate,omitempty"`

	// 🔄 Retention Metrics
	RepeatOpens  int `json:"repeatOpens"`  // Number of times same user opened
	RepeatClicks int `json:"repeatClicks"` // Number of times same user clicked
}

type LinkAnalytics struct {
	URL                string         `json:"url"`
	ClickCount         int            `json:"clickCount"`
	UniqueClicks       int            `json:"uniqueClicks"`
	ClickRate          float64        `json:"clickRate"`
	FirstClickTime     string         `json:"firstClickTime"`
	LastClickTime      string         `json:"lastClickTime"`
	AverageTimeToClick time.Duration  `json:"averageTimeToClick"`
	DeviceBreakdown    map[string]int `json:"deviceBreakdown"`
}

type TimelineDataPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"eventType"`
	Count       int       `json:"count"`
	UniqueCount int       `json:"uniqueCount"`
	DeviceType  string    `json:"deviceType,omitempty"`
	Country     string    `json:"country,omitempty"`
	Browser     string    `json:"browser,omitempty"`
}

// 📊 processEmailAnalytics processes email analytics data
// @Description Process email analytics data
func processEmailAnalytics(tracking []models.EmailTracking, timeZone string) EmailAnalytics {
	// Parse timezone or default to UTC
	location, err := time.LoadLocation(timeZone)
	if err != nil {
		location = time.UTC
	}

	analytics := EmailAnalytics{
		DeviceBreakdown:    make(map[string]int),
		BrowserBreakdown:   make(map[string]int),
		OSBreakdown:        make(map[string]int),
		GeoBreakdown:       make(map[string]int),
		CityBreakdown:      make(map[string]int),
		RegionBreakdown:    make(map[string]int),
		HourlyBreakdown:    make(map[int]int),
		DayOfWeekBreakdown: make(map[string]int),
	}

	uniqueOpens := make(map[string]bool)
	uniqueClicks := make(map[string]bool)
	clickedLinks := make(map[string]*LinkAnalytics)
	linkClickers := map[string]map[string]bool{}
	userOpenTimes := make(map[string][]time.Time)
	userClickTimes := make(map[string][]time.Time)

	var firstOpenTime time.Time
	totalReadTime := time.Duration(0)
	readTimeCount := 0

	for _, t := range tracking {
		// Convert timestamp to user's timezone
		localTime := t.Timestamp.In(location)

		// Update hourly and daily breakdowns
		analytics.HourlyBreakdown[localTime.Hour()]++
		analytics.DayOfWeekBreakdown[localTime.Weekday().String()]++

		// Device & Browser analytics
		analytics.DeviceBreakdown[t.DeviceType]++
		analytics.BrowserBreakdown[t.Browser]++
		analytics.OSBreakdown[t.OS]++

		// Geographic analytics
		if t.Country != "" {
			analytics.GeoBreakdown[t.Country]++
		}
		if t.City != "" {
			analytics.CityBreakdown[t.City]++
		}
		if t.Region != "" {
			analytics.RegionBreakdown[t.Region]++
		}

		switch t.Event {
		case models.EmailTrackingEventOpen:
			analytics.OpenCount++
			uniqueOpens[t.EmailID] = true
			userOpenTimes[t.EmailID] = append(userOpenTimes[t.EmailID], t.Timestamp)

			if firstOpenTime.IsZero() || t.Timestamp.Before(firstOpenTime) {
				firstOpenTime = t.Timestamp
			}

		case models.EmailTrackingEventClick:
			analytics.ClickCount++
			uniqueClicks[t.EmailID] = true
			userClickTimes[t.EmailID] = append(userClickTimes[t.EmailID], t.Timestamp)

			// Track link analytics
			if _, exists := clickedLinks[t.URL]; !exists {
				clickedLinks[t.URL] = &LinkAnalytics{
					URL:             t.URL,
					DeviceBreakdown: make(map[string]int),
					FirstClickTime:  t.Timestamp.Format(time.RFC3339),
				}
			}
			if linkClickers[t.URL] == nil {
				linkClickers[t.URL] = map[string]bool{}
			}
			linkClickers[t.URL][t.EmailID] = true
			link := clickedLinks[t.URL]
			link.ClickCount++
			link.DeviceBreakdown[t.DeviceType]++
			link.LastClickTime = t.Timestamp.Format(time.RFC3339)

		case models.EmailTrackingEventBounce:
			analytics.BounceCount++

		case models.EmailTrackingEventComplaint:
			analytics.ComplaintCount++
		}
	}

	// Calculate engagement metrics
	analytics.UniqueOpens = len(uniqueOpens)
	analytics.UniqueClicks = len(uniqueClicks)
	analytics.RepeatOpens = analytics.OpenCount - analytics.UniqueOpens
	analytics.RepeatClicks = analytics.ClickCount - analytics.UniqueClicks

	if analytics.UniqueOpens > 0 {
		both := 0
		for id := range uniqueClicks {
			if uniqueOpens[id] {
				both++
			}
		}
		analytics.ClickToOpenRate = float64(both) / float64(analytics.UniqueOpens) * 100
	}

	// Calculate average read time (time between open and click)
	for userID, openTimes := range userOpenTimes {
		if clickTimes, hasClicks := userClickTimes[userID]; hasClicks {
			for _, openTime := range openTimes {
				for _, clickTime := range clickTimes {
					if clickTime.After(openTime) {
						readTime := clickTime.Sub(openTime)
						if readTime < 30*time.Minute { // Consider only reasonable read times
							totalReadTime += readTime
							readTimeCount++
						}
						break
					}
				}
			}
		}
	}

	if readTimeCount > 0 {
		analytics.AverageReadTime = totalReadTime / time.Duration(readTimeCount)
	}

	// Process clicked links
	for _, link := range clickedLinks {
		uniqueClicksCount := len(linkClickers[link.URL])
		link.UniqueClicks = uniqueClicksCount

		analytics.ClickedLinks = append(analytics.ClickedLinks, *link)
	}

	// Calculate engagement score (example formula)
	analytics.EngagementScore = calculateEngagementScore(analytics)

	return analytics
}

// 🎯 Calculate engagement score based on various metrics
// @Description Calculate engagement score based on various metrics
func calculateEngagementScore(analytics EmailAnalytics) float64 {
	// Example scoring formula (customize based on your needs)
	score := 0.0

	// Weight different factors
	if analytics.UniqueOpens > 0 {
		score += 20.0 * (float64(analytics.UniqueClicks) / float64(analytics.UniqueOpens))
	}

	// Add points for repeat engagement
	score += math.Min(float64(analytics.RepeatOpens)*0.5, 20.0)
	score += math.Min(float64(analytics.RepeatClicks), 20.0)

	// Subtract for negative signals
	bounceDeduction := math.Min(float64(analytics.BounceCount)*2.0, 20.0)
	complaintDeduction := math.Min(float64(analytics.ComplaintCount)*5.0, 20.0)
	score = math.Max(0, score-bounceDeduction-complaintDeduction)

	return math.Min(score, 100.0) // Cap at 100
}

// 📊 processCampaignAnalytics processes campaign analytics data
// @Description Process campaign analytics data
func processCampaignAnalytics(tracking []models.EmailTracking) EmailAnalytics {
	// Similar to processEmailAnalytics but with campaign-specific metrics
	return processEmailAnalytics(tracking, "UTC") // For now, reuse email analytics
}

// Advanced Analytics Response Structures

// 📊 TeamOverview represents team-wide analytics
type TeamOverview struct {
	TotalEmails      int                 `json:"totalEmails"`
	TotalOpens       int                 `json:"totalOpens"`
	TotalClicks      int                 `json:"totalClicks"`
	AverageOpenRate  float64             `json:"averageOpenRate"`
	AverageClickRate float64             `json:"averageClickRate"`
	TopCampaigns     []CampaignSummary   `json:"topCampaigns"`
	DeviceStats      map[string]int      `json:"deviceStats"`
	GeoStats         map[string]int      `json:"geoStats"`
	MonthlyStats     []MonthlyEngagement `json:"monthlyStats"`
	TopPerformers    []PerformerMetrics  `json:"topPerformers"`
}

// 📈 CampaignSummary represents summarized campaign data
type CampaignSummary struct {
	CampaignID      string  `json:"campaignId"`
	Name            string  `json:"name"`
	OpenRate        float64 `json:"openRate"`
	ClickRate       float64 `json:"clickRate"`
	EngagementScore float64 `json:"engagementScore"`
}

// 📅 MonthlyEngagement represents monthly engagement metrics
type MonthlyEngagement struct {
	Month       string  `json:"month"`
	OpenRate    float64 `json:"openRate"`
	ClickRate   float64 `json:"clickRate"`
	TotalEmails int     `json:"totalEmails"`
}

// 🎯 PerformerMetrics represents metrics for top-performing content
type PerformerMetrics struct {
	Type           string  `json:"type"` // subject, template, time, etc.
	Value          string  `json:"value"`
	EngagementRate float64 `json:"engagementRate"`
	SampleSize     int     `json:"sampleSize"`
}

// 🌡️ HeatmapData represents click heatmap data
type HeatmapData struct {
	URL      string         `json:"url"`
	Clicks   []ClickPoint   `json:"clicks"`
	Segments []ClickSegment `json:"segments"`
}

type ClickPoint struct {
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Count   int    `json:"count"`
	Element string `json:"element"`
}

type ClickSegment struct {
	Selector  string  `json:"selector"`
	ClickRate float64 `json:"clickRate"`
	Count     int     `json:"count"`
}

// ⏰ EngagementTimeData represents optimal engagement time data
type EngagementTimeData struct {
	HourlyBreakdown   map[int]EngagementMetrics    `json:"hourlyBreakdown"`
	DailyBreakdown    map[string]EngagementMetrics `json:"dailyBreakdown"`
	TimeZoneBreakdown map[string]EngagementMetrics `json:"timeZoneBreakdown"`
	OptimalSendTimes  []OptimalTimeSlot            `json:"optimalSendTimes"`
}

type EngagementMetrics struct {
	OpenCount  int     `json:"openCount"`
	ClickCount int     `json:"clickCount"`
	OpenRate   float64 `json:"openRate"`
	ClickRate  float64 `json:"clickRate"`
}

type OptimalTimeSlot struct {
	DayOfWeek       string  `json:"dayOfWeek"`
	Hour            int     `json:"hour"`
	EngagementScore float64 `json:"engagementScore"`
	Confidence      float64 `json:"confidence"`
}

// 👥 AudienceInsights represents audience analysis data
type AudienceInsights struct {
	Demographics map[string]int               `json:"demographics"`
	Behaviors    map[string]BehaviorMetrics   `json:"behaviors"`
	Segments     []AudienceSegment            `json:"segments"`
	Preferences  map[string]PreferenceMetrics `json:"preferences"`
}

type BehaviorMetrics struct {
	Count          int     `json:"count"`
	EngagementRate float64 `json:"engagementRate"`
	Trend          string  `json:"trend"`
}

type AudienceSegment struct {
	Name      string  `json:"name"`
	Size      int     `json:"size"`
	OpenRate  float64 `json:"openRate"`
	ClickRate float64 `json:"clickRate"`
	Growth    float64 `json:"growth"`
}

type PreferenceMetrics struct {
	Count      int     `json:"count"`
	Score      float64 `json:"score"`
	Confidence float64 `json:"confidence"`
}

// 📊 GetTeamOverview returns team-wide analytics
// @Summary Get team overview
// @Description Get team overview
// @Accept json
// @Produce json
// @Param teamId query string true "Team ID"
// @Success 200 {object} TeamOverview "Team overview"
// @Failure 400 {object} map[string]string "Validation error or team not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/analytics/team/overview [get]
func (h *TrackingHandler) GetTeamOverview(c echo.Context) error {
	f, err := h.legacyFilter(c)
	if err != nil {
		return err
	}
	r, err := analytics.Build(c.Request().Context(), h.db, f)
	if err != nil {
		return err
	}
	overview := TeamOverview{TotalEmails: int(r.Summary.Accepted), TotalOpens: int(r.Summary.Opened), TotalClicks: int(r.Summary.ClickedMessages), DeviceStats: map[string]int{}, GeoStats: map[string]int{}, TopCampaigns: []CampaignSummary{}}
	if r.Summary.ClickRate.Value != nil {
		overview.AverageClickRate = *r.Summary.ClickRate.Value * 100
	}
	if r.Summary.OpenRate.Value != nil {
		overview.AverageOpenRate = *r.Summary.OpenRate.Value * 100
	}
	c.Response().Header().Set("Deprecation", "true")
	return c.JSON(200, overview)
}

// 🔄 CompareCampaigns compares multiple campaigns
// @Summary Compare multiple campaigns
// @Description Compare multiple campaigns
// @Accept json
// @Produce json
// @Param campaignIds query string true "Campaign IDs"
// @Success 200 {object} map[string]EmailAnalytics "Campaign analytics"
// @Failure 400 {object} map[string]string "Validation error or campaignIds missing"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/analytics/campaign/compare [get]
func (h *TrackingHandler) CompareCampaigns(c echo.Context) error {
	ids := strings.Split(c.QueryParam("campaignIds"), ",")
	if len(ids) == 0 || ids[0] == "" {
		return echo.NewHTTPError(400, "Missing campaignIds")
	}
	results := map[string]EmailAnalytics{}
	for _, id := range ids {
		r, err := h.legacyDetail(c, "", id)
		if err != nil {
			return err
		}
		results[id] = r
	}
	return c.JSON(200, results)
}

// 🎯 GetClickHeatmap returns click heatmap data
// @Summary Get click heatmap
// @Description Get click heatmap
// @Accept json
// @Produce json
// @Param emailId query string true "Email ID"
// @Param campaignId query string true "Campaign ID"
// @Success 200 {object} HeatmapData "Click heatmap"
// @Failure 400 {object} map[string]string "Validation error or emailId missing"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/analytics/heatmap [get]
func (h *TrackingHandler) GetClickHeatmap(c echo.Context) error {
	emailID := c.QueryParam("emailId")
	campaignID := c.QueryParam("campaignId")

	if emailID == "" && campaignID == "" {
		return c.String(http.StatusBadRequest, "Missing both emailId and campaignId. At least one is required")
	}

	var tracking []models.EmailTracking
	query := h.db.Where("event = ?", models.EmailTrackingEventClick)

	if emailID != "" {
		query = query.Where("email_id = ?", emailID)
	} else {
		query = query.Where("campaign_id = ?", campaignID)
	}

	if err := query.Find(&tracking).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Failed to fetch click data")
	}

	// Process click data into heatmap format
	// Note: This would require additional client-side data about click coordinates
	heatmap := HeatmapData{
		Clicks:   make([]ClickPoint, 0),
		Segments: make([]ClickSegment, 0),
	}

	// Group clicks by URL and store the URL in heatmap
	urlClicks := make(map[string]int)
	if len(tracking) > 0 && tracking[0].URL != "" {
		heatmap.URL = tracking[0].URL // Store the URL being analyzed
	}

	for _, t := range tracking {
		urlClicks[t.URL]++
	}

	// Create segments for each clicked element
	for elementURL, count := range urlClicks {
		heatmap.Segments = append(heatmap.Segments, ClickSegment{
			Selector:  elementURL,
			Count:     count,
			ClickRate: float64(count) / float64(len(tracking)) * 100,
		})
	}

	return c.JSON(http.StatusOK, heatmap)
}

// ⏰ GetEngagementTimes returns optimal engagement time data
// @Summary Get engagement times
// @Description Get optimal engagement time data
// @Accept json
// @Produce json
// @Param teamId query string true "Team ID"
// @Success 200 {object} EngagementTimeData "Engagement time data"
// @Failure 400 {object} map[string]string "Validation error or team not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/analytics/engagement-times [get]

func (h *TrackingHandler) GetEngagementTimes(c echo.Context) error {
	teamID := c.QueryParam("teamId")
	if teamID == "" {
		return c.String(http.StatusBadRequest, "Missing teamId")
	}

	timeData := EngagementTimeData{
		HourlyBreakdown:   make(map[int]EngagementMetrics),
		DailyBreakdown:    make(map[string]EngagementMetrics),
		TimeZoneBreakdown: make(map[string]EngagementMetrics),
	}

	// Get all tracking data for the team
	var tracking []models.EmailTracking
	if err := h.db.Table("email_trackings").
		Joins("JOIN emails ON email_trackings.email_id = emails.id").
		Where("emails.team_id = ?", teamID).
		Find(&tracking).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Failed to fetch tracking data")
	}

	// Process tracking data by time
	for _, t := range tracking {
		hour := t.Timestamp.Hour()
		day := t.Timestamp.Weekday().String()

		// Update hourly metrics
		hourMetrics := timeData.HourlyBreakdown[hour]
		dayMetrics := timeData.DailyBreakdown[day]

		switch t.Event {
		case models.EmailTrackingEventOpen:
			hourMetrics.OpenCount++
			dayMetrics.OpenCount++
		case models.EmailTrackingEventClick:
			hourMetrics.ClickCount++
			dayMetrics.ClickCount++
		}

		timeData.HourlyBreakdown[hour] = hourMetrics
		timeData.DailyBreakdown[day] = dayMetrics
	}

	// Calculate optimal send times
	timeData.OptimalSendTimes = []OptimalTimeSlot{} // Observed activity is not evidence of an optimal send time.

	return c.JSON(http.StatusOK, timeData)
}

// 👥 GetAudienceInsights returns audience analysis
// @Summary Get audience insights
// @Description Get audience analysis
// @Accept json
// @Produce json
// @Param teamId query string true "Team ID"
// @Success 200 {object} AudienceInsights "Audience insights"
// @Failure 400 {object} map[string]string "Validation error or team not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/analytics/audience [get]
func (h *TrackingHandler) GetAudienceInsights(c echo.Context) error {
	return h.AudienceReport(c)
}

// 📈 GetTrendAnalysis returns trend analysis
// @Summary Get trend analysis
// @Description Get trend analysis
// @Accept json
// @Produce json
// @Param teamId query string true "Team ID"
// @Success 200 {object} []trendPoint "Trend analysis"
// @Failure 400 {object} map[string]string "Validation error or team not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/analytics/trends [get]
func (h *TrackingHandler) GetTrendAnalysis(c echo.Context) error {
	f, err := h.legacyFilter(c)
	if err != nil {
		return err
	}
	interval := c.QueryParam("interval")
	if interval == "" {
		interval = "daily"
	}
	if interval != "daily" && interval != "weekly" && interval != "monthly" {
		return echo.NewHTTPError(400, "Invalid interval")
	}
	data, err := analytics.Trends(c.Request().Context(), h.db, f, interval)
	if err != nil {
		return err
	}
	return c.JSON(200, data)
}

// 📊 ExportEmailAnalytics exports email analytics
// @Summary Export email analytics
// @Description Export email analytics
// @Accept json
// @Produce json
// @Param emailId query string true "Email ID"
// @Success 200 {object} []byte "Exported email analytics"
// @Failure 400 {object} map[string]string "Validation error or emailId missing"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/analytics/export/email [get]
func (h *TrackingHandler) ExportEmailAnalytics(c echo.Context) error {
	emailID := c.QueryParam("emailId")
	if emailID == "" {
		return c.String(http.StatusBadRequest, "Missing emailId")
	}

	format := c.QueryParam("format") // csv, xlsx
	var tracking []models.EmailTracking
	if err := h.db.Where("email_id = ?", emailID).Order("timestamp,id").Limit(100001).Find(&tracking).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Failed to fetch tracking data")
	}

	if len(tracking) > 100000 {
		return echo.NewHTTPError(400, "Export is too large; use filtered aggregate reports")
	}
	// Generate export data
	data := generateExportData(tracking, format)

	// Set appropriate headers for download
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=email_analytics_%s.%s", emailID, format))
	switch format {
	case "csv":
		return c.Blob(http.StatusOK, "text/csv", data)
	case "xlsx":
		return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
	default:
		return c.String(http.StatusBadRequest, "Unsupported format")
	}
}

// 📊 ExportCampaignAnalytics exports campaign analytics
// @Summary Export campaign analytics
// @Description Export campaign analytics
// @Accept json
// @Produce json
// @Param campaignId query string true "Campaign ID"
// @Success 200 {object} []byte "Exported campaign analytics"
// @Failure 400 {object} map[string]string "Validation error or campaignId missing"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/v1/analytics/export/campaign [get]
func (h *TrackingHandler) ExportCampaignAnalytics(c echo.Context) error {
	campaignID := c.QueryParam("campaignId")
	if campaignID == "" {
		return c.String(http.StatusBadRequest, "Missing campaignId")
	}

	format := c.QueryParam("format") // csv, xlsx
	var tracking []models.EmailTracking
	if err := h.db.Where("campaign_id = ?", campaignID).Order("timestamp,id").Limit(100001).Find(&tracking).Error; err != nil {
		return c.String(http.StatusInternalServerError, "Failed to fetch tracking data")
	}

	if len(tracking) > 100000 {
		return echo.NewHTTPError(400, "Export is too large; use filtered aggregate reports")
	}
	// Generate export data
	data := generateExportData(tracking, format)

	// Set appropriate headers for download
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=campaign_analytics_%s.%s", campaignID, format))
	switch format {
	case "csv":
		return c.Blob(http.StatusOK, "text/csv", data)
	case "xlsx":
		return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
	default:
		return c.String(http.StatusBadRequest, "Unsupported format")
	}
}

// Helper functions
func calculateOptimalSendTimes(data EngagementTimeData) []OptimalTimeSlot {
	slots := make([]OptimalTimeSlot, 0)

	// Analyze each hour of each day
	for day := 0; day < 7; day++ {
		dayName := time.Weekday(day).String()
		dayMetrics := data.DailyBreakdown[dayName]

		for hour := 0; hour < 24; hour++ {
			hourMetrics := data.HourlyBreakdown[hour]

			// Calculate engagement score for this time slot
			engagementScore := calculateTimeSlotScore(hourMetrics, dayMetrics)

			// Calculate confidence based on sample size
			confidence := calculateConfidence(hourMetrics.OpenCount + hourMetrics.ClickCount)

			if engagementScore > 0 {
				slots = append(slots, OptimalTimeSlot{
					DayOfWeek:       dayName,
					Hour:            hour,
					EngagementScore: engagementScore,
					Confidence:      confidence,
				})
			}
		}
	}

	// Sort by engagement score
	sort.Slice(slots, func(i, j int) bool {
		return slots[i].EngagementScore > slots[j].EngagementScore
	})

	return slots
}

// 📈 calculateTimeSlotScore computes engagement score for a time slot
func calculateTimeSlotScore(hourly, daily EngagementMetrics) float64 {
	if hourly.OpenCount == 0 && hourly.ClickCount == 0 {
		return 0
	}

	// Weight different metrics
	openWeight := 0.6
	clickWeight := 0.4

	// Calculate normalized rates
	hourlyScore := (float64(hourly.OpenCount)*openWeight + float64(hourly.ClickCount)*clickWeight)
	dailyScore := (float64(daily.OpenCount)*openWeight + float64(daily.ClickCount)*clickWeight)

	// Combine scores with time-based weighting
	return (hourlyScore*0.7 + dailyScore*0.3) * 100
}

// 📊 calculateConfidence determines confidence level based on sample size
func calculateConfidence(sampleSize int) float64 {
	// Basic confidence calculation
	// Could be enhanced with statistical methods
	baseLine := 30.0 // Minimum sample size for reasonable confidence
	maxConfidence := 95.0

	if sampleSize < int(baseLine) {
		return (float64(sampleSize) / baseLine) * maxConfidence
	}
	return maxConfidence
}

// 📈 calculateEngagementTrend analyzes engagement pattern over time
func calculateEngagementTrend(tracking []models.EmailTracking) string {
	if len(tracking) < 2 {
		return "insufficient_data"
	}

	// Sort tracking by timestamp
	sort.Slice(tracking, func(i, j int) bool {
		return tracking[i].Timestamp.Before(tracking[j].Timestamp)
	})

	// Calculate engagement rates over time
	type periodMetrics struct {
		opens  int
		clicks int
		total  int
	}

	periods := make(map[string]periodMetrics)
	for _, t := range tracking {
		period := t.Timestamp.Format("2006-01")
		metrics := periods[period]
		metrics.total++

		switch t.Event {
		case models.EmailTrackingEventOpen:
			metrics.opens++
		case models.EmailTrackingEventClick:
			metrics.clicks++
		}
		periods[period] = metrics
	}

	// Analyze trend
	var rates []float64
	keys := make([]string, 0, len(periods))
	for key := range periods {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		metrics := periods[key]
		if metrics.total > 0 {
			rate := float64(metrics.opens+metrics.clicks) / float64(metrics.total)
			rates = append(rates, rate)
		}
	}

	if len(rates) < 2 {
		return "stable"
	}

	// Calculate trend
	firstHalf := rates[:len(rates)/2]
	secondHalf := rates[len(rates)/2:]

	firstAvg := average(firstHalf)
	secondAvg := average(secondHalf)

	if firstAvg == 0 {
		if secondAvg > 0 {
			return "increasing"
		}
		return "stable"
	}
	changePct := ((secondAvg - firstAvg) / firstAvg) * 100

	switch {
	case changePct > 10:
		return "increasing"
	case changePct < -10:
		return "decreasing"
	default:
		return "stable"
	}
}

type trendPoint struct {
	Period         string                 `json:"period"`
	OpenCount      int                    `json:"openCount"`
	ClickCount     int                    `json:"clickCount"`
	EngagementRate float64                `json:"engagementRate"`
	Devices        map[string]int         `json:"devices"`
	Locations      map[string]int         `json:"locations"`
	Growth         float64                `json:"growth"`
	Events         []models.EmailTracking `json:"events"`
}

// 📊 processTrends generates trend analysis data
func processTrends(tracking []models.EmailTracking, interval string) []trendPoint {

	trends := make(map[string]*trendPoint)

	// Group by interval
	for _, t := range tracking {
		var period string
		switch interval {
		case "daily":
			period = t.Timestamp.Format("2006-01-02")
		case "weekly":
			year, week := t.Timestamp.ISOWeek()
			period = fmt.Sprintf("%d-W%02d", year, week)
		case "monthly":
			period = t.Timestamp.Format("2006-01")
		default:
			period = t.Timestamp.Format("2006-01") // Default to monthly
		}

		if trends[period] == nil {
			trends[period] = &trendPoint{
				Period:    period,
				Devices:   make(map[string]int),
				Locations: make(map[string]int),
			}
		}

		point := trends[period]
		point.Events = append(point.Events, t)

		switch t.Event {
		case models.EmailTrackingEventOpen:
			point.OpenCount++
		case models.EmailTrackingEventClick:
			point.ClickCount++
		}

		point.Devices[t.DeviceType]++
		if t.Country != "" {
			point.Locations[t.Country]++
		}
	}

	// Calculate rates and growth
	var sortedPeriods []string
	for period := range trends {
		sortedPeriods = append(sortedPeriods, period)
	}
	sort.Strings(sortedPeriods)

	var result []trendPoint
	for i, period := range sortedPeriods {
		point := trends[period]
		total := len(point.Events)
		if total > 0 {
			point.EngagementRate = float64(point.OpenCount+point.ClickCount) / float64(total) * 100
		}

		if i > 0 {
			prevPoint := trends[sortedPeriods[i-1]]
			if prevPoint.EngagementRate > 0 {
				point.Growth = ((point.EngagementRate - prevPoint.EngagementRate) / prevPoint.EngagementRate) * 100
			}
		}

		result = append(result, *point)
	}

	return result
}

// 📋 generateExportData creates formatted export data
func generateExportData(tracking []models.EmailTracking, format string) []byte {
	switch format {
	case "csv":
		return generateCSV(tracking)
	case "xlsx":
		return generateXLSX(tracking)
	default:
		return nil
	}
}

// 📄 generateCSV creates CSV formatted data
func generateCSV(tracking []models.EmailTracking) []byte {
	buffer := &bytes.Buffer{}
	writer := csv.NewWriter(buffer)

	// Write header
	writer.Write([]string{
		"Timestamp",
		"Event",
		"Device",
		"Browser",
		"OS",
		"Country",
		"City",
		"Region",
		"URL",
	})

	// Write data
	for _, t := range tracking {
		writer.Write([]string{
			t.Timestamp.Format(time.RFC3339),
			string(t.Event),
			t.DeviceType,
			t.Browser,
			t.OS,
			t.Country,
			t.City,
			t.Region,
			t.URL,
		})
	}

	writer.Flush()
	return buffer.Bytes()
}

// 📊 generateXLSX creates Excel formatted data
func generateXLSX(tracking []models.EmailTracking) []byte {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			trackingLog.Error("Failed to close Excel file", err)
		}
	}()

	sheet := "Sheet1"
	// Write header
	headers := []string{
		"Timestamp",
		"Event",
		"Device",
		"Browser",
		"OS",
		"Country",
		"City",
		"Region",
		"URL",
	}

	for i, header := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheet, col+"1", header)
	}

	// Write data
	for i, t := range tracking {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), t.Timestamp.Format(time.RFC3339))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), string(t.Event))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), t.DeviceType)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), t.Browser)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), t.OS)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), t.Country)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), t.City)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), t.Region)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), t.URL)
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
		trackingLog.Error("Failed to write Excel to buffer", err)
		return nil
	}
	return buffer.Bytes()
}

// 📊 average calculates the average of a slice of float64
func average(numbers []float64) float64 {
	if len(numbers) == 0 {
		return 0
	}

	sum := 0.0
	for _, n := range numbers {
		sum += n
	}
	return sum / float64(len(numbers))
}
