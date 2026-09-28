package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"gorm.io/gorm"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Filter struct {
	Team, List, Tag, Campaign, Timezone   string
	From, To, AsOf, PreviousFrom          time.Time
	Page, Limit                           int
	Sort, Direction, Kind, Cohort, Search string
}

func Parse(q url.Values, team string, now time.Time) (Filter, error) {
	f := Filter{Team: team, List: q.Get("listId"), Tag: q.Get("tagId"), Campaign: q.Get("campaignId"), Timezone: q.Get("timezone"), Page: 1, Limit: 10, Sort: q.Get("sort"), Direction: q.Get("direction"), Kind: q.Get("kind"), Cohort: q.Get("cohort"), Search: strings.TrimSpace(q.Get("search")), AsOf: now.UTC()}
	if f.Timezone == "" {
		f.Timezone = "UTC"
	}
	loc, err := time.LoadLocation(f.Timezone)
	if err != nil {
		return f, fmt.Errorf("Invalid timezone")
	}
	today := time.Date(now.In(loc).Year(), now.In(loc).Month(), now.In(loc).Day(), 0, 0, 0, 0, loc)
	f.From = today.AddDate(0, 0, -29)
	f.To = today.AddDate(0, 0, 1)
	for _, v := range []struct {
		k      string
		target *time.Time
	}{{"from", &f.From}, {"to", &f.To}, {"asOf", &f.AsOf}} {
		if raw := q.Get(v.k); raw != "" {
			t, e := time.Parse(time.RFC3339, raw)
			if e != nil {
				t, e = time.ParseInLocation("2006-01-02", raw, loc)
			}
			if e != nil {
				return f, fmt.Errorf("Invalid %s", v.k)
			}
			*v.target = t
		}
	}
	if !f.From.Before(f.To) || f.To.Sub(f.From) > 367*24*time.Hour || f.From.After(now) || f.AsOf.After(now.Add(time.Minute)) || !f.AsOf.After(f.From) {
		return f, fmt.Errorf("Choose a past or current range of at most 366 days")
	}
	if f.AsOf.After(f.To) {
		f.AsOf = f.To
	}
	localTo, localFrom := f.To.In(loc), f.From.In(loc)
	days := int(time.Date(localTo.Year(), localTo.Month(), localTo.Day(), 0, 0, 0, 0, time.UTC).Sub(time.Date(localFrom.Year(), localFrom.Month(), localFrom.Day(), 0, 0, 0, 0, time.UTC)).Hours() / 24)
	if days < 1 {
		f.PreviousFrom = f.From.Add(-f.To.Sub(f.From))
	} else {
		f.PreviousFrom = f.From.AddDate(0, 0, -days)
	}
	for _, v := range []struct {
		k      string
		target *int
		max    int
	}{{"page", &f.Page, 1000000}, {"limit", &f.Limit, 100}} {
		if raw := q.Get(v.k); raw != "" {
			n, e := strconv.Atoi(raw)
			if e != nil || n < 1 || n > v.max {
				return f, fmt.Errorf("Invalid %s", v.k)
			}
			*v.target = n
		}
	}
	if f.Direction == "" {
		f.Direction = "desc"
	}
	if f.Direction != "asc" && f.Direction != "desc" {
		return f, fmt.Errorf("Invalid sort direction")
	}
	if len(f.Search) > 200 {
		return f, fmt.Errorf("Search is too long")
	}
	if f.Cohort != "" && !strings.Contains("|recent|earlier|no_clicks|not_contacted|insufficient|", "|"+f.Cohort+"|") {
		return f, fmt.Errorf("Invalid cohort")
	}
	return f, nil
}

type Rate struct {
	Value       *float64 `json:"value"`
	Numerator   int64    `json:"numerator"`
	Denominator int64    `json:"denominator"`
}

