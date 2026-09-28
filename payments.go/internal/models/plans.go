package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// StringArray is a custom type that properly handles PostgreSQL text[] arrays
type StringArray []string

// Scan implements the sql.Scanner interface for StringArray
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return a.scanBytes(v)
	case string:
		return a.scanBytes([]byte(v))
	default:
		return fmt.Errorf("cannot scan %T into StringArray", value)
	}
}

// scanBytes parses PostgreSQL array format {item1,item2,item3}
func (a *StringArray) scanBytes(src []byte) error {
	str := string(src)
	if str == "{}" {
		*a = StringArray{}
		return nil
	}

	// Remove the braces
	if len(str) < 2 || str[0] != '{' || str[len(str)-1] != '}' {
		return fmt.Errorf("invalid array format: %s", str)
	}
	str = str[1 : len(str)-1]

	// Split by comma and handle quoted strings
	var result []string
	if str != "" {
		// Simple split for now - this handles basic cases
		// For more complex cases with quotes and escapes, would need more sophisticated parsing
		parts := strings.Split(str, ",")
		for _, part := range parts {
			// Remove quotes if present
			part = strings.TrimSpace(part)
			if len(part) >= 2 && part[0] == '"' && part[len(part)-1] == '"' {
				part = part[1 : len(part)-1]
			}
			result = append(result, part)
		}
	}

	*a = StringArray(result)
	return nil
}

// Value implements the driver.Valuer interface for StringArray
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}

	// Quote each element to handle special characters
	quoted := make([]string, len(a))
	for i, s := range a {
		// Escape quotes and backslashes in the string
		escaped := strings.ReplaceAll(s, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		quoted[i] = `"` + escaped + `"`
	}

	return "{" + strings.Join(quoted, ",") + "}", nil
}

// JSONMap is a custom type that properly handles PostgreSQL JSONB fields
type JSONMap map[string]interface{}

// Scan implements the sql.Scanner interface for JSONMap
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = JSONMap{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}

	if len(bytes) == 0 {
		*j = JSONMap{}
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = JSONMap(result)
	return nil
}

// Value implements the driver.Valuer interface for JSONMap
func (j JSONMap) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "{}", nil
	}

	bytes, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}

	return string(bytes), nil
}

type PlanStatus string

const (
	PlanStatusActive   PlanStatus = "active"   // @Description Plan is active and available for new subscriptions
	PlanStatusInactive PlanStatus = "inactive" // @Description Plan is inactive but existing subscriptions continue
	PlanStatusArchived PlanStatus = "archived" // @Description Plan is permanently hidden, no new subscriptions allowed
)

type Plan struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	DodoPlanID  string     `json:"dodo_plan_id" gorm:"unique;not null;index"` // Dodo Payments product ID
	Name        string     `json:"name" gorm:"not null"`
	Description *string    `json:"description,omitempty"`
	Status      PlanStatus `json:"status" gorm:"default:active"`

	// Pricing
	Price    decimal.Decimal `json:"price" gorm:"type:decimal(10,2)"`
	Currency string          `json:"currency" gorm:"default:USD"`

	// Legacy pricing (for backward compatibility)
	PricePerSeat decimal.Decimal `json:"price_per_seat" gorm:"type:decimal(10,2)"` // Will be same as PriceMonthly
	BillingCycle BillingCycle    `json:"billing_cycle" gorm:"default:MONTHLY"`

	// Trial settings
	TrialPeriodDays int  `json:"trial_period_days" gorm:"default:7"`
	TrialEnabled    bool `json:"trial_enabled" gorm:"default:true"`

	// Display & Ordering
	DisplayOrder  int            `json:"display_order" gorm:"default:0"`
	IsPopular     bool           `json:"is_popular" gorm:"default:false"`
	IsEnterprise  bool           `json:"is_enterprise" gorm:"default:false"`
	Features      []PlanFeatures `json:"features,omitempty" gorm:"-"`
	TrailFeatures []PlanFeatures `json:"trail_features,omitempty" gorm:"-"`

	// Metadata
	Metadata  JSONMap   `json:"metadata" gorm:"type:jsonb"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Subscriptions []Subscription `json:"subscriptions,omitempty" gorm:"foreignKey:PlanID;references:ID"`
}

type PlanFeature string

type PlanFeatures struct {
	ID        string      `json:"id" gorm:"primaryKey"`
	PlanID    string      `json:"plan_id" gorm:"not null;index"`
	Plan      Plan        `json:"plan" gorm:"foreignKey:PlanID;references:ID"`
	Quota     int         `json:"quota" gorm:"not null"`
	Feature   PlanFeature `json:"feature" gorm:"not null"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Trail     bool        `json:"trail" gorm:"default:false"`
}

