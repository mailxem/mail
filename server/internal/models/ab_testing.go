package models

import (
	"time"

	"gorm.io/datatypes"
)

// TestStatus represents the status of an A/B test
type TestStatus string

const (
	TestStatusDraft     TestStatus = "DRAFT"
	TestStatusRunning   TestStatus = "RUNNING"
	TestStatusCompleted TestStatus = "COMPLETED"
	TestStatusStopped   TestStatus = "STOPPED"
)

// TestType represents the type of A/B test
type TestType string

const (
	TestTypeSubject  TestType = "SUBJECT"
	TestTypeContent  TestType = "CONTENT"
	TestTypeSendTime TestType = "SEND_TIME"
	TestTypeFull     TestType = "FULL"
)

// WinnerMetric represents the metric used to determine the winner
type WinnerMetric string

const (
	MetricOpenRate       WinnerMetric = "OPEN_RATE"
	MetricClickRate      WinnerMetric = "CLICK_RATE"
	MetricConversionRate WinnerMetric = "CONVERSION_RATE"
	MetricRevenue        WinnerMetric = "REVENUE"
)

// ABTest represents an A/B test
type ABTest struct {
	Base
	TeamID               string       `gorm:"type:uuid;not null;index"`
	Team                 *Team        `json:"team,omitempty"`
	AutomationID         *string      `gorm:"type:uuid;index"`
	Automation           *Automation  `json:"automation,omitempty"`
	CampaignID           *string      `gorm:"type:uuid;index"`
	Campaign             *Campaign    `json:"campaign,omitempty"`
	Name                 string       `gorm:"not null"`
	Description          string       `json:"description"`
	TestType             TestType     `gorm:"not null"`
	WinnerMetric         WinnerMetric `gorm:"not null"`
	Status               TestStatus   `gorm:"not null;default:'DRAFT';index"`
	WinnerVariantID      *string      `gorm:"type:uuid"`
	WinnerVariant        *TestVariant `json:"winnerVariant,omitempty" gorm:"foreignKey:WinnerVariantID"`
	ConfidenceThreshold  float64      `gorm:"not null;default:0.95"` // 95% confidence
	MinimumSampleSize    int          `gorm:"not null;default:100"`  // Minimum sends per variant
	AutoDeclareWinner    bool         `gorm:"default:true"`          // Automatically declare winner when threshold met
	StartedAt            *time.Time   `json:"startedAt"`
	EndedAt              *time.Time   `json:"endedAt"`
	Variants             []TestVariant `json:"variants,omitempty" gorm:"foreignKey:TestID"`
}

// TestVariant represents a variant in an A/B test
type TestVariant struct {
	Base
	TestID          string         `gorm:"type:uuid;not null;index"`
	Test            *ABTest        `json:"test,omitempty"`
	VariantName     string         `gorm:"not null"` // "A", "B", "C"
	SplitPercentage int            `gorm:"not null"` // 33, 33, 34 for 3-way split
	SubjectLine     *string        `json:"subjectLine"`
	EmailContent    *string        `gorm:"type:text" json:"emailContent"`
	SendHour        *int           `json:"sendHour"` // For send-time tests (0-23)
	Metadata        datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
	Result          *TestResult    `json:"result,omitempty" gorm:"foreignKey:VariantID"`
}

// TestResult represents aggregated results for a test variant
type TestResult struct {
	Base
	TestID      string   `gorm:"type:uuid;not null;index"`
	Test        *ABTest  `json:"test,omitempty"`
	VariantID   string   `gorm:"type:uuid;not null;index;unique"`
	Variant     *TestVariant `json:"variant,omitempty"`
	Sends       int      `gorm:"not null;default:0"`
	Opens       int      `gorm:"not null;default:0"`
	Clicks      int      `gorm:"not null;default:0"`
	Conversions int      `gorm:"not null;default:0"`
	Revenue     float64  `gorm:"type:decimal(10,2);default:0"`
	OpenRate    float64  `gorm:"type:decimal(5,4);default:0"` // Calculated field
	ClickRate   float64  `gorm:"type:decimal(5,4);default:0"` // Calculated field
	ConversionRate float64 `gorm:"type:decimal(5,4);default:0"` // Calculated field
}

// VariantAssignment tracks which contact received which variant
type VariantAssignment struct {
	TestID     string    `gorm:"type:uuid;not null;primaryKey"`
	Test       *ABTest   `json:"test,omitempty"`
	ContactID  string    `gorm:"type:uuid;not null;primaryKey"`
	Contact    *Contact  `json:"contact,omitempty"`
	VariantID  string    `gorm:"type:uuid;not null;index"`
	Variant    *TestVariant `json:"variant,omitempty"`
	AssignedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// TableName for ABTest
func (ABTest) TableName() string {
	return "ab_tests"
}

// TableName for TestVariant
func (TestVariant) TableName() string {
	return "test_variants"
}

// TableName for TestResult
func (TestResult) TableName() string {
	return "test_results"
}

// TableName for VariantAssignment
func (VariantAssignment) TableName() string {
	return "variant_assignments"
}

// CalculateRates updates the calculated rate fields
func (r *TestResult) CalculateRates() {
	if r.Sends > 0 {
		r.OpenRate = float64(r.Opens) / float64(r.Sends)
		r.ClickRate = float64(r.Clicks) / float64(r.Sends)
		r.ConversionRate = float64(r.Conversions) / float64(r.Sends)
	}
}

// GetRate returns the rate for the specified metric
func (r *TestResult) GetRate(metric WinnerMetric) float64 {
	switch metric {
	case MetricOpenRate:
		return r.OpenRate
	case MetricClickRate:
		return r.ClickRate
	case MetricConversionRate:
		return r.ConversionRate
	case MetricRevenue:
		return r.Revenue
	default:
		return 0
	}
}

// IsSignificant checks if the test has reached minimum sample size
func (t *ABTest) IsSignificant() bool {
	// Would need to load results to check
	return true // Placeholder
}

// DefaultVariantSplit returns equal split percentages for n variants
func DefaultVariantSplit(n int) []int {
	if n <= 0 {
		return []int{}
	}

	basePercent := 100 / n
	remainder := 100 % n

	splits := make([]int, n)
	for i := 0; i < n; i++ {
		splits[i] = basePercent
		if i < remainder {
			splits[i]++
		}
	}

	return splits
}
