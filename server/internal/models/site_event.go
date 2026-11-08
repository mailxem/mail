package models

import "time"

// SiteEvent is one page view or click on a newsletter website. There are no
// cookies and no IP addresses: Session is a random id kept in the visitor's
// browser tab, so it only tells views in one visit apart.
type SiteEvent struct {
	ID           uint      `gorm:"primaryKey"`
	NewsletterID string    `gorm:"type:uuid;not null;index:idx_site_events_newsletter_time,priority:1"`
	CreatedAt    time.Time `gorm:"not null;index:idx_site_events_newsletter_time,priority:2"`
	Kind         string    `gorm:"size:8;not null"`
	Path         string    `gorm:"size:200;not null"`
	Referrer     string    `gorm:"size:100"`
	Session      string    `gorm:"size:24"`
	Device       string    `gorm:"size:8"`
	// Link is the address of a clicked link, if any.
	Link string `gorm:"size:200"`
	// X is how far across the page a click was (0 to 1); Y is its distance
	// from the top in CSS pixels, and Height the page's height at the time.
	X      float32
	Y      int32
	Height int32
}