func Ratio(n, d int64) Rate {
	r := Rate{Numerator: n, Denominator: d}
	if d > 0 {
		v := float64(n) / float64(d)
		r.Value = &v
	}
	return r
}

type Summary struct {
	Active            int64 `json:"activeSubscribers"`
	New               int64 `json:"newContactRecords"`
	Reached           int64 `json:"recipientsReached"`
	Clicked           int64 `json:"recipientsClicked"`
	Accepted          int64 `json:"accepted"`
	Opened            int64 `json:"opened"`
	ClickedMessages   int64 `json:"clickedMessages"`
	Bounced           int64 `json:"bounced"`
	Complaints        int64 `json:"complaints"`
	Unsubscribes      int64 `json:"unsubscribes"`
	Failed            int64 `json:"failed"`
	Unknown           int64 `json:"deliveryUnknown"`
	TrackingKnown     int64 `json:"trackingKnown"`
	TrackingEnabled   int64 `json:"trackingEnabled"`
	AudienceClickRate Rate  `json:"audienceClickRate" gorm:"-"`
	ClickRate         Rate  `json:"clickRate" gorm:"-"`
	OpenRate          Rate  `json:"openRate" gorm:"-"`
	BounceRate        Rate  `json:"bounceRate" gorm:"-"`
	ComplaintRate     Rate  `json:"complaintRate" gorm:"-"`
}

func (s *Summary) rates() {
	s.AudienceClickRate = Ratio(s.Clicked, s.Reached)
	s.ClickRate = Ratio(s.ClickedMessages, s.Accepted)
	s.OpenRate = Ratio(s.Opened, s.Accepted)
	s.BounceRate = Ratio(s.Bounced, s.Accepted)
	s.ComplaintRate = Ratio(s.Complaints, s.Accepted)
}

type Day struct {
	Date    string `json:"date"`
	Reached int64  `json:"reached"`
	Clicked int64  `json:"clicked"`
}
type Cohort struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}
type GrowthDay struct {
	Date   string `json:"date"`
	Active int64  `json:"active"`
}
type Growth struct {
	Since     *time.Time  `json:"since"`
	Available bool        `json:"available"`
	Start     *int64      `json:"start"`
	End       *int64      `json:"end"`
	Net       *int64      `json:"net"`
	Series    []GrowthDay `json:"series"`
}
type Report struct {
	Version  string    `json:"metricVersion"`
	AsOf     time.Time `json:"asOf"`
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
	Timezone string    `json:"timezone"`
	Summary  Summary   `json:"summary"`
	Previous Summary   `json:"previous"`
	Activity []Day     `json:"activity"`
	Cohorts  []Cohort  `json:"cohorts"`
	Growth   Growth    `json:"growth"`
	Warnings []string  `json:"warnings"`
	Charts   Charts    `json:"charts"`
}

