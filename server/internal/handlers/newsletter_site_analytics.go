package handlers

import (
	"encoding/json"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"kori/internal/models"
)

var siteSession = regexp.MustCompile(`^[A-Za-z0-9]{8,24}$`)

func siteDevice(width int) string {
	switch {
	case width < 768:
		return "mobile"
	case width < 1100:
		return "tablet"
	}
	return "desktop"
}

// RecordSiteEvent stores a page view or click sent by a newsletter website.
// Browsers send it with navigator.sendBeacon, so the body is JSON in a
// text/plain request.
func (h *MarketingHandler) RecordSiteEvent(c echo.Context) error {
	n, err := h.publicNewsletter(c)
	if err != nil {
		return err
	}
	raw, _ := io.ReadAll(io.LimitReader(c.Request().Body, 2048))
	var in struct {
		T string  `json:"t"`
		P string  `json:"p"`
		R string  `json:"r"`
		S string  `json:"s"`
		L string  `json:"l"`
		W int     `json:"w"`
		X float32 `json:"x"`
		Y int32   `json:"y"`
		H int32   `json:"h"`
	}
	if json.Unmarshal(raw, &in) != nil || (in.T != "view" && in.T != "click") || !strings.HasPrefix(in.P, "/") || len(in.P) > 200 ||
		!siteSession.MatchString(in.S) || in.W < 200 || in.W > 8000 || in.X < 0 || in.X > 1 || in.Y < 0 || in.Y > 200000 || in.H < 0 || in.H > 200000 {
		return bad("Invalid event")
	}
	path, _, _ := strings.Cut(in.P, "?")
	e := models.SiteEvent{NewsletterID: n.ID, Kind: in.T, Path: path, Session: in.S, Device: siteDevice(in.W), Height: in.H}
	if in.T == "click" {
		e.X, e.Y = in.X, in.Y
		if u, err := url.Parse(in.L); err == nil && (u.Scheme == "https" || u.Scheme == "http" || u.Scheme == "mailto") && len(in.L) <= 200 {
			e.Link = in.L
		}
	}
	if u, err := url.Parse(in.R); err == nil && u.Host != "" && len(u.Host) <= 100 {
		e.Referrer = strings.TrimPrefix(strings.ToLower(u.Host), "www.")
	}
	if err := h.service.DB.WithContext(c.Request().Context()).Create(&e).Error; err != nil {
		return err
	}
	return c.NoContent(204)
}

// siteDay is the SQL for a site event's day, "YYYY-MM-DD".
func siteDay(db *gorm.DB) string {
	if db.Dialector.Name() == "sqlite" {
		return "strftime('%Y-%m-%d', created_at)"
	}
	return "to_char(created_at, 'YYYY-MM-DD')"
}

type siteCount struct {
	Key      string `json:"key"`
	Views    int64  `json:"views"`
	Visitors int64  `json:"visitors"`
}

// SiteAnalytics reports a newsletter website's traffic over the last ?days
// (7, 30 or 90), and a click heatmap for ?path on ?device (desktop, tablet or
// mobile; desktop by default).
func (h *MarketingHandler) SiteAnalytics(c echo.Context) error {
	n, err := h.scopedNewsletter(c)
	if err != nil {
		return err
	}
	days, _ := strconv.Atoi(c.QueryParam("days"))
	if days != 7 && days != 90 {
		days = 30
	}
	since := time.Now().UTC().AddDate(0, 0, -days)
	db := h.service.DB.WithContext(c.Request().Context())
	events := func() *gorm.DB {
		return db.Model(&models.SiteEvent{}).Where("newsletter_id = ? AND created_at >= ?", n.ID, since)
	}

	var totals struct{ Views, Visitors, Clicks int64 }
	if err := events().Select("COALESCE(SUM(CASE WHEN kind = 'view' THEN 1 ELSE 0 END), 0) AS views, COUNT(DISTINCT CASE WHEN kind = 'view' THEN session END) AS visitors, COALESCE(SUM(CASE WHEN kind = 'click' THEN 1 ELSE 0 END), 0) AS clicks").Scan(&totals).Error; err != nil {
		return err
	}
	var signups int64
	if err := db.Model(&models.NewsletterSignup{}).Where("newsletter_id = ? AND created_at >= ?", n.ID, since).Count(&signups).Error; err != nil {
		return err
	}
	group := func(col string, limit int, extra string) ([]siteCount, error) {
		out := []siteCount{}
		q := events().Where("kind = 'view'")
		if extra != "" {
			q = q.Where(extra)
		}
		err := q.Select(col + " AS key, COUNT(*) AS views, COUNT(DISTINCT session) AS visitors").Group(col).Order("views DESC").Limit(limit).Scan(&out).Error
		return out, err
	}
	daily, err := group(siteDay(db), days+1, "")
	if err != nil {
		return err
	}
	sort.Slice(daily, func(i, j int) bool { return daily[i].Key < daily[j].Key })
	pages, err := group("path", 10, "")
	if err != nil {
		return err
	}
	referrers, err := group("referrer", 8, "referrer <> ''")
	if err != nil {
		return err
	}
	devices, err := group("device", 3, "")
	if err != nil {
		return err
	}
	links := []struct {
		Link   string `json:"link"`
		Clicks int64  `json:"clicks"`
	}{}
	if err := events().Where("kind = 'click' AND link <> ''").Select("link, COUNT(*) AS clicks").Group("link").Order("clicks DESC").Limit(8).Scan(&links).Error; err != nil {
		return err
	}

	path := c.QueryParam("path")
	if path == "" && len(pages) > 0 {
		path = pages[0].Key
	}
	device := c.QueryParam("device")
	if device != "mobile" && device != "tablet" {
		device = "desktop"
	}
	// ponytail: aggregates up to 20k recent clicks in Go; move to a SQL
	// grid query if a site outgrows that.
	var clicks []struct {
		X      float32
		Y      int32
		Height int32
	}
	if err := events().Where("kind = 'click' AND path = ? AND device = ?", path, device).Select("x, y, height").Order("created_at DESC").Limit(20000).Scan(&clicks).Error; err != nil {
		return err
	}
	const cols, rowPx = 40, 20
	cells := map[[2]int]int{}
	var height int32
	for _, k := range clicks {
		cells[[2]int{min(int(k.X*cols), cols-1), int(k.Y) / rowPx}]++
		height = max(height, k.Height)
	}
	type cell struct{ X, Y, N int }
	points := make([]cell, 0, len(cells))
	for k, v := range cells {
		points = append(points, cell{k[0], k[1], v})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].N > points[j].N })

	return c.JSON(200, map[string]any{
		"days": days, "views": totals.Views, "visitors": totals.Visitors, "clicks": totals.Clicks, "signups": signups,
		"daily": daily, "pages": pages, "referrers": referrers, "devices": devices, "links": links,
		"heatmap": map[string]any{"path": path, "device": device, "clicks": len(clicks), "height": height, "cols": cols, "rowPx": rowPx, "cells": points},
	})
}
