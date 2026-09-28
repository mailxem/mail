package models

import (
	"time"

	"gorm.io/datatypes"
)

// ContactEngagementPattern tracks when a contact opens emails
type ContactEngagementPattern struct {
	Base
	ContactID        string         `gorm:"type:uuid;not null;uniqueIndex"`
	Contact          *Contact       `json:"contact,omitempty"`
	Timezone         string         `gorm:"default:'UTC'"` // IANA timezone (America/New_York)
	TimezoneDetected bool           `gorm:"default:false"`

	// Hourly engagement data (0-23 hours)
	HourlyEngagement datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"hourlyEngagement"` // {0: 5, 1: 2, 14: 10}

	// Day of week engagement (0=Sunday, 6=Saturday)
	DailyEngagement  datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"dailyEngagement"` // {0: 3, 1: 15, 2: 12}

	// Combined day+hour heatmap
	HeatmapData      datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"heatmapData"` // {"monday_14": 5, "tuesday_9": 8}

	// Calculated optimal times
	OptimalHour      *int           `json:"optimalHour"`      // 0-23
	OptimalDayOfWeek *int           `json:"optimalDayOfWeek"` // 0-6
	OptimalScore     float64        `gorm:"default:0"`        // Confidence score 0-100

	// Statistics
	TotalOpens       int            `gorm:"default:0"`
	LastOpenAt       *time.Time     `json:"lastOpenAt"`
	LastCalculatedAt *time.Time     `json:"lastCalculatedAt"`
	DataQuality      float64        `gorm:"default:0"` // 0-100, based on sample size
}

// EmailOpenEvent represents individual open events for learning
type EmailOpenEvent struct {
	Base
	ContactID        string         `gorm:"type:uuid;not null;index"`
	Contact          *Contact       `json:"contact,omitempty"`
	EmailID          string         `gorm:"type:uuid;not null;index"`
	Email            *Email         `json:"email,omitempty"`
	OpenedAt         time.Time      `gorm:"not null;index"`
	OpenedHour       int            `gorm:"not null;index"` // 0-23
	OpenedDayOfWeek  int            `gorm:"not null;index"` // 0-6
	OpenedTimezone   string         `json:"openedTimezone"`
	IPAddress        *string        `json:"ipAddress"`
	UserAgent        *string        `json:"userAgent"`
	IsFirstOpen      bool           `gorm:"default:false"` // First open vs subsequent opens
	TimeToOpen       *int           `json:"timeToOpen"`    // Minutes from send to open
}

// TeamSendTimeDefaults represents team-level send time preferences
type TeamSendTimeDefaults struct {
	Base
	TeamID           string         `gorm:"type:uuid;not null;uniqueIndex"`
	Team             *Team          `json:"team,omitempty"`

	// Default sending windows
	DefaultTimezone  string         `gorm:"default:'UTC'"`
	DefaultSendHour  int            `gorm:"default:10"` // 10 AM
	EarliestSendHour int            `gorm:"default:8"`  // 8 AM
	LatestSendHour   int            `gorm:"default:20"` // 8 PM

	// Send day preferences
	SendMonday       bool           `gorm:"default:true"`
	SendTuesday      bool           `gorm:"default:true"`
	SendWednesday    bool           `gorm:"default:true"`
	SendThursday     bool           `gorm:"default:true"`
	SendFriday       bool           `gorm:"default:true"`
	SendSaturday     bool           `gorm:"default:false"`
	SendSunday       bool           `gorm:"default:false"`

	// Optimization settings
	UseSTOByDefault  bool           `gorm:"default:false"` // Enable STO by default for all campaigns
	MinimumOpensRequired int        `gorm:"default:5"`     // Minimum opens before using STO
	FallbackStrategy string         `gorm:"default:'team_average'"` // "team_average", "industry_average", "fixed_time"

	// Team-level aggregate data
	TeamAverageOptimalHour *int   `json:"teamAverageOptimalHour"`
	TeamAverageOptimalDay  *int   `json:"teamAverageOptimalDay"`
}

// SendTimeQueue represents emails queued for optimal send time
type SendTimeQueue struct {
	Base
	TeamID           string         `gorm:"type:uuid;not null;index"`
	Team             *Team          `json:"team,omitempty"`
	EmailID          string         `gorm:"type:uuid;not null;index"`
	Email            *Email         `json:"email,omitempty"`
	ContactID        string         `gorm:"type:uuid;not null;index"`
	Contact          *Contact       `json:"contact,omitempty"`
	AutomationID     *string        `gorm:"type:uuid;index"`
	Automation       *Automation    `json:"automation,omitempty"`

	// Original schedule
	RequestedSendAt  time.Time      `gorm:"not null"`

	// Optimized schedule
	OptimizedSendAt  time.Time      `gorm:"not null;index"`
	OptimizationUsed string         `json:"optimizationUsed"` // "contact_pattern", "segment_average", "team_default"

	// Status
	Status           string         `gorm:"not null;default:'QUEUED';index"` // "QUEUED", "SENT", "FAILED", "CANCELLED"
	SentAt           *time.Time     `json:"sentAt"`
	ProcessedAt      *time.Time     `json:"processedAt"`
	Error            *string        `gorm:"type:text" json:"error"`

	// Metadata
	Metadata         datatypes.JSON `gorm:"type:jsonb;default:'{}'"` // Additional context
}

// TableName for ContactEngagementPattern
func (ContactEngagementPattern) TableName() string {
	return "contact_engagement_patterns"
}

// TableName for EmailOpenEvent
func (EmailOpenEvent) TableName() string {
	return "email_open_events"
}

// TableName for TeamSendTimeDefaults
func (TeamSendTimeDefaults) TableName() string {
	return "team_send_time_defaults"
}

// TableName for SendTimeQueue
func (SendTimeQueue) TableName() string {
	return "send_time_queue"
}

// HasSufficientData checks if the contact has enough data for STO
func (c *ContactEngagementPattern) HasSufficientData(minOpens int) bool {
	return c.TotalOpens >= minOpens && c.DataQuality >= 50
}

// GetOptimalSendTime returns the next optimal send time after the given time
func (c *ContactEngagementPattern) GetOptimalSendTime(after time.Time) time.Time {
	if c.OptimalHour == nil || c.OptimalDayOfWeek == nil {
		return after
	}

	// Parse timezone
	loc, err := time.LoadLocation(c.Timezone)
	if err != nil {
		loc = time.UTC
	}

	// Convert to contact's timezone
	afterInTZ := after.In(loc)

	// Find next occurrence of optimal day/hour
	targetDay := *c.OptimalDayOfWeek
	targetHour := *c.OptimalHour

	// Calculate days until target day
	currentDay := int(afterInTZ.Weekday())
	daysUntil := (targetDay - currentDay + 7) % 7

	if daysUntil == 0 && afterInTZ.Hour() >= targetHour {
		// Target day is today but hour has passed, move to next week
		daysUntil = 7
	}

	// Calculate optimal send time
	optimalTime := afterInTZ.AddDate(0, 0, daysUntil)
	optimalTime = time.Date(
		optimalTime.Year(),
		optimalTime.Month(),
		optimalTime.Day(),
		targetHour,
		0, 0, 0,
		loc,
	)

	return optimalTime
}

// IncrementEngagement increments engagement count for a specific hour/day
func (c *ContactEngagementPattern) IncrementEngagement(openedAt time.Time) {
	// This would be implemented in the service layer with JSON manipulation
	c.TotalOpens++
	c.LastOpenAt = &openedAt
}

// CalculateDataQuality calculates the quality score based on sample size
func (c *ContactEngagementPattern) CalculateDataQuality() float64 {
	// Quality score based on sample size
	// 5 opens = 25%, 10 opens = 50%, 20 opens = 75%, 50+ opens = 100%
	switch {
	case c.TotalOpens >= 50:
		return 100
	case c.TotalOpens >= 20:
		return 75
	case c.TotalOpens >= 10:
		return 50
	case c.TotalOpens >= 5:
		return 25
	default:
		return float64(c.TotalOpens) * 5 // 0-25% for 0-5 opens
	}
}

// IsWithinSendWindow checks if a time is within allowed send window
func (t *TeamSendTimeDefaults) IsWithinSendWindow(checkTime time.Time) bool {
	hour := checkTime.Hour()
	dayOfWeek := int(checkTime.Weekday())

	// Check hour window
	if hour < t.EarliestSendHour || hour > t.LatestSendHour {
		return false
	}

	// Check day of week
	switch dayOfWeek {
	case 0:
		return t.SendSunday
	case 1:
		return t.SendMonday
	case 2:
		return t.SendTuesday
	case 3:
		return t.SendWednesday
	case 4:
		return t.SendThursday
	case 5:
		return t.SendFriday
	case 6:
		return t.SendSaturday
	}

	return false
}

// GetNextAllowedSendTime gets the next time within the send window
func (t *TeamSendTimeDefaults) GetNextAllowedSendTime(after time.Time) time.Time {
	loc, _ := time.LoadLocation(t.DefaultTimezone)
	checkTime := after.In(loc)

	// Try next 14 days to find valid send time
	for i := 0; i < 14; i++ {
		// Check if this day is allowed
		if t.IsWithinSendWindow(checkTime) {
			// If hour is before window, set to earliest hour
			if checkTime.Hour() < t.EarliestSendHour {
				return time.Date(
					checkTime.Year(),
					checkTime.Month(),
					checkTime.Day(),
					t.EarliestSendHour,
					0, 0, 0,
					loc,
				)
			}
			return checkTime
		}

		// Move to next day at earliest hour
		checkTime = checkTime.AddDate(0, 0, 1)
		checkTime = time.Date(
			checkTime.Year(),
			checkTime.Month(),
			checkTime.Day(),
			t.EarliestSendHour,
			0, 0, 0,
			loc,
		)
	}

	// Fallback: return original time
	return after
}
