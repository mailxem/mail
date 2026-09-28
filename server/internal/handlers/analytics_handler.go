package handlers

import (
	"context"
	"github.com/labstack/echo/v4"
	"kori/internal/analytics"
	"kori/internal/api/middleware"
	"kori/internal/models"
	"net/http"
	"time"
)

func (h *TrackingHandler) analyticsFilter(c echo.Context) (analytics.Filter, error) {
	f, err := analytics.Parse(c.QueryParams(), middleware.GetTeamID(c), time.Now())
	if err != nil {
		return f, echo.NewHTTPError(400, err.Error())
	}
	if f.Team == "" {
		return f, echo.NewHTTPError(401, "Sign in to view analytics")
	}
	return f, nil
}
func (h *TrackingHandler) AudienceReport(c echo.Context) error {
	f, err := h.analyticsFilter(c)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 20*time.Second)
	defer cancel()
	report, err := analytics.Build(ctx, h.db, f)
	if err != nil {
		trackingLog.Error("Analytics report failed", err)
		return echo.NewHTTPError(500, "Unable to load analytics. Please try again.")
	}
	return c.JSON(http.StatusOK, report)
}
func (h *TrackingHandler) AnalyticsBreakdown(c echo.Context) error {
	f, err := h.analyticsFilter(c)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 20*time.Second)
	defer cancel()
	switch f.Kind {
	case "lists", "campaigns", "domains", "sources", "links":
	default:
		return echo.NewHTTPError(400, "Invalid breakdown")
	}
	switch f.Sort {
	case "", "name", "accepted", "clicked", "active", "bounced", "clickRate", "new":
	default:
		return echo.NewHTTPError(400, "Invalid sort")
	}
	report, err := analytics.Breakdown(ctx, h.db, f)
	if err != nil {
		trackingLog.Error("Analytics breakdown failed", err)
		return echo.NewHTTPError(500, "Unable to load breakdown. Please try again.")
	}
	return c.JSON(200, report)
}
func (h *TrackingHandler) AnalyticsPeople(c echo.Context) error {
	f, err := h.analyticsFilter(c)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 20*time.Second)
	defer cancel()
	report, err := analytics.People(ctx, h.db, f)
	if err != nil {
		return echo.NewHTTPError(500, "Unable to load audience")
	}
	return c.JSON(200, report)
}
func (h *TrackingHandler) AnalyticsOptions(c echo.Context) error {
	team := middleware.GetTeamID(c)
	if team == "" {
		return echo.NewHTTPError(401, "Sign in to view analytics")
	}
	type Option struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	lists, tags := []Option{}, []Option{}
	if err := h.db.WithContext(c.Request().Context()).Model(&models.MailingList{}).Where("team_id=? AND is_deleted=false", team).Select("id,name").Order("name,id").Limit(200).Scan(&lists).Error; err != nil {
		return err
	}
	if err := h.db.WithContext(c.Request().Context()).Model(&models.Tag{}).Where("team_id=? AND is_deleted=false", team).Select("id,name").Order("name,id").Limit(200).Scan(&tags).Error; err != nil {
		return err
	}
	return c.JSON(200, map[string]interface{}{"lists": lists, "tags": tags})
}

func (h *TrackingHandler) EmailOverview(c echo.Context) error {
	f, err := h.analyticsFilter(c)
	if err != nil {
		return err
	}
	if f.List != "" || f.Tag != "" || f.Campaign != "" {
		return echo.NewHTTPError(400, "Use campaign analytics for audience filters")
	}
	ctx, cancel := context.WithTimeout(c.Request().Context(), 20*time.Second)
	defer cancel()
	report, err := analytics.BuildEmailReport(ctx, h.db, f)
	if err != nil {
		trackingLog.Error("Email overview failed", err)
		return echo.NewHTTPError(500, "Unable to load email activity. Please try again.")
	}
	return c.JSON(http.StatusOK, report)
}