func (p *PlanFeatures) ToResponse() *PlanFeaturesResponse {
	return &PlanFeaturesResponse{
		ID:      p.ID,
		Quota:   p.Quota,
		Feature: p.Feature,
	}
}

type PlanFeaturesResponse struct {
	ID      string      `json:"id" gorm:"primaryKey"`
	PlanID  string      `json:"plan_id" gorm:"not null;index"`
	Quota   int         `json:"quota" gorm:"not null"`
	Feature PlanFeature `json:"feature" gorm:"not null"`
}

func (p *PlanFeaturesResponse) ToPlanFeature() PlanFeatures {
	return PlanFeatures{
		ID:      p.ID,
		PlanID:  p.PlanID,
		Quota:   p.Quota,
		Feature: p.Feature,
	}
}

const (
	PlanFeatureQueries            = "queries"
	PlanFeatureChats              = "chats"
	PlanFeatureDataSources        = "data_sources"
	PlanFeatureAPIAccess          = "api_access"
	PlanFeatureCustomIntegrations = "custom_integrations"
	PlanFeatureWhiteLabel         = "white_label"
)

// UnlimitedQuota represents unlimited usage for a feature
// Use this instead of magic numbers to check for unlimited quotas
const UnlimitedQuota = -1

// IsUnlimitedQuota checks if a quota value represents unlimited usage
// Supports both new (-1) and legacy (99999999) unlimited values
func IsUnlimitedQuota(quota int) bool {
	return quota == UnlimitedQuota || quota <= 0 || quota == 99999999
}

// CreatePlanRequest represents the request to create a new plan
type CreatePlanRequest struct {
	DodoPlanID      string          `json:"dodo_plan_id" validate:"required" example:"pro-plan-2024" binding:"required"`
	Name            string          `json:"name" validate:"required" example:"Professional" binding:"required"`
	Description     *string         `json:"description,omitempty" example:"For growing teams and businesses"`
	Price           decimal.Decimal `json:"price" validate:"required" example:"49.00" binding:"required"`
	Currency        string          `json:"currency" validate:"required" example:"USD" binding:"required"`
	TrialPeriodDays int             `json:"trial_period_days" example:"14"`
	TrialEnabled    bool            `json:"trial_enabled" example:"true"`
	DisplayOrder    int             `json:"display_order" example:"2"`
	IsPopular       bool            `json:"is_popular" example:"true"`
	IsEnterprise    bool            `json:"is_enterprise" example:"false"`
	Features        StringArray     `json:"features"`
	Metadata        JSONMap         `json:"metadata,omitempty"`
}

// UpdatePlanRequest represents the request to update an existing plan
type UpdatePlanRequest struct {
	Name            *string          `json:"name,omitempty" example:"Professional Plus"`
	Description     *string          `json:"description,omitempty" example:"Enhanced professional plan"`
	Status          *PlanStatus      `json:"status,omitempty" example:"active"`
	Price           *decimal.Decimal `json:"price,omitempty" example:"59.00"`
	Currency        *string          `json:"currency,omitempty" example:"USD"`
	TrialPeriodDays *int             `json:"trial_period_days,omitempty" example:"21"`
	TrialEnabled    *bool            `json:"trial_enabled,omitempty" example:"true"`
	DisplayOrder    *int             `json:"display_order,omitempty" example:"3"`
	IsPopular       *bool            `json:"is_popular,omitempty" example:"false"`
	IsEnterprise    *bool            `json:"is_enterprise,omitempty" example:"false"`
	Features        StringArray      `json:"features,omitempty"`
	Metadata        JSONMap          `json:"metadata,omitempty"`
}

