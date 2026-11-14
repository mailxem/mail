package models

import "time"

// Author is someone a workspace's campaigns and newsletter issues are signed
// by. Readers see "By <name>" with the photo, and replies go to Email unless
// a campaign sets its own Reply-To.
type Author struct {
	Base
	TeamID string `gorm:"type:uuid;not null;index" json:"teamId"`
	Name   string `gorm:"not null" json:"name"`
	// PhotoURL is a square, public HTTPS image; "" shows no photo.
	PhotoURL string `json:"photoUrl"`
	Email    string `json:"email"`
	// BylineEmail and BylineWeb show the byline at the top of emails and
	// under the issue's title on the newsletter website.
	BylineEmail bool `gorm:"not null" json:"bylineEmail"`
	BylineWeb   bool `gorm:"not null" json:"bylineWeb"`
	// IsDefault signs campaigns and issues that pick no author.
	IsDefault bool `gorm:"not null" json:"isDefault"`
}

// IssuePage is how a sent issue reads on the newsletter website once its team
// edited it. The campaign keeps the subject and body that were emailed.
type IssuePage struct {
	CampaignID  string    `gorm:"type:uuid;primaryKey" json:"campaignId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Cover       string    `json:"cover"`
	HTML        string    `gorm:"type:text" json:"html"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
