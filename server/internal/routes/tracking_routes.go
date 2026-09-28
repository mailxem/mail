package routes

import (
	"kori/internal/api/middleware"
	"kori/internal/config"
	"kori/internal/handlers"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// 📊 RegisterTrackingRoutes registers all tracking related routes
func RegisterTrackingRoutes(e *echo.Echo, h *handlers.TrackingHandler, cfg *config.Config, db *gorm.DB) {
	// Public tracking endpoints (no auth required)
	trackGroup := e.Group("/t")
	trackGroup.GET("/click/*", h.HandleEmailClick) // The * captures the rest of the URL
	trackGroup.GET("/open", h.HandleEmailOpen)
	trackGroup.GET("/unsubscribe", h.HandleEmailUnsubscribe)
	trackGroup.POST("/unsubscribe", h.HandleEmailUnsubscribe)
	trackGroup.GET("/resubscribe", h.HandleEmailResubscribe)
	trackGroup.POST("/resubscribe", h.HandleEmailResubscribe) // Resubscribe to an email list

	// Analytics endpoints (require auth)
	analyticsGroup := e.Group("/api/v1/analytics")
	// Add authentication middleware
	auth := middleware.NewAuthMiddleware(cfg.JWT.Secret)
	analyticsGroup.Use(auth.Middleware())

	analyticsGroup.Use(middleware.AnalyticsAccess(db))

	// Basic analytics
	// @Summary Get email analytics
	analyticsGroup.GET("/email", h.GetEmailAnalytics)

	// @Summary Get campaign analytics
	// @Description Get campaign analytics
	analyticsGroup.GET("/campaign", h.GetCampaignAnalytics)

	// Advanced analytics
	// @Summary Get team overview
	analyticsGroup.GET("/team/overview", h.GetTeamOverview) // Team-wide stats

	// @Summary Compare multiple campaigns
	// @Description Compare multiple campaigns
	analyticsGroup.GET("/campaign/compare", h.CompareCampaigns) // Compare multiple campaigns

	// @Summary Get click heatmap
	// @Description Get click heatmap
	analyticsGroup.GET("/heatmap", h.GetClickHeatmap) // Click heatmap data

	// @Summary Get best engagement times
	// @Description Get best engagement times
	analyticsGroup.GET("/engagement-times", h.GetEngagementTimes) // Best engagement times

	// @Summary Get audience insights
	// @Description Get audience insights
	analyticsGroup.GET("/report", h.AudienceReport)
	analyticsGroup.GET("/email-overview", h.EmailOverview)
	analyticsGroup.GET("/breakdown", h.AnalyticsBreakdown)
	analyticsGroup.GET("/people", h.AnalyticsPeople)
	analyticsGroup.GET("/options", h.AnalyticsOptions)
	analyticsGroup.GET("/audience", h.AudienceReport) // Audience insights

	// @Summary Get trend analysis
	// @Description Get trend analysis
	analyticsGroup.GET("/trends", h.GetTrendAnalysis) // Trend analysis

	// Export endpoints
	// @Summary Export email analytics
	analyticsGroup.GET("/export/email", h.ExportEmailAnalytics) // Export email analytics

	// @Summary Export campaign analytics
	// @Description Export campaign analytics
	analyticsGroup.GET("/export/campaign", h.ExportCampaignAnalytics) // Export campaign analytics
}
