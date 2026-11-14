package handlers

import (
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"kori/internal/events"
	"kori/internal/marketing"
	"kori/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// recipientRow is one recipient of a campaign and what they did. Opens and
// clicks are people's only: machine events are left out, as in the report.
type recipientRow struct {
	EmailID   string `json:"emailId"`
	ContactID string `json:"contactId,omitempty"`
	Email     string `json:"email"`
	// Delivery is waiting, received, bounced or not_delivered.
	Delivery       string     `json:"delivery"`
	Error          string     `json:"error,omitempty"`
	Tracked        bool       `json:"tracked"`
	SentAt         *time.Time `json:"sentAt,omitempty"`
	BouncedAt      *time.Time `json:"bouncedAt,omitempty"`
	OpenedAt       *time.Time `json:"openedAt,omitempty"`
	Opens          int        `json:"opens"`
	ClickedAt      *time.Time `json:"clickedAt,omitempty"`
	Links          []string   `json:"links,omitempty"`
	Spam           bool       `json:"spam"`
	SpamAt         *time.Time `json:"spamAt,omitempty"`
	UnsubscribedAt *time.Time `json:"unsubscribedAt,omitempty"`
}

// recipientEmail is the part of an email the recipient report reads.
type recipientEmail struct {
	ID                   string
	ContactID            string
	To                   string
	Status               string
	Error                string
	SentAt               time.Time
	ClickTrackingEnabled *bool
}

// recipientFilters are the outcomes the recipient list can be narrowed to.
var recipientFilters = []string{"received", "opened", "not_opened", "clicked", "waiting", "not_delivered", "bounced", "spam", "unsubscribed"}

// recipientOutcomes joins a campaign's emails with their tracking events.
// A bounce wins over any delivered status, and an open or click shows a
// waiting email arrived.
func recipientOutcomes(emails []recipientEmail, tracking []models.EmailTracking) []recipientRow {
	rows := make([]recipientRow, len(emails))
	byEmail := make(map[string]*recipientRow, len(emails))
	for i, e := range emails {
		rows[i] = recipientRow{EmailID: e.ID, ContactID: e.ContactID, Email: e.To, Error: e.Error, Tracked: e.ClickTrackingEnabled == nil || *e.ClickTrackingEnabled, Spam: e.Status == "COMPLAINED"}
		if !e.SentAt.IsZero() {
			rows[i].SentAt = &e.SentAt
		}
		byEmail[e.ID] = &rows[i]
	}
	first := func(at **time.Time, t time.Time) {
		if *at == nil || t.Before(**at) {
			*at = &t
		}
	}
	for _, t := range tracking {
		r := byEmail[t.EmailID]
		if r == nil || !t.Human() {
			continue
		}
		switch t.Event {
		case models.EmailTrackingEventOpen:
			r.Opens++
			first(&r.OpenedAt, t.Timestamp)
		case models.EmailTrackingEventClick:
			first(&r.ClickedAt, t.Timestamp)
			if !slices.Contains(r.Links, t.URL) {
				r.Links = append(r.Links, t.URL)
			}
		case models.EmailTrackingEventBounce:
			first(&r.BouncedAt, t.Timestamp)
		case models.EmailTrackingEventComplaint:
			r.Spam = true
			first(&r.SpamAt, t.Timestamp)
		case models.EmailTrackingEventUnsubscribe:
			first(&r.UnsubscribedAt, t.Timestamp)
		}
	}
	for i, e := range emails {
		r := &rows[i]
		switch {
		case e.Status == "BOUNCED" || r.BouncedAt != nil:
			r.Delivery = "bounced"
		case r.OpenedAt != nil || r.ClickedAt != nil || r.Spam:
			r.Delivery = "received"
		case e.Status == "FAILED" || e.Status == "SUPPRESSED":
			r.Delivery = "not_delivered"
		case e.Status == "PENDING" || e.Status == "QUEUED" || e.Status == "SENDING" || e.Status == "ACCEPTED" || e.Status == "DELIVERY_UNKNOWN":
			r.Delivery = "waiting"
		default:
			r.Delivery = "received"
		}
	}
	return rows
}

// matches reports whether r has the outcome filter names; url narrows
// "clicked" to one link.
func (r recipientRow) matches(filter, url string) bool {
	switch filter {
	case "":
		return true
	case "received", "waiting", "not_delivered", "bounced":
		return r.Delivery == filter
	case "opened":
		return r.OpenedAt != nil
	case "not_opened":
		// Untracked emails can't tell, so they are left out.
		return r.Delivery == "received" && r.Tracked && r.OpenedAt == nil && r.ClickedAt == nil
	case "clicked":
		return r.ClickedAt != nil && (url == "" || slices.Contains(r.Links, url))
	case "spam":
		return r.Spam
	case "unsubscribed":
		return r.UnsubscribedAt != nil
	}
	return false
}

// hourBucket counts the people who first opened, and first clicked, in one hour.
type hourBucket struct {
	Hour   time.Time `json:"hour"`
	Opens  int       `json:"opens"`
	Clicks int       `json:"clicks"`
}

// maxTimelineHours is how far the timeline reaches: a week after sending.
const maxTimelineHours = 7 * 24

// hourlyEngagement buckets each person's first open and first click by the
// hour since start, through now and at most maxTimelineHours. Later ones are
// counted in later, so the bars and later add up to the report's totals.
func hourlyEngagement(start, now time.Time, tracking []models.EmailTracking) (buckets []hourBucket, later hourBucket) {
	start = start.Truncate(time.Hour)
	n := min(max(int(now.Sub(start)/time.Hour)+1, 1), maxTimelineHours)
	buckets = make([]hourBucket, n)
	for i := range buckets {
		buckets[i].Hour = start.Add(time.Duration(i) * time.Hour)
	}
	opened, clicked := map[string]time.Time{}, map[string]time.Time{}
	for _, t := range tracking {
		if !t.Human() {
			continue
		}
		seen := map[models.EmailTrackingEvent]map[string]time.Time{models.EmailTrackingEventOpen: opened, models.EmailTrackingEventClick: clicked}[t.Event]
		if seen == nil {
			continue
		}
		if at, ok := seen[t.EmailID]; !ok || t.Timestamp.Before(at) {
			seen[t.EmailID] = t.Timestamp
		}
	}
	add := func(at time.Time, count func(*hourBucket)) {
		i := max(int(at.Sub(start)/time.Hour), 0)
		if i < n {
			count(&buckets[i])
		} else {
			count(&later)
		}
	}
	for _, at := range opened {
		add(at, func(b *hourBucket) { b.Opens++ })
	}
	for _, at := range clicked {
		add(at, func(b *hourBucket) { b.Clicks++ })
	}
	return buckets, later
}

// campaignRecipients loads a workspace campaign's recipients and outcomes.
// ponytail: loads every email and tracking row; page in SQL if a campaign passes ~100k recipients.
func (h *MarketingHandler) campaignRecipients(c echo.Context) (models.Campaign, []recipientRow, []models.EmailTracking, error) {
	var campaign models.Campaign
	if err := h.scoped(c).First(&campaign, "id = ?", c.Param("id")).Error; err != nil {
		return campaign, nil, nil, missing()
	}
	db := h.service.DB.WithContext(c.Request().Context())
	sent := db.Model(&models.Email{}).Where("campaign_id = ? AND team_id = ? AND is_deleted = ? AND test = ?", campaign.ID, campaign.TeamID, false, false)
	var emails []recipientEmail
	if err := sent.Session(&gorm.Session{}).Select(`id, contact_id, "to", status, error, sent_at, click_tracking_enabled`).Order(`"to", id`).Scan(&emails).Error; err != nil {
		return campaign, nil, nil, err
	}
	var tracking []models.EmailTracking
	if err := db.Where("is_deleted = ? AND email_id IN (?)", false, sent.Session(&gorm.Session{}).Select("id")).Order("timestamp, id").Find(&tracking).Error; err != nil {
		return campaign, nil, nil, err
	}
	return campaign, recipientOutcomes(emails, tracking), tracking, nil
}

// CampaignRecipients lists a campaign's recipients, searchable by address and
// filterable by outcome, with how many recipients have each outcome.
func (h *MarketingHandler) CampaignRecipients(c echo.Context) error {
	_, rows, _, err := h.campaignRecipients(c)
	if err != nil {
		return err
	}
	filter, url, q := c.QueryParam("filter"), c.QueryParam("url"), strings.ToLower(strings.TrimSpace(c.QueryParam("q")))
	if filter != "" && !slices.Contains(recipientFilters, filter) {
		return bad("Unknown filter")
	}
	counts := map[string]int{"all": len(rows)}
	var links []string
	for _, r := range rows {
		for _, f := range recipientFilters {
			if r.matches(f, "") {
				counts[f]++
			}
		}
		for _, l := range r.Links {
			if !slices.Contains(links, l) {
				links = append(links, l)
			}
		}
	}
	slices.Sort(links)
	matched := filterRecipients(rows, filter, url, q)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	page, limit = max(page, 1), min(max(limit, 1), 200)
	if c.QueryParam("limit") == "" {
		limit = 50
	}
	from := min((page-1)*limit, len(matched))
	return c.JSON(200, map[string]any{
		"recipients": matched[from:min(from+limit, len(matched))],
		"total":      len(matched),
		"page":       page,
		"limit":      limit,
		"counts":     counts,
		"links":      links,
	})
}

func filterRecipients(rows []recipientRow, filter, url, q string) []recipientRow {
	matched := []recipientRow{}
	for _, r := range rows {
		if r.matches(filter, url) && (q == "" || strings.Contains(strings.ToLower(r.Email), q)) {
			matched = append(matched, r)
		}
	}
	return matched
}

// CampaignTimeline returns opens and clicks, by people, hour by hour since the campaign started sending.
func (h *MarketingHandler) CampaignTimeline(c echo.Context) error {
	_, rows, tracking, err := h.campaignRecipients(c)
	if err != nil {
		return err
	}
	var start time.Time
	for _, r := range rows {
		if r.SentAt != nil && (start.IsZero() || r.SentAt.Before(start)) {
			start = *r.SentAt
		}
	}
	if start.IsZero() {
		return c.JSON(200, map[string]any{"buckets": []hourBucket{}, "later": hourBucket{}})
	}
	buckets, later := hourlyEngagement(start, time.Now(), tracking)
	return c.JSON(200, map[string]any{"buckets": buckets, "later": later})
}

// recipientAudience is a filtered recipient view to save.
type recipientAudience struct {
	Name   string `json:"name"`
	Filter string `json:"filter"`
	URL    string `json:"url"`
	Q      string `json:"q"`
}

// saveRecipientSegment saves the contacts of a filtered recipient view as a
// static segment. Contacts deleted since sending are left out.
func saveRecipientSegment(tx *gorm.DB, campaign models.Campaign, rows []recipientRow, in recipientAudience) (models.Segment, error) {
	in.Name = strings.TrimSpace(in.Name)
	if n := utf8.RuneCountInString(in.Name); n < 2 || n > 100 {
		return models.Segment{}, bad("Give the segment a name of 2 to 100 characters")
	}
	if in.Filter != "" && !slices.Contains(recipientFilters, in.Filter) {
		return models.Segment{}, bad("Unknown filter")
	}
	var ids []string
	for _, r := range filterRecipients(rows, in.Filter, in.URL, strings.ToLower(strings.TrimSpace(in.Q))) {
		if r.ContactID != "" && !slices.Contains(ids, r.ContactID) {
			ids = append(ids, r.ContactID)
		}
	}
	var contacts []string
	for i := 0; i < len(ids); i += 500 {
		var page []string
		if err := marketing.Scope(tx, campaign.TeamID).Model(&models.Contact{}).Where("id IN ?", ids[i:min(i+500, len(ids))]).Pluck("id", &page).Error; err != nil {
			return models.Segment{}, err
		}
		contacts = append(contacts, page...)
	}
	if len(contacts) == 0 {
		return models.Segment{}, bad("No contacts match this view")
	}
	now := time.Now()
	segment := models.Segment{
		TeamID: campaign.TeamID, Name: in.Name, Description: "Recipients of the campaign " + campaign.Name,
		Rules: datatypes.JSON(`{"rules":[],"matchType":"ALL"}`), MatchType: models.MatchTypeAll,
		IsDynamic: false, ContactCount: len(contacts), LastRefresh: &now,
	}
	if err := tx.Omit(clause.Associations).Create(&segment).Error; err != nil {
		return segment, err
	}
	// Create skips false for a column that defaults to true; a dynamic segment
	// with no rules would refresh to every contact.
	if err := tx.Model(&segment).Update("is_dynamic", false).Error; err != nil {
		return segment, err
	}
	members := make([]models.SegmentContact, len(contacts))
	for i, id := range contacts {
		members[i] = models.SegmentContact{SegmentID: segment.ID, ContactID: id, AddedAt: now}
	}
	return segment, tx.CreateInBatches(&members, 500).Error
}

// SaveRecipientSegment saves a filtered recipient view as a segment.
func (h *MarketingHandler) SaveRecipientSegment(c echo.Context) error {
	var in recipientAudience
	if err := decodeMarketingInput(c, &in); err != nil {
		return err
	}
	campaign, rows, _, err := h.campaignRecipients(c)
	if err != nil {
		return err
	}
	var segment models.Segment
	if err := h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) (err error) {
		segment, err = saveRecipientSegment(tx, campaign, rows, in)
		return err
	}); err != nil {
		return err
	}
	return c.JSON(201, segment)
}

