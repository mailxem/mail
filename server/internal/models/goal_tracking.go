package models

import (
	"time"

	"gorm.io/datatypes"
)

// GoalType represents the type of goal
type GoalType string

const (
	GoalTypeEmailOpen     GoalType = "EMAIL_OPEN"
	GoalTypeEmailClick    GoalType = "EMAIL_CLICK"
	GoalTypeFormSubmit    GoalType = "FORM_SUBMIT"
	GoalTypePageVisit     GoalType = "PAGE_VISIT"
	GoalTypePurchase      GoalType = "PURCHASE"
	GoalTypeCustomEvent   GoalType = "CUSTOM_EVENT"
	GoalTypeRevenueTarget GoalType = "REVENUE_TARGET"
)

// AutomationGoal represents a goal for an automation
type AutomationGoal struct {
	Base
	AutomationID string         `gorm:"type:uuid;not null;index"`
	Automation   *Automation    `json:"automation,omitempty"`
	Name         string         `gorm:"not null"`
	Description  string         `json:"description"`
	GoalType     GoalType       `gorm:"not null"`
	TargetValue  *float64       `json:"targetValue"`           // For revenue/conversion rate targets
	IsActive     bool           `gorm:"not null;default:true"` // Can be disabled without deletion
	Metadata     datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
	Conversions  []GoalConversion `json:"conversions,omitempty" gorm:"foreignKey:GoalID"`
}

// GoalConversion represents when a contact achieves a goal
type GoalConversion struct {
	Base
	GoalID            string         `gorm:"type:uuid;not null;index"`
	Goal              *AutomationGoal `json:"goal,omitempty"`
	AutomationID      string         `gorm:"type:uuid;not null;index"`
	Automation        *Automation    `json:"automation,omitempty"`
	ContactID         string         `gorm:"type:uuid;not null;index"`
	Contact           *Contact       `json:"contact,omitempty"`
	ExecutionID       *string        `gorm:"type:uuid;index"`
	Execution         *AutomationExecution `json:"execution,omitempty"`
	ConversionValue   float64        `gorm:"type:decimal(10,2);default:0"` // Revenue or custom value
	Metadata          datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
	ConvertedAt       time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// GoalStats represents aggregated statistics for a goal
type GoalStats struct {
	GoalID            string    `json:"goalId"`
	GoalName          string    `json:"goalName"`
	TotalConversions  int64     `json:"totalConversions"`
	UniqueContacts    int64     `json:"uniqueContacts"`
	TotalValue        float64   `json:"totalValue"`
	ConversionRate    float64   `json:"conversionRate"`    // Conversions / Total executions
	AverageTimeToGoal float64   `json:"averageTimeToGoal"` // Hours from start to conversion
	FirstConversion   *time.Time `json:"firstConversion"`
	LastConversion    *time.Time `json:"lastConversion"`
}

// AutomationGoalSummary represents a summary of all goals for an automation
type AutomationGoalSummary struct {
	AutomationID      string    `json:"automationId"`
	TotalExecutions   int64     `json:"totalExecutions"`
	TotalConversions  int64     `json:"totalConversions"`
	TotalRevenue      float64   `json:"totalRevenue"`
	ConversionRate    float64   `json:"conversionRate"`
	Goals             []GoalStats `json:"goals"`
}

// TableName for AutomationGoal
func (AutomationGoal) TableName() string {
	return "automation_goals"
}

// TableName for GoalConversion
func (GoalConversion) TableName() string {
	return "goal_conversions"
}

// IsRevenueGoal checks if the goal tracks revenue
func (g *AutomationGoal) IsRevenueGoal() bool {
	return g.GoalType == GoalTypeRevenueTarget || g.GoalType == GoalTypePurchase
}

// HasTarget checks if the goal has a target value set
func (g *AutomationGoal) HasTarget() bool {
	return g.TargetValue != nil && *g.TargetValue > 0
}

// DefaultGoals returns common goal templates
func DefaultGoals() []struct {
	Name        string
	Description string
	GoalType    GoalType
} {
	return []struct {
		Name        string
		Description string
		GoalType    GoalType
	}{
		{
			Name:        "Email Engagement",
			Description: "Contact opened an email",
			GoalType:    GoalTypeEmailOpen,
		},
		{
			Name:        "Link Click",
			Description: "Contact clicked a link in an email",
			GoalType:    GoalTypeEmailClick,
		},
		{
			Name:        "Form Submission",
			Description: "Contact submitted a form",
			GoalType:    GoalTypeFormSubmit,
		},
		{
			Name:        "Demo Request",
			Description: "Contact requested a demo",
			GoalType:    GoalTypeCustomEvent,
		},
		{
			Name:        "Purchase",
			Description: "Contact made a purchase",
			GoalType:    GoalTypePurchase,
		},
	}
}
