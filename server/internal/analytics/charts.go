package analytics

import (
	"gorm.io/gorm"
	"time"
)

type VolumeDay struct {
	Date             string `json:"date"`
	Campaigns        int64  `json:"campaigns"`
	Newsletters      int64  `json:"newsletters"`
	CampaignClicks   int64  `json:"campaignClicks"`
	NewsletterClicks int64  `json:"newsletterClicks"`
}
type DeviceCount struct {
	Device string `json:"device"`
	Count  int64  `json:"count"`
}
type ClickHour struct {
	Day   int   `json:"day"`
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}
type Charts struct {
	Volume      []VolumeDay   `json:"volume"`
	Devices     []DeviceCount `json:"devices"`
	ClickHours  []ClickHour   `json:"clickHours"`
	ClickEvents int64         `json:"clickEvents"`
}

// Send-cohort bars reconcile to summary message totals. Device and hour buckets
// count the same observed click events, including repeat clicks and older sends.
func loadCharts(db *gorm.DB, f Filter, c *Charts) error {
	c.Volume = []VolumeDay{}
	c.Devices = []DeviceCount{}
	c.ClickHours = []ClickHour{}
	if err := db.Raw(baseSQL+`SELECT to_char(sent_at AT TIME ZONE @tz,'YYYY-MM-DD') date,
 COUNT(*) FILTER(WHERE newsletter_id IS NULL) campaigns,
 COUNT(*) FILTER(WHERE newsletter_id IS NOT NULL) newsletters,
 COUNT(*) FILTER(WHERE newsletter_id IS NULL AND clicked) campaign_clicks,
 COUNT(*) FILTER(WHERE newsletter_id IS NOT NULL AND clicked) newsletter_clicks
 FROM facts WHERE sent_at>=CAST(@from AS timestamptz) GROUP BY 1 ORDER BY 1`, f.args()).Scan(&c.Volume).Error; err != nil {
		return err
	}
	type bucket struct {
		Day, Hour int
		Device    string
		Count     int64
	}
	buckets := []bucket{}
	if err := db.Raw(baseSQL+`SELECT (extract(isodow FROM t.timestamp AT TIME ZONE @tz)::int-1) AS "day",
 extract(hour FROM t.timestamp AT TIME ZONE @tz)::int AS "hour",
 CASE WHEN lower(t.device_type) IN ('desktop','mobile','tablet','other') THEN lower(t.device_type) ELSE 'unknown' END device,
 COUNT(*) count FROM accepted e JOIN email_trackings t ON t.email_id=e.id
 WHERE t.is_deleted=false AND t.event='click' AND t.timestamp>=GREATEST(CAST(@from AS timestamptz),e.sent_at)
 AND t.timestamp<CAST(@asof AS timestamptz) GROUP BY 1,2,3`, f.args()).Scan(&buckets).Error; err != nil {
		return err
	}
	deviceCounts := map[string]int64{}
	hours := [7][24]int64{}
	for _, b := range buckets {
		deviceCounts[b.Device] += b.Count
		hours[b.Day][b.Hour] += b.Count
		c.ClickEvents += b.Count
	}
	for _, device := range []string{"desktop", "mobile", "tablet", "other", "unknown"} {
		c.Devices = append(c.Devices, DeviceCount{Device: device, Count: deviceCounts[device]})
	}
	for day := 0; day < 7; day++ {
		for hour := 0; hour < 24; hour++ {
			c.ClickHours = append(c.ClickHours, ClickHour{Day: day, Hour: hour, Count: hours[day][hour]})
		}
	}
	byDate := map[string]VolumeDay{}
	for _, day := range c.Volume {
		byDate[day.Date] = day
	}
	c.Volume = []VolumeDay{}
	loc, _ := time.LoadLocation(f.Timezone)
	for day := f.From.In(loc); day.Before(f.AsOf); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		value := byDate[key]
		value.Date = key
		c.Volume = append(c.Volume, value)
	}
	return nil
}
