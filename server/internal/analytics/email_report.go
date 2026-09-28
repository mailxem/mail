package analytics

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type EmailTotals struct {
	Total     int64 `json:"total"`
	Sent      int64 `json:"sent"`
	Queued    int64 `json:"queued"`
	Failed    int64 `json:"failed"`
	Drafts    int64 `json:"drafts"`
	Unknown   int64 `json:"unknown"`
	Opened    int64 `json:"opened"`
	Clicked   int64 `json:"clicked"`
	Bounced   int64 `json:"bounced"`
	Campaigns int64 `json:"campaigns"`
	Other     int64 `json:"other"`
}
type EmailDay struct {
	Date string `json:"date"`
	EmailTotals
}
type EmailReport struct {
	Version    string      `json:"metricVersion"`
	From       time.Time   `json:"from"`
	To         time.Time   `json:"to"`
	AsOf       time.Time   `json:"asOf"`
	Timezone   string      `json:"timezone"`
	Summary    EmailTotals `json:"summary"`
	OpenRate   Rate        `json:"openRate"`
	ClickRate  Rate        `json:"clickRate"`
	BounceRate Rate        `json:"bounceRate"`
	Series     []EmailDay  `json:"series"`
}

// BuildEmailReport includes every workspace email, without requiring a campaign,
// list or contact. Sent messages are attributed to sent_at; unsent/unknown records
// to created_at. Queue/failure/draft statuses reflect the current database state.
func BuildEmailReport(ctx context.Context, db *gorm.DB, f Filter) (EmailReport, error) {
	r := EmailReport{Version: "email-v1", From: f.From, To: f.To, AsOf: f.AsOf, Timezone: f.Timezone, Series: []EmailDay{}}
	if f.Team == "" {
		return r, fmt.Errorf("Workspace is required")
	}
	if f.List != "" || f.Tag != "" || f.Campaign != "" {
		return r, fmt.Errorf("Email overview supports date and timezone filters; use campaign analytics for audience filters")
	}
	const query = `WITH scope AS (
 SELECT e.*, (e.sent_at>'2000-01-01' AND e.status IN ('SENT','OPENED','CLICKED','BOUNCED')) accepted,
 CASE WHEN e.sent_at>'2000-01-01' AND e.status IN ('SENT','OPENED','CLICKED','BOUNCED') THEN e.sent_at ELSE e.created_at END recorded_at
 FROM emails e WHERE e.team_id=@team AND e.is_deleted=false AND e.test=false
 ), selected AS (
 SELECT * FROM scope WHERE recorded_at>=CAST(@from AS timestamptz) AND recorded_at<CAST(@asof AS timestamptz)
 ), facts AS (
 SELECT e.*,
 e.accepted AND EXISTS(SELECT 1 FROM email_trackings t WHERE t.email_id=e.id AND t.is_deleted=false AND t.event='open' AND t.timestamp>=e.sent_at AND t.timestamp<CAST(@asof AS timestamptz)) opened,
 e.accepted AND EXISTS(SELECT 1 FROM email_trackings t WHERE t.email_id=e.id AND t.is_deleted=false AND t.event='click' AND t.timestamp>=e.sent_at AND t.timestamp<CAST(@asof AS timestamptz)) clicked,
 e.accepted AND (e.status='BOUNCED' OR EXISTS(SELECT 1 FROM email_trackings t WHERE t.email_id=e.id AND t.is_deleted=false AND t.event='bounce' AND t.timestamp>=e.sent_at AND t.timestamp<CAST(@asof AS timestamptz)) OR EXISTS(SELECT 1 FROM email_bounces b WHERE b.email_id=e.id AND b.team_id=@team AND b.is_deleted=false AND b.created_at>=e.sent_at AND b.created_at<CAST(@asof AS timestamptz))) bounced
 FROM selected e)
 SELECT to_char(recorded_at AT TIME ZONE @tz,'YYYY-MM-DD') date,
 COUNT(*) total,
 COUNT(*) FILTER(WHERE accepted) sent,
 COUNT(*) FILTER(WHERE NOT accepted AND status IN ('QUEUED','PENDING','SENDING','SCHEDULED')) queued,
 COUNT(*) FILTER(WHERE NOT accepted AND status='FAILED') failed,
 COUNT(*) FILTER(WHERE NOT accepted AND status='DRAFT') drafts,
 COUNT(*) FILTER(WHERE NOT accepted AND status NOT IN ('QUEUED','PENDING','SENDING','SCHEDULED','FAILED','DRAFT')) unknown,
 COUNT(*) FILTER(WHERE opened) opened, COUNT(*) FILTER(WHERE clicked) clicked, COUNT(*) FILTER(WHERE bounced) bounced,
 COUNT(*) FILTER(WHERE accepted AND campaign_id IS NOT NULL) campaigns,
 COUNT(*) FILTER(WHERE accepted AND campaign_id IS NULL) other
 FROM facts GROUP BY 1 ORDER BY 1`
	rows := []EmailDay{}
	if err := db.WithContext(ctx).Raw(query, f.args()).Scan(&rows).Error; err != nil {
		return r, err
	}
	byDate := map[string]EmailDay{}
	for _, d := range rows {
		byDate[d.Date] = d
		r.Summary.Total += d.Total
		r.Summary.Sent += d.Sent
		r.Summary.Queued += d.Queued
		r.Summary.Failed += d.Failed
		r.Summary.Drafts += d.Drafts
		r.Summary.Unknown += d.Unknown
		r.Summary.Opened += d.Opened
		r.Summary.Clicked += d.Clicked
		r.Summary.Bounced += d.Bounced
		r.Summary.Campaigns += d.Campaigns
		r.Summary.Other += d.Other
	}
	loc, err := time.LoadLocation(f.Timezone)
	if err != nil {
		return r, err
	}
	for day := f.From.In(loc); day.Before(f.AsOf); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		d := byDate[key]
		d.Date = key
		r.Series = append(r.Series, d)
	}
	r.OpenRate = Ratio(r.Summary.Opened, r.Summary.Sent)
	r.ClickRate = Ratio(r.Summary.Clicked, r.Summary.Sent)
	r.BounceRate = Ratio(r.Summary.Bounced, r.Summary.Sent)
	return r, nil
}