func (f Filter) args() map[string]interface{} {
	return map[string]interface{}{"team": f.Team, "list": f.List, "tag": f.Tag, "campaign": f.Campaign, "from": f.From, "to": f.To, "asof": f.AsOf, "previous": f.PreviousFrom, "tz": f.Timezone, "recent": f.AsOf.AddDate(0, 0, -30), "lookback": f.AsOf.AddDate(0, 0, -90), "limit": f.Limit, "offset": (f.Page - 1) * f.Limit, "cohort": f.Cohort, "search": "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(strings.ToLower(f.Search)) + "%"}
}

// Names in this SQL are static. Values (including timezone and filters) are bound.
const baseSQL = `WITH selected_contacts AS NOT MATERIALIZED (
 SELECT c.id,c.team_id,c.list_id,c.email,c.first_name,c.last_name,c.status,c.created_at,c.import_id,c.metadata,lower(trim(c.email)) recipient FROM contacts c
 JOIN mailing_lists l ON l.id=c.list_id AND l.team_id=c.team_id AND l.is_deleted=false
 WHERE c.team_id=@team AND c.is_deleted=false
 AND (@list ='' OR c.list_id::text=@list)
 AND (@tag ='' OR EXISTS(SELECT 1 FROM contact_tags ct JOIN tags t ON t.id=ct.tag_id AND t.team_id=c.team_id AND t.is_deleted=false WHERE ct.contact_id=c.id AND ct.tag_id::text=@tag))
 AND (@campaign ='' OR EXISTS(SELECT 1 FROM campaigns cp WHERE cp.id::text=@campaign AND cp.team_id=c.team_id AND cp.list_id=c.list_id))
), recipients AS (
 SELECT recipient, bool_or(status='ACTIVE') AND NOT EXISTS(SELECT 1 FROM suppression_list s WHERE s.team_id=@team AND lower(trim(s.email_address))=recipient AND s.is_active=true AND s.is_deleted=false AND (s.expires_at IS NULL OR s.expires_at>CAST(@asof AS timestamptz))) active
 FROM selected_contacts GROUP BY recipient
), messages AS (
 SELECT e.id,e.team_id,e.campaign_id,e.contact_id,e.status,e.sent_at,e.created_at,e.click_tracking_enabled,lower(trim(e."to")) recipient,cp.list_id,cp.name campaign_name,cp.newsletter_id,
 COALESCE(NULLIF(lower(split_part(e."from",'@',2)),''),'Unknown') sender_domain
 FROM emails e JOIN campaigns cp ON cp.id=e.campaign_id AND cp.team_id=e.team_id AND cp.is_deleted=false
 WHERE e.team_id=@team AND e.is_deleted=false AND e.test=false
 AND (e.cc IS NULL OR e.cc='') AND (e.bcc IS NULL OR e.bcc='') AND e."to" !~ '[,;<> ]'
 AND (@list ='' OR cp.list_id::text=@list) AND (@campaign ='' OR cp.id::text=@campaign)
 AND (@tag ='' OR EXISTS(SELECT 1 FROM selected_contacts c WHERE c.recipient=lower(trim(e."to")) AND c.list_id=cp.list_id))
), accepted AS (
 SELECT * FROM messages WHERE sent_at>'2000-01-01' AND sent_at<CAST(@asof AS timestamptz) AND status IN ('SENT','OPENED','CLICKED','BOUNCED')
), outcomes AS (
 SELECT email_id,event,timestamp FROM email_trackings WHERE is_deleted=false
 UNION ALL SELECT email_id,'bounce',created_at FROM email_bounces WHERE team_id=@team AND is_deleted=false AND email_id IS NOT NULL
 UNION ALL SELECT email_id,'complaint',created_at FROM complaint_reports WHERE team_id=@team AND is_deleted=false AND email_id IS NOT NULL
), event_facts AS (
 SELECT t.email_id,bool_or(t.event='open') opened,bool_or(t.event='click') clicked,
 bool_or(t.event='bounce') bounced,bool_or(t.event='complaint') complained,bool_or(t.event='unsubscribe') unsubscribed
 FROM outcomes t JOIN accepted e ON e.id=t.email_id
 WHERE e.sent_at>=LEAST(CAST(@from AS timestamptz),CAST(@lookback AS timestamptz)) AND t.timestamp>=e.sent_at AND t.timestamp<CAST(@asof AS timestamptz) GROUP BY t.email_id
), facts AS (
 SELECT e.*,COALESCE(t.opened,false) opened,COALESCE(t.clicked,false) clicked,
 COALESCE(t.bounced,false) OR e.status='BOUNCED' bounced,COALESCE(t.complained,false) complained,COALESCE(t.unsubscribed,false) unsubscribed
 FROM accepted e LEFT JOIN event_facts t ON t.email_id=e.id WHERE e.sent_at>=LEAST(CAST(@from AS timestamptz),CAST(@lookback AS timestamptz))
), click_activity AS (
 SELECT e.recipient,MAX(t.timestamp) last_click FROM accepted e JOIN email_trackings t ON t.email_id=e.id
 WHERE t.is_deleted=false AND t.event='click' AND t.timestamp>=GREATEST(CAST(@lookback AS timestamptz),e.sent_at) AND t.timestamp<CAST(@asof AS timestamptz) GROUP BY e.recipient
), send_activity AS (
 SELECT recipient,COUNT(*) opportunities,COUNT(*) FILTER(WHERE click_tracking_enabled=true) tracked_opportunities
 FROM accepted WHERE sent_at>=CAST(@lookback AS timestamptz) GROUP BY recipient
), cohort_data AS (
 SELECT r.recipient,r.active,CASE
 WHEN c.last_click>=CAST(@recent AS timestamptz) THEN 'recent'
 WHEN c.last_click>=CAST(@lookback AS timestamptz) THEN 'earlier'
 WHEN COALESCE(s.tracked_opportunities,0)>=3 THEN 'no_clicks'
 WHEN COALESCE(s.opportunities,0)=0 THEN 'not_contacted'
 ELSE 'insufficient' END cohort FROM recipients r
 LEFT JOIN click_activity c ON c.recipient=r.recipient LEFT JOIN send_activity s ON s.recipient=r.recipient
) `

const summarySQL = `SELECT
 (SELECT COUNT(*) FROM recipients WHERE active) active,
 (SELECT COUNT(*) FROM selected_contacts WHERE created_at>=CAST(@from AS timestamptz) AND created_at<CAST(@asof AS timestamptz)) new,
 COUNT(DISTINCT recipient) reached,COUNT(DISTINCT recipient) FILTER(WHERE clicked) clicked,
 COUNT(*) accepted,COUNT(*) FILTER(WHERE opened) opened,COUNT(*) FILTER(WHERE clicked) clicked_messages,
 COUNT(*) FILTER(WHERE bounced) bounced,COUNT(*) FILTER(WHERE complained) complaints,COUNT(*) FILTER(WHERE unsubscribed) unsubscribes,
 COUNT(*) FILTER(WHERE click_tracking_enabled IS NOT NULL) tracking_known, COUNT(*) FILTER(WHERE click_tracking_enabled=true) tracking_enabled,
 (SELECT COUNT(*) FROM messages WHERE status='FAILED' AND created_at>=CAST(@from AS timestamptz) AND created_at<CAST(@asof AS timestamptz)) failed,
 (SELECT COUNT(*) FROM messages WHERE status='DELIVERY_UNKNOWN' AND created_at>=CAST(@from AS timestamptz) AND created_at<CAST(@asof AS timestamptz)) unknown
 FROM facts WHERE sent_at>=CAST(@from AS timestamptz)`

func Build(ctx context.Context, db *gorm.DB, f Filter) (Report, error) {
	r := Report{Version: "audience-v2", AsOf: f.AsOf, From: f.From, To: f.To, Timezone: f.Timezone, Activity: []Day{}, Cohorts: []Cohort{}, Growth: Growth{Series: []GrowthDay{}}, Warnings: []string{
		"Accepted means acknowledged by SMTP, not confirmed inbox delivery. Rates exclude test messages and messages with CC/BCC or unresolved multi-recipient addresses.",
		"Clicks and opens are observed tracking events and may include privacy proxies or automated scanners. Missing activity does not prove inactivity.",
		"Historical list and tag filters use current membership. Current active subscribers exclude active workspace suppressions; growth charts measure recorded subscription status, not deliverability eligibility.",
		"Bounce and complaint counts include only captured outcomes. Opt-out attribution is incomplete for older sends; unattributed subscription changes appear in growth, not campaign rates.",
	}}
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(baseSQL+summarySQL, f.args()).Scan(&r.Summary).Error; err != nil {
			return err
		}
		r.Summary.rates()
		prev := f
		prev.From = f.PreviousFrom
		prev.To = f.From
		prev.AsOf = f.From
		if err := tx.Raw(baseSQL+summarySQL, prev.args()).Scan(&r.Previous).Error; err != nil {
			return err
		}
		r.Previous.rates()
		if err := tx.Raw(baseSQL+`, daily_reach AS (SELECT to_char(sent_at AT TIME ZONE @tz,'YYYY-MM-DD') date,COUNT(DISTINCT recipient) reached FROM accepted WHERE sent_at>=CAST(@from AS timestamptz) GROUP BY 1),
 daily_click AS (SELECT to_char(t.timestamp AT TIME ZONE @tz,'YYYY-MM-DD') date,COUNT(DISTINCT e.recipient) clicked FROM accepted e JOIN email_trackings t ON t.email_id=e.id WHERE t.is_deleted=false AND t.event='click' AND t.timestamp>=GREATEST(CAST(@from AS timestamptz),e.sent_at) AND t.timestamp<CAST(@asof AS timestamptz) GROUP BY 1)
 SELECT COALESCE(r.date,c.date) date,COALESCE(r.reached,0) reached,COALESCE(c.clicked,0) clicked FROM daily_reach r FULL JOIN daily_click c ON r.date=c.date ORDER BY 1`, f.args()).Scan(&r.Activity).Error; err != nil {
			return err
		}
		counts := []struct {
			Key   string
			Count int64
		}{}
		if err := tx.Raw(baseSQL+`SELECT cohort key,COUNT(*) count FROM cohort_data WHERE active GROUP BY cohort`, f.args()).Scan(&counts).Error; err != nil {
			return err
		}
		for _, c := range []Cohort{{Key: "recent", Label: "Clicked in 30 days"}, {Key: "earlier", Label: "Clicked 31–90 days ago"}, {Key: "no_clicks", Label: "No clicks after 3+ tracked sends"}, {Key: "not_contacted", Label: "Not contacted in 90 days"}, {Key: "insufficient", Label: "Insufficient tracking evidence"}} {
			for _, n := range counts {
				if c.Key == n.Key {
					c.Count = n.Count
				}
			}
			r.Cohorts = append(r.Cohorts, c)
		}
		if err := loadCharts(tx, f, &r.Charts); err != nil {
			return err
		}
		return loadGrowth(tx, f, &r.Growth)
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return r, err
	}
	byDate := map[string]Day{}
	for _, d := range r.Activity {
		byDate[d.Date] = d
	}
	r.Activity = []Day{}
	loc, _ := time.LoadLocation(f.Timezone)
	for d := f.From.In(loc); d.Before(f.AsOf); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		v := byDate[key]
		v.Date = key
		r.Activity = append(r.Activity, v)
	}
	return r, nil
}