// CreateFollowUp saves a filtered recipient view as a segment and drafts a
// copy of the campaign that goes only to it. The draft never sends until
// SendCampaignDraft.
func (h *MarketingHandler) CreateFollowUp(c echo.Context) error {
	var in recipientAudience
	if err := decodeMarketingInput(c, &in); err != nil {
		return err
	}
	campaign, rows, _, err := h.campaignRecipients(c)
	if err != nil {
		return err
	}
	var draft models.Campaign
	err = h.service.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		segment, err := saveRecipientSegment(tx, campaign, rows, in)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		draft = models.Campaign{
			Base: models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now}, TeamID: campaign.TeamID,
			Name: "Follow-up: " + campaign.Name, Subject: campaign.Subject, Description: campaign.Description,
			TemplateID: campaign.TemplateID, ListID: campaign.ListID, SegmentID: &segment.ID, SMTPConfigID: campaign.SMTPConfigID,
			HTMLBody: campaign.HTMLBody, PostalAddress: campaign.PostalAddress, ReplyTo: campaign.ReplyTo,
			TrackEngagement: campaign.TrackEngagement, SendInStages: campaign.SendInStages, Data: campaign.Data,
			Status: models.CampaignStatusDraft, Schedule: models.CampaignScheduleOneTime, Timezone: campaign.Timezone,
			BatchSize: campaign.BatchSize, BatchDelay: campaign.BatchDelay,
		}
		// Hooks would schedule it; a draft waits for SendCampaignDraft.
		return tx.Session(&gorm.Session{SkipHooks: true}).Omit(clause.Associations).Create(&draft).Error
	})
	if err != nil {
		return err
	}
	return c.JSON(201, draft)
}

// SendCampaignDraft sends a draft campaign now.
func (h *MarketingHandler) SendCampaignDraft(c echo.Context) error {
	var campaign models.Campaign
	if err := h.scoped(c).Where("id = ? AND status = ? AND newsletter_id IS NULL", c.Param("id"), models.CampaignStatusDraft).First(&campaign).Error; err != nil {
		return missing()
	}
	now := time.Now().UTC()
	result := h.scoped(c).Model(&models.Campaign{}).Where("id = ? AND status = ?", campaign.ID, models.CampaignStatusDraft).
		Updates(map[string]any{"status": models.CampaignStatusScheduled, "scheduled_for": now})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return missing()
	}
	campaign.Status, campaign.ScheduledFor = models.CampaignStatusScheduled, now
	events.Emit("campaign.created", &campaign) // queues the send, as on launch
	return c.JSON(200, map[string]string{"status": string(models.CampaignStatusScheduled)})
}
