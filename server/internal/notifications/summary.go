package notifications

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"kori/internal/analytics"
	"kori/internal/models"
	"kori/internal/onboardingemails"
	"kori/internal/workflowconfig"
)

// Summaries are queued from 08:00 in the workspace time zone: daily ones every
// day for the day before, weekly ones on Monday for Monday to Sunday.
const summaryHour = 8

type Period struct {
	Frequency, Timezone string
	From, To            time.Time
}

func (p Period) days() int {
	if p.Frequency == "daily" {
		return 1
	}
	return 7
}

// Previous is the start of the period before, which has the same length.
func (p Period) Previous() time.Time { return p.From.AddDate(0, 0, -p.days()) }

func (p Period) Reference() string {
	return p.Frequency + "|" + p.From.Format("2006-01-02") + "|" + p.Timezone
}

func ParsePeriod(ref string) (Period, error) {
	parts := strings.Split(ref, "|")
	if len(parts) != 3 || (parts[0] != "daily" && parts[0] != "weekly") || !workflowconfig.ValidTimezone(parts[2]) {
		return Period{}, fmt.Errorf("invalid summary period")
	}
	loc, _ := time.LoadLocation(parts[2])
	from, err := time.ParseInLocation("2006-01-02", parts[1], loc)
	if err != nil {
		return Period{}, err
	}
	p := Period{Frequency: parts[0], Timezone: parts[2], From: from}
	p.To = from.AddDate(0, 0, p.days())
	return p, nil
}

// LastPeriod is the most recent completed period at now, and whether its
// summary is due to be sent.
func LastPeriod(freq, timezone string, now time.Time) (Period, bool) {
	loc := workflowconfig.Location(timezone, "UTC")
	local := now.In(loc)
	p := Period{Frequency: freq, Timezone: loc.String(), To: time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)}
	due := local.Hour() >= summaryHour
	if freq == "weekly" {
		p.To = p.To.AddDate(0, 0, -((int(local.Weekday()) + 6) % 7))
		due = due && local.Weekday() == time.Monday
	} else if freq != "daily" {
		return p, false
	}
	p.From = p.To.AddDate(0, 0, -p.days())
	return p, due
}

// SummaryOwner is the workspace owner, or its earliest admin when no owner is recorded.
func SummaryOwner(db *gorm.DB, team string) (models.User, error) {
	var t models.Team
	if err := db.Select("id", "owner_user_id").Where("id = ? AND is_deleted = false", team).First(&t).Error; err != nil {
		return models.User{}, err
	}
	var owner models.User
	q := db.Where("team_id = ? AND is_deleted = false AND role IN ?", team, []models.UserRole{models.UserRoleAdmin, models.UserRoleSuperAdmin})
	if t.OwnerUserID != "" {
		q = q.Where("id = ?", t.OwnerUserID)
	}
	return owner, q.Order("created_at, id").First(&owner).Error
}

// QueueSummaries queues each due summary for the owner and the admins who
// asked for it. Notice keys make repeated ticks harmless.
func QueueSummaries(ctx context.Context, db *gorm.DB, now time.Time) error {
	db = db.WithContext(ctx)
	var teams []struct{ ID, SummaryFrequency, Timezone string }
	if err := db.Raw(`SELECT t.id, t.summary_frequency, COALESCE(d.default_timezone, '') timezone FROM teams t
 LEFT JOIN team_send_time_defaults d ON d.team_id = t.id AND d.is_deleted = false
 WHERE t.is_deleted = false AND t.summary_frequency IN ('daily', 'weekly')`).Scan(&teams).Error; err != nil {
		return err
	}
	for _, t := range teams {
		p, due := LastPeriod(t.SummaryFrequency, t.Timezone, now)
		if !due {
			continue
		}
		owner, err := SummaryOwner(db, t.ID)
		if err != nil {
			continue
		}
		var admins []models.User
		if err := db.Where("team_id = ? AND is_deleted = false AND summary_emails = true AND id <> ? AND role IN ?", t.ID, owner.ID, []models.UserRole{models.UserRoleAdmin, models.UserRoleSuperAdmin}).Find(&admins).Error; err != nil {
			return err
		}
		for _, u := range append([]models.User{owner}, admins...) {
			if err := models.QueueSummaryNotice(db, u, p.Reference(), now); err != nil {
				return err
			}
		}
	}
	return nil
}