// PlanResponse represents the response structure for plan data
type PlanResponse struct {
	ID          string     `json:"id" example:"prof-123e4567-e89b-12d3-a456-426614174000"`
	DodoPlanID  string     `json:"dodo_plan_id" example:"pro-plan-2024"`
	Name        string     `json:"name" example:"Professional"`
	Description *string    `json:"description,omitempty" example:"For growing teams and businesses"`
	Status      PlanStatus `json:"status" example:"active"`

	// Pricing
	Price    decimal.Decimal `json:"price" example:"49.00"`
	Currency string          `json:"currency" example:"USD"`

	// Legacy for backward compatibility
	PricePerSeat decimal.Decimal `json:"price_per_seat" example:"49.00"`
	BillingCycle BillingCycle    `json:"billing_cycle" example:"MONTHLY"`

	// Features & Limits
	Features      []PlanFeaturesResponse `json:"features"`
	TrailFeatures []PlanFeaturesResponse `json:"trail_features"`

	// Trial settings
	TrialPeriodDays int  `json:"trial_period_days" example:"14"`
	TrialEnabled    bool `json:"trial_enabled" example:"true"`

	// Display & Ordering
	DisplayOrder int  `json:"display_order" example:"2"`
	IsPopular    bool `json:"is_popular" example:"true"`
	IsEnterprise bool `json:"is_enterprise" example:"false"`

	// Metadata
	Metadata  JSONMap   `json:"metadata,omitempty"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:00:00Z"`

	// Statistics (optional)
	ActiveSubscriptions int `json:"active_subscriptions,omitempty" example:"25"`
}

// BeforeCreate hook to generate UUID
func (p *Plan) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	// Set legacy fields for backward compatibility
	p.PricePerSeat = p.Price
	return nil
}

// BeforeUpdate hook to maintain backward compatibility
func (p *Plan) BeforeUpdate(tx *gorm.DB) error {
	p.PricePerSeat = p.Price
	return nil
}

// ToResponse converts Plan model to PlanResponse
func (p *Plan) ToResponse() *PlanResponse {
	// Convert []PlanFeatures to []PlanFeaturesResponse
	var featuresResponse []PlanFeaturesResponse
	for _, feature := range p.Features {
		featuresResponse = append(featuresResponse, *feature.ToResponse())
	}

	var trailFeaturesResponse []PlanFeaturesResponse
	for _, feature := range p.TrailFeatures {
		trailFeaturesResponse = append(trailFeaturesResponse, *feature.ToResponse())
	}

	return &PlanResponse{
		ID:              p.ID,
		DodoPlanID:      p.DodoPlanID,
		Name:            p.Name,
		Description:     p.Description,
		Status:          p.Status,
		Price:           p.Price,
		Currency:        p.Currency,
		PricePerSeat:    p.PricePerSeat,
		BillingCycle:    p.BillingCycle,
		TrialPeriodDays: p.TrialPeriodDays,
		TrialEnabled:    p.TrialEnabled,
		DisplayOrder:    p.DisplayOrder,
		IsPopular:       p.IsPopular,
		IsEnterprise:    p.IsEnterprise,
		Features:        featuresResponse,
		TrailFeatures:   trailFeaturesResponse,
		Metadata:        p.Metadata,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

// GetPriceForBillingCycle returns the price for a specific billing cycle
func (p *Plan) GetPriceForBillingCycle(cycle BillingCycle) decimal.Decimal {
	switch cycle {
	case BillingYearly:
		return p.Price
	default:
		return p.Price
	}
}

// IsActive returns true if the plan is currently active
func (p *Plan) IsActive() bool {
	return p.Status == PlanStatusActive
}

// CanHaveTrial returns true if the plan supports trials
func (p *Plan) CanHaveTrial() bool {
	return p.TrialEnabled && p.TrialPeriodDays > 0
}