func loadGrowth(db *gorm.DB, f Filter, g *Growth) error {
	var coverage struct{ StartedAt *time.Time }
	if err := db.Raw(`SELECT started_at FROM analytics_coverage WHERE key='subscription-history'`).Scan(&coverage).Error; err != nil {
		return err
	}
	g.Since = coverage.StartedAt
	if g.Since == nil || f.From.Before(*g.Since) || f.Tag != "" {
		return nil
	}
	g.Available = true
	// Latest membership state at each boundary, including tombstones, reconstructs
	// deduplicated subscriber totals without pretending created_at is opt-in time.
	args := f.args()
	sql := `WITH boundaries AS (SELECT CAST(@from AS timestamptz) ::timestamptz at UNION SELECT CAST(@asof AS timestamptz) ::timestamptz UNION SELECT local_day AT TIME ZONE @tz FROM generate_series(date_trunc('day',CAST(@from AS timestamptz) AT TIME ZONE @tz) + interval '1 day', CAST(@asof AS timestamptz) AT TIME ZONE @tz, interval '1 day') local_day)
 SELECT to_char(b.at AT TIME ZONE @tz,'YYYY-MM-DD HH24:MI') date,COUNT(DISTINCT s.email) FILTER(WHERE s.status='ACTIVE' AND s.is_deleted=false AND (@list ='' OR s.list_id::text=@list) AND (@campaign ='' OR s.list_id IN (SELECT list_id FROM campaigns WHERE id::text=@campaign AND team_id=@team))) active
 FROM boundaries b LEFT JOIN LATERAL (SELECT DISTINCT ON(contact_id) email,status,is_deleted,list_id FROM subscription_events WHERE team_id=@team AND occurred_at<=b.at ORDER BY contact_id,occurred_at DESC,id DESC) s ON true
 GROUP BY b.at ORDER BY b.at`
	if err := db.Raw(sql, args).Scan(&g.Series).Error; err != nil {
		return err
	}
	if len(g.Series) > 0 {
		start, end := g.Series[0].Active, g.Series[len(g.Series)-1].Active
		net := end - start
		g.Start = &start
		g.End = &end
		g.Net = &net
	}
	return nil
}