type Numbers struct {
	Sent, Bounced, Complaints, NewContacts, Unsubscribes int64
	OpenRate, ClickRate, BounceRate                      analytics.Rate
}

type Summary struct {
	Period
	Now, Before Numbers
	FirstTime   int64 // people who got their first email from the workspace in the period
	Changes     int64 // subscription changes in the period, whatever their source
	Top         *analytics.Row
}

// BuildSummary reads the same reports as the dashboard Overview, so the email
// and the dashboard show the same numbers for the same dates.
func BuildSummary(ctx context.Context, db *gorm.DB, team string, p Period) (Summary, error) {
	s := Summary{Period: p}
	f := analytics.Filter{Team: team, Timezone: p.Timezone, From: p.From, To: p.To, AsOf: p.To, PreviousFrom: p.Previous(), Page: 1, Limit: 1, Kind: "campaigns", Sort: "accepted", Direction: "desc"}
	before := f
	before.From, before.To, before.AsOf = p.Previous(), p.From, p.From
	email, err := analytics.BuildEmailReport(ctx, db, f)
	if err != nil {
		return s, err
	}
	earlier, err := analytics.BuildEmailReport(ctx, db, before)
	if err != nil {
		return s, err
	}
	audience, err := analytics.Build(ctx, db, f)
	if err != nil {
		return s, err
	}
	top, err := analytics.Breakdown(ctx, db, f)
	if err != nil {
		return s, err
	}
	if len(top.Items) > 0 && top.Items[0].Accepted > 0 {
		s.Top = &top.Items[0]
	}
	s.Now = Numbers{Sent: email.Summary.Sent, Bounced: email.Summary.Bounced, Complaints: audience.Summary.Complaints, NewContacts: audience.Summary.New, Unsubscribes: audience.Summary.Unsubscribes, OpenRate: email.OpenRate, ClickRate: email.ClickRate, BounceRate: email.BounceRate}
	s.Before = Numbers{Sent: earlier.Summary.Sent, Bounced: earlier.Summary.Bounced, Complaints: audience.Previous.Complaints, NewContacts: audience.Previous.New, Unsubscribes: audience.Previous.Unsubscribes, OpenRate: earlier.OpenRate, ClickRate: earlier.ClickRate, BounceRate: earlier.BounceRate}
	accepted := `team_id = @team AND is_deleted = false AND test = false AND status IN ('SENT','OPENED','CLICKED','BOUNCED')`
	args := map[string]any{"team": team, "from": p.From, "to": p.To}
	// ponytail: scans the workspace's whole send history once a period; keep a first-sent table if that gets slow.
	if err := db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM (SELECT lower(trim("to")) FROM emails WHERE `+accepted+` AND sent_at >= @from AND sent_at < @to
 EXCEPT SELECT lower(trim("to")) FROM emails WHERE `+accepted+` AND sent_at > '2000-01-01' AND sent_at < @from) first_time`, args).Scan(&s.FirstTime).Error; err != nil {
		return s, err
	}
	err = db.WithContext(ctx).Raw(`SELECT COUNT(*) FROM subscription_events WHERE team_id = @team AND occurred_at >= @from AND occurred_at < @to AND kind <> 'baseline'`, args).Scan(&s.Changes).Error
	return s, err
}

// Quiet periods had no sends and no subscriber changes; their summary is skipped.
func (s Summary) Quiet() bool { return s.Now.Sent == 0 && s.Changes == 0 && s.Now.NewContacts == 0 }

func (s Summary) words() (period, before string) {
	if s.Frequency == "daily" {
		return "yesterday", "the day before"
	}
	return "last week", "the week before"
}

// points is the change between two rates in whole percentage points; ok is
// false when either period had nothing to measure.
func points(now, before analytics.Rate) (int, bool) {
	if now.Value == nil || before.Value == nil {
		return 0, false
	}
	return int(math.Round((*now.Value - *before.Value) * 100)), true
}

func plural(n int64, one, many string) string {
	if n == 1 || n == -1 {
		return one
	}
	return many
}

// Headline is one plain sentence on what changed most and, when the numbers
// show it, the likely reason. Health problems come first, then engagement,
// then volume. It never guesses: a change without a supporting number gets no reason.
func (s Summary) Headline() string {
	period, before := s.words()
	firstTime := ""
	if s.FirstTime > 0 && s.FirstTime*5 >= s.Now.Sent {
		firstTime = fmt.Sprintf("; %s %s got their first email from you", commas(s.FirstTime), plural(s.FirstTime, "person", "people"))
	}
	if s.Now.Complaints >= 3 && s.Now.Complaints > s.Before.Complaints {
		return fmt.Sprintf("Spam complaints rose to %s %s, from %s %s%s.", commas(s.Now.Complaints), period, commas(s.Before.Complaints), before, firstTime)
	}
	if pts, ok := points(s.Now.BounceRate, s.Before.BounceRate); ok && pts >= 2 {
		return fmt.Sprintf("Bounces rose %d points to %s of sends%s.", pts, percent(s.Now.BounceRate), firstTime)
	}
	for _, m := range []struct {
		name      string
		now, prev analytics.Rate
		min       int
	}{{"Opens", s.Now.OpenRate, s.Before.OpenRate, 3}, {"Clicks", s.Now.ClickRate, s.Before.ClickRate, 1}} {
		pts, ok := points(m.now, m.prev)
		if !ok || (pts < m.min && pts > -m.min) {
			continue
		}
		if pts > 0 {
			return fmt.Sprintf("%s rose %d %s to %s.", m.name, pts, plural(int64(pts), "point", "points"), percent(m.now))
		}
		reason := firstTime
		if reason == "" && s.Before.Sent > 0 && s.Now.Sent*2 >= s.Before.Sent*3 {
			reason = fmt.Sprintf("; you sent %d%% more email than %s", (s.Now.Sent-s.Before.Sent)*100/s.Before.Sent, before)
		}
		return fmt.Sprintf("%s fell %d %s to %s%s.", m.name, -pts, plural(int64(pts), "point", "points"), percent(m.now), reason)
	}
	if s.Now.Sent == 0 {
		if s.Now.NewContacts > 0 {
			return fmt.Sprintf("No emails went out %s, and %s new %s joined.", period, commas(s.Now.NewContacts), plural(s.Now.NewContacts, "contact", "contacts"))
		}
		return fmt.Sprintf("No emails went out %s, but your subscriber lists changed.", period)
	}
	if s.Before.Sent == 0 {
		return fmt.Sprintf("You sent %s %s %s, after none %s.", commas(s.Now.Sent), plural(s.Now.Sent, "email", "emails"), period, before)
	}
	change := (s.Now.Sent - s.Before.Sent) * 100 / s.Before.Sent
	if change >= 25 {
		reason := ""
		if s.Top != nil && s.Top.Accepted*2 >= s.Now.Sent {
			reason = fmt.Sprintf("; %s went to %s %s", s.Top.Name, commas(s.Top.Reached), plural(s.Top.Reached, "person", "people"))
		}
		return fmt.Sprintf("You sent %d%% more email than %s%s.", change, before, reason)
	}
	if change <= -25 {
		return fmt.Sprintf("You sent %d%% less email than %s.", -change, before)
	}
	return fmt.Sprintf("You sent %s %s %s, about as many as %s.", commas(s.Now.Sent), plural(s.Now.Sent, "email", "emails"), period, before)
}

func commas(n int64) string {
	s := strconv.FormatInt(n, 10)
	if n < 0 {
		return "-" + commas(-n)
	}
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func percent(r analytics.Rate) string {
	if r.Value == nil {
		return "—"
	}
	return fmt.Sprintf("%.1f%%", *r.Value*100)
}

func (s Summary) count(now, before int64) string {
	_, word := s.words()
	switch {
	case before == 0 && now == 0:
		return "0"
	case before == 0:
		return commas(now) + " (none " + word + ")"
	case now == before:
		return commas(now) + " (same as " + word + ")"
	case now > before:
		return fmt.Sprintf("%s (up %d%% from %s)", commas(now), (now-before)*100/before, commas(before))
	}
	return fmt.Sprintf("%s (down %d%% from %s)", commas(now), (before-now)*100/before, commas(before))
}

func (s Summary) rate(now, before analytics.Rate) string {
	if now.Value == nil {
		return "No tracked emails"
	}
	if before.Value == nil {
		return percent(now)
	}
	diff := (*now.Value - *before.Value) * 100
	direction := "up"
	if diff < 0 {
		direction, diff = "down", -diff
	}
	if math.Round(diff*10) == 0 {
		return percent(now) + " (same as before)"
	}
	return fmt.Sprintf("%s (%s %.1f points from %s)", percent(now), direction, diff, percent(before))
}

func (s Summary) Email(workspace string) onboardingemails.SummaryEmail {
	dates := s.From.Format("Mon 2 Jan 2006")
	if s.Frequency == "weekly" {
		dates = s.From.Format("Mon 2 Jan") + " – " + s.To.AddDate(0, 0, -1).Format("Mon 2 Jan 2006")
	}
	details := [][2]string{
		{"Workspace", workspace},
		{"Period", dates + " (" + s.Timezone + ")"},
		{"Emails sent", s.count(s.Now.Sent, s.Before.Sent)},
		{"Opened", s.rate(s.Now.OpenRate, s.Before.OpenRate)},
		{"Clicked", s.rate(s.Now.ClickRate, s.Before.ClickRate)},
		{"Bounces", s.count(s.Now.Bounced, s.Before.Bounced)},
		{"Spam complaints", s.count(s.Now.Complaints, s.Before.Complaints)},
		{"New contacts", s.count(s.Now.NewContacts, s.Before.NewContacts)},
		{"Unsubscribes", s.count(s.Now.Unsubscribes, s.Before.Unsubscribes)},
	}
	if s.Top != nil {
		details = append(details, [2]string{"Top campaign", fmt.Sprintf("%s: %s sent, %s clicked", s.Top.Name, commas(s.Top.Accepted), percent(s.Top.ClickRate))})
	}
	return onboardingemails.SummaryEmail{Frequency: s.Frequency, Workspace: workspace, Headline: s.Headline(), Details: details}
}

// summary renders a queued summary, or errSkip when the period was quiet or
// the workspace no longer wants this frequency.
func (w *Worker) summary(ctx context.Context, n models.ServiceNotice, workspace string) (onboardingemails.SummaryEmail, error) {
	p, err := ParsePeriod(n.ReferenceID)
	if err != nil {
		return onboardingemails.SummaryEmail{}, errSkip
	}
	var frequency string
	if err := w.DB.WithContext(ctx).Model(&models.Team{}).Where("id = ?", n.TeamID).Pluck("summary_frequency", &frequency).Error; err != nil {
		return onboardingemails.SummaryEmail{}, err
	}
	if frequency != p.Frequency {
		return onboardingemails.SummaryEmail{}, errSkip
	}
	s, err := BuildSummary(ctx, w.DB, n.TeamID, p)
	if err != nil {
		return onboardingemails.SummaryEmail{}, err
	}
	if s.Quiet() {
		return onboardingemails.SummaryEmail{}, errSkip
	}
	return s.Email(workspace), nil
}
