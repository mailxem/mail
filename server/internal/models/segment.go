package models

import (
	"time"

	"gorm.io/datatypes"
)

// Segment represents a group of contacts based on rules
type Segment struct {
	Base
	TeamID       string         `gorm:"type:uuid;not null;index" json:"teamId" validate:"required,uuid"`
	Team         *Team          `json:"team,omitempty"`
	Name         string         `gorm:"not null" json:"name" validate:"required,min=2,max=100"`
	Description  string         `json:"description"`
	Rules        datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"rules"`
	MatchType    MatchType      `gorm:"not null;default:'ALL'" json:"matchType"`
	IsDynamic    bool           `gorm:"default:true" json:"isDynamic"` // Auto-update membership
	ContactCount int            `gorm:"default:0" json:"contactCount"` // Cached count
	LastRefresh  *time.Time     `json:"lastRefresh"`
}

// SegmentContact represents static segment membership
type SegmentContact struct {
	SegmentID string    `gorm:"type:uuid;not null;primaryKey" json:"segmentId"`
	Segment   *Segment  `json:"segment,omitempty"`
	ContactID string    `gorm:"type:uuid;not null;primaryKey" json:"contactId"`
	Contact   *Contact  `json:"contact,omitempty"`
	AddedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"addedAt"`
}

// MatchType defines how rules are combined
type MatchType string

const (
	MatchTypeAll MatchType = "ALL" // AND logic - all rules must match
	MatchTypeAny MatchType = "ANY" // OR logic - any rule can match
)

// SegmentRule represents a single rule in a segment
type SegmentRule struct {
	Field    string      `json:"field"`    // Contact field name (e.g., "email", "tags", "country")
	Operator RuleOperator `json:"operator"` // Comparison operator
	Value    interface{} `json:"value"`    // Value to compare against
}

// RuleOperator defines comparison operators for segment rules
type RuleOperator string

const (
	// String operators
	OpEquals       RuleOperator = "equals"
	OpNotEquals    RuleOperator = "not_equals"
	OpContains     RuleOperator = "contains"
	OpNotContains  RuleOperator = "not_contains"
	OpStartsWith   RuleOperator = "starts_with"
	OpEndsWith     RuleOperator = "ends_with"
	OpIsSet        RuleOperator = "is_set"
	OpIsNotSet     RuleOperator = "is_not_set"

	// Number operators
	OpGreaterThan    RuleOperator = "greater_than"
	OpLessThan       RuleOperator = "less_than"
	OpGreaterOrEqual RuleOperator = "greater_or_equal"
	OpLessOrEqual    RuleOperator = "less_or_equal"

	// Array operators
	OpIncludes    RuleOperator = "includes"
	OpNotIncludes RuleOperator = "not_includes"
	OpIsEmpty     RuleOperator = "is_empty"
	OpIsNotEmpty  RuleOperator = "is_not_empty"

	// Date operators
	OpBefore      RuleOperator = "before"
	OpAfter       RuleOperator = "after"
	OpWithinDays  RuleOperator = "within_days"
	OpOlderThan   RuleOperator = "older_than"

	// Boolean operators
	OpIsTrue  RuleOperator = "is_true"
	OpIsFalse RuleOperator = "is_false"
)

// SegmentRules represents the complete rule set for a segment
type SegmentRules struct {
	Rules     []SegmentRule `json:"rules"`
	MatchType MatchType     `json:"matchType"` // ALL or ANY
}

// TableName specifies the table name for SegmentContact
func (SegmentContact) TableName() string {
	return "segment_contacts"
}