type Row struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Kind              string `json:"kind"`
	Active            int64  `json:"active"`
	Suppressed        int64  `json:"suppressed"`
	Reached           int64  `json:"reached"`
	Clicked           int64  `json:"clicked"`
	Accepted          int64  `json:"accepted"`
	ClickedMessages   int64  `json:"clickedMessages"`
	Bounced           int64  `json:"bounced"`
	Complaints        int64  `json:"complaints"`
	Unsubscribes      int64  `json:"unsubscribes"`
	New               int64  `json:"new"`
	ClickRate         Rate   `json:"clickRate" gorm:"-"`
	AudienceClickRate Rate   `json:"audienceClickRate" gorm:"-"`
}
type Page struct {
	Items []Row     `json:"items"`
	Total int64     `json:"totalCount"`
	Page  int       `json:"page"`
	Limit int       `json:"pageSize"`
	AsOf  time.Time `json:"asOf"`
}

func Breakdown(ctx context.Context, db *gorm.DB, f Filter) (Page, error) {
	p := Page{Items: []Row{}, Page: f.Page, Limit: f.Limit, AsOf: f.AsOf}
	var group string
	switch f.Kind {
	case "lists":
		group = `SELECT l.id::text id,l.name,'list' kind,
 (SELECT COUNT(DISTINCT c.recipient) FROM selected_contacts c JOIN recipients r ON r.recipient=c.recipient WHERE c.list_id=l.id AND c.status='ACTIVE' AND r.active) active,
 (SELECT COUNT(DISTINCT recipient) FROM selected_contacts c WHERE c.list_id=l.id AND c.status<>'ACTIVE') suppressed,
 COUNT(DISTINCT f.recipient) reached, COUNT(DISTINCT f.recipient) FILTER(WHERE f.clicked) clicked,COUNT(f.id) accepted,COUNT(f.id) FILTER(WHERE f.clicked) clicked_messages,COUNT(f.id) FILTER(WHERE f.bounced) bounced,COUNT(f.id) FILTER(WHERE f.complained) complaints,COUNT(f.id) FILTER(WHERE f.unsubscribed) unsubscribes,0::bigint new
 FROM mailing_lists l LEFT JOIN facts f ON f.list_id=l.id AND f.sent_at>=CAST(@from AS timestamptz) WHERE l.team_id=@team AND l.is_deleted=false AND (@list ='' OR l.id::text=@list) AND (@campaign ='' OR l.id IN(SELECT list_id FROM campaigns WHERE id::text=@campaign AND team_id=@team)) AND (@tag ='' OR EXISTS(SELECT 1 FROM selected_contacts c WHERE c.list_id=l.id)) GROUP BY l.id,l.name`
	case "campaigns":
		group = `SELECT cp.id::text id,cp.name,CASE WHEN cp.newsletter_id IS NULL THEN 'campaign' ELSE 'newsletter' END kind,0::bigint active,0::bigint suppressed,COUNT(DISTINCT f.recipient) reached,COUNT(DISTINCT f.recipient) FILTER(WHERE f.clicked) clicked,COUNT(f.id) accepted,COUNT(f.id) FILTER(WHERE f.clicked) clicked_messages,COUNT(f.id) FILTER(WHERE f.bounced) bounced,COUNT(f.id) FILTER(WHERE f.complained) complaints,COUNT(f.id) FILTER(WHERE f.unsubscribed) unsubscribes,0::bigint new FROM campaigns cp JOIN facts f ON f.campaign_id=cp.id AND f.sent_at>=CAST(@from AS timestamptz) GROUP BY cp.id,cp.name,cp.newsletter_id`
	case "domains":
		group = `SELECT sender_domain id,sender_domain name,'domain' kind,0::bigint active,0::bigint suppressed,COUNT(DISTINCT recipient) reached,COUNT(DISTINCT recipient) FILTER(WHERE clicked) clicked,COUNT(*) accepted,COUNT(*) FILTER(WHERE clicked) clicked_messages,COUNT(*) FILTER(WHERE bounced) bounced,COUNT(*) FILTER(WHERE complained) complaints,COUNT(*) FILTER(WHERE unsubscribed) unsubscribes,0::bigint new FROM facts WHERE sent_at>=CAST(@from AS timestamptz) GROUP BY sender_domain`
	case "sources":
		group = `SELECT src id,src name,'source' kind,COUNT(DISTINCT c.recipient) FILTER(WHERE c.status='ACTIVE') active,0::bigint suppressed,COUNT(DISTINCT f.recipient) reached,COUNT(DISTINCT f.recipient) FILTER(WHERE f.clicked) clicked,COUNT(DISTINCT f.id) accepted,COUNT(DISTINCT f.id) FILTER(WHERE f.clicked) clicked_messages,COUNT(DISTINCT f.id) FILTER(WHERE f.bounced) bounced,COUNT(DISTINCT f.id) FILTER(WHERE f.complained) complaints,COUNT(DISTINCT f.id) FILTER(WHERE f.unsubscribed) unsubscribes,COUNT(DISTINCT c.id) new FROM (SELECT c.*,CASE WHEN import_id IS NOT NULL THEN 'Import' WHEN metadata->>'source'='form' THEN 'Form' ELSE 'Unknown' END src FROM selected_contacts c WHERE c.created_at>=CAST(@from AS timestamptz) AND c.created_at<CAST(@asof AS timestamptz)) c LEFT JOIN facts f ON f.recipient=c.recipient AND f.list_id=c.list_id AND f.sent_at>=c.created_at GROUP BY src`
	case "links":
		group = `SELECT t.url id,t.url name,'link' kind,0::bigint active,0::bigint suppressed,COUNT(DISTINCT e.recipient) reached,COUNT(DISTINCT e.recipient) clicked, (SELECT COUNT(*) FROM facts WHERE sent_at>=CAST(@from AS timestamptz)) accepted,COUNT(DISTINCT e.id) clicked_messages,0::bigint bounced,0::bigint complaints,0::bigint unsubscribes,0::bigint new FROM facts e JOIN email_trackings t ON t.email_id=e.id WHERE e.sent_at>=CAST(@from AS timestamptz) AND t.event='click' AND t.is_deleted=false AND t.timestamp>=e.sent_at AND t.timestamp<CAST(@asof AS timestamptz) GROUP BY t.url`
	default:
		return p, fmt.Errorf("Invalid breakdown")
	}
	sorts := map[string]string{"name": "name", "accepted": "accepted", "clicked": "clicked", "active": "active", "bounced": "bounced", "clickRate": "CAST(clicked_messages AS double precision)/NULLIF(accepted,0)", "new": "new"}
	if f.Sort == "" {
		f.Sort = "accepted"
	}
	sort, ok := sorts[f.Sort]
	if !ok {
		return p, fmt.Errorf("Invalid sort")
	}
	query := baseSQL + `, breakdown AS (` + group + `) `
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(query+`SELECT COUNT(*) FROM breakdown`, f.args()).Scan(&p.Total).Error; err != nil {
			return err
		}
		pages := (p.Total + int64(f.Limit) - 1) / int64(f.Limit)
		if pages < 1 {
			pages = 1
		}
		if int64(f.Page) > pages {
			f.Page = int(pages)
			p.Page = f.Page
		}
		return tx.Raw(query+`SELECT * FROM breakdown ORDER BY `+sort+` `+f.Direction+` NULLS LAST,id ASC LIMIT @limit OFFSET @offset`, f.args()).Scan(&p.Items).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	for i := range p.Items {
		r := &p.Items[i]
		r.ClickRate = Ratio(r.ClickedMessages, r.Accepted)
		r.AudienceClickRate = Ratio(r.Clicked, r.Reached)
	}
	return p, err
}

