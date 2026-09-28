package models

import (
	"time"

	"gorm.io/gorm"
)

// ScoreGrade represents lead quality grades
type ScoreGrade string

const (
	GradeA ScoreGrade = "A" // 80-100 - Hot lead, sales ready
	GradeB ScoreGrade = "B" // 60-79 - Warm lead, nurture more
	GradeC ScoreGrade = "C" // 40-59 - Cold lead, engage
	GradeD ScoreGrade = "D" // 20-39 - Very cold, re-engagement campaign
	GradeF ScoreGrade = "F" // 0-19 - Inactive, consider removing
)

// LeadScore represents the current score for a contact
type LeadScore struct {
	ContactID      string     `gorm:"type:uuid;primaryKey"`
	Contact        *Contact   `json:"contact,omitempty"`
	Score          int        `gorm:"not null;default:0;index"`
	Grade          ScoreGrade `gorm:"type:varchar(1);not null;default:'F';index"`
	LastActivityAt *time.Time `json:"lastActivityAt"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

// ScoreActivity represents the audit log of score changes
type ScoreActivity struct {
	Base
	ContactID    string    `gorm:"type:uuid;not null;index"`
	Contact      *Contact  `json:"contact,omitempty"`
	ActivityType string    `gorm:"not null;index"` // email_opened, link_clicked, etc.
	Points       int       `gorm:"not null"`       // Can be negative for decay
	Description  string    `json:"description"`
	AutomationID *string   `gorm:"type:uuid;index"`
	Automation   *Automation `json:"automation,omitempty"`
	CampaignID   *string   `gorm:"type:uuid;index"`
	Campaign     *Campaign `json:"campaign,omitempty"`
	EmailID      *string   `gorm:"type:uuid;index"`
	Metadata     string    `gorm:"type:jsonb;default:'{}'"`
}

// ScoreRule represents configurable scoring rules per team
type ScoreRule struct {
	Base
	TeamID       string `gorm:"type:uuid;not null;index"`
	Team         *Team  `json:"team,omitempty"`
	ActivityType string `gorm:"not null;index"` // email_opened, link_clicked, etc.
	Points       int    `gorm:"not null"`
	DecayDays    *int   `json:"decayDays"` // Auto-subtract after X days
	IsActive     bool   `gorm:"not null;default:true;index"`
	Description  string `json:"description"`
}

// TableName for LeadScore
func (LeadScore) TableName() string {
	return "lead_scores"
}

// TableName for ScoreActivity
func (ScoreActivity) TableName() string {
	return "score_activities"
}

// TableName for ScoreRule
func (ScoreRule) TableName() string {
	return "score_rules"
}

// CalculateGrade determines the grade based on score
func CalculateGrade(score int) ScoreGrade {
	switch {
	case score >= 80:
		return GradeA
	case score >= 60:
		return GradeB
	case score >= 40:
		return GradeC
	case score >= 20:
		return GradeD
	default:
		return GradeF
	}
}

// BeforeSave hook to calculate grade before saving
func (ls *LeadScore) BeforeSave(tx *gorm.DB) error {
	ls.Grade = CalculateGrade(ls.Score)
	return nil
}

// DefaultScoreRules returns the default scoring rules for a new team
func DefaultScoreRules(teamID string) []ScoreRule {
	return []ScoreRule{
		{
			TeamID:       teamID,
			ActivityType: "email_opened",
			Points:       5,
			IsActive:     true,
			Description:  "Contact opened an email",
		},
		{
			TeamID:       teamID,
			ActivityType: "email_clicked",
			Points:       10,
			IsActive:     true,
			Description:  "Contact clicked a link in an email",
		},
		{
			TeamID:       teamID,
			ActivityType: "form_submitted",
			Points:       20,
			IsActive:     true,
			Description:  "Contact submitted a form",
		},
		{
			TeamID:       teamID,
			ActivityType: "page_visited",
			Points:       2,
			IsActive:     true,
			Description:  "Contact visited a tracked page",
		},
		{
			TeamID:       teamID,
			ActivityType: "demo_requested",
			Points:       50,
			IsActive:     true,
			Description:  "Contact requested a demo",
		},
		{
			TeamID:       teamID,
			ActivityType: "purchase",
			Points:       100,
			IsActive:     true,
			Description:  "Contact made a purchase",
		},
		{
			TeamID:       teamID,
			ActivityType: "unsubscribed",
			Points:       -50,
			IsActive:     true,
			Description:  "Contact unsubscribed",
		},
		{
			TeamID:       teamID,
			ActivityType: "bounced",
			Points:       -10,
			IsActive:     true,
			Description:  "Email bounced",
		},
		{
			TeamID:       teamID,
			ActivityType: "complained",
			Points:       -25,
			IsActive:     true,
			Description:  "Contact marked email as spam",
		},
	}
}