type Person struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	ListID    string `json:"listId"`
	Cohort    string `json:"cohort"`
}
type PeoplePage struct {
	Items []Person `json:"items"`
	Total int64    `json:"totalCount"`
	Page  int      `json:"page"`
	Limit int      `json:"pageSize"`
}

func People(ctx context.Context, db *gorm.DB, f Filter) (PeoplePage, error) {
	p := PeoplePage{Items: []Person{}, Page: f.Page, Limit: f.Limit}
	query := baseSQL + `, people AS (SELECT DISTINCT ON(c.recipient) c.id,c.email,c.first_name,c.last_name,c.list_id,d.cohort FROM selected_contacts c JOIN cohort_data d ON d.recipient=c.recipient AND d.active WHERE c.status='ACTIVE' AND (@cohort ='' OR d.cohort=@cohort) AND lower(c.email||' '||COALESCE(c.first_name,'')||' '||COALESCE(c.last_name,'')) LIKE @search ESCAPE '!' ORDER BY c.recipient,c.id) `
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(query+`SELECT COUNT(*) FROM people`, f.args()).Scan(&p.Total).Error; err != nil {
			return err
		}
		return tx.Raw(query+`SELECT * FROM people ORDER BY email,id LIMIT @limit OFFSET @offset`, f.args()).Scan(&p.Items).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	return p, err
}

// Trends groups accepted sends, not arbitrary tracking-event totals. Rates are fractions.
func Trends(ctx context.Context, db *gorm.DB, f Filter, interval string) ([]map[string]interface{}, error) {
	format := map[string]string{"daily": "YYYY-MM-DD", "weekly": "IYYY-\"W\"IW", "monthly": "YYYY-MM"}[interval]
	if format == "" {
		return nil, fmt.Errorf("Invalid interval")
	}
	type point struct {
		Period                             string
		Accepted, Opened, Clicked, Bounced int64
	}
	points := []point{}
	args := f.args()
	args["format"] = format
	err := db.WithContext(ctx).Raw(baseSQL+`SELECT to_char(sent_at AT TIME ZONE @tz,@format) period,COUNT(*) accepted,COUNT(*) FILTER(WHERE opened) opened,COUNT(*) FILTER(WHERE clicked) clicked,COUNT(*) FILTER(WHERE bounced) bounced FROM facts WHERE sent_at>=CAST(@from AS timestamptz) GROUP BY 1 ORDER BY 1`, args).Scan(&points).Error
	result := []map[string]interface{}{}
	for _, p := range points {
		result = append(result, map[string]interface{}{"period": p.Period, "date": p.Period, "accepted": p.Accepted, "openCount": p.Opened, "clickCount": p.Clicked, "openRate": Ratio(p.Opened, p.Accepted).Value, "clickRate": Ratio(p.Clicked, p.Accepted).Value, "bounceRate": Ratio(p.Bounced, p.Accepted).Value})
	}
	return result, err
}
