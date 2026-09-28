package models

import (
	"time"

	"gorm.io/datatypes"
)

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusPending  SubscriptionStatus = "pending"
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
	SubscriptionStatusPaused   SubscriptionStatus = "paused"
	SubscriptionStatusFailed   SubscriptionStatus = "failed"
)

// ProductFeature represents a feature that can be enabled/disabled for products
type ProductFeature string

const (
	FeatureEmailCampaigns    ProductFeature = "email_campaigns"
	FeatureTemplateLibrary   ProductFeature = "template_library"
	FeatureAdvancedAnalytics ProductFeature = "advanced_analytics"
	FeatureCustomDomain      ProductFeature = "custom_domain"
	FeatureAPIAccess         ProductFeature = "api_access"
	FeatureTeamCollaboration ProductFeature = "team_collaboration"
	FeatureAutomation        ProductFeature = "automation"
	FeatureSegmentation      ProductFeature = "segmentation"
)

// Product represents a subscription product
type Product struct {
	Base
	Name        string                 `json:"name" gorm:"not null"`
	Description string                 `json:"description"`
	Price       float64                `json:"price" gorm:"not null"`
	Interval    string                 `json:"interval" gorm:"not null"` // monthly, yearly
	DodoID      string                 `json:"dodo_id" gorm:"unique"`    // ID from Dodo Payments
	Features    []ProductFeatureConfig `json:"features" gorm:"foreignKey:ProductID"`
}

// ProductFeatureConfig represents the configuration of a feature for a product
type ProductFeatureConfig struct {
	Base
	ProductID string         `json:"product_id" gorm:"type:uuid;not null"`
	Feature   ProductFeature `json:"feature" gorm:"not null"`
	Enabled   bool           `json:"enabled" gorm:"not null;default:false"`
	Limit     int            `json:"limit"` // Optional limit for the feature (e.g., number of campaigns)
}

// Subscription represents a team's subscription
type Subscription struct {
	Base
	TeamID             string             `json:"team_id" gorm:"type:uuid;not null;unique"`
	Team               *Team              `json:"team,omitempty"`
	ProductID          string             `json:"product_id" gorm:"type:uuid;not null"`
	Product            *Product           `json:"product,omitempty"`
	Status             SubscriptionStatus `json:"status" gorm:"not null;default:'pending'"`
	DodoSubscriptionID string             `json:"dodo_subscription_id" gorm:"unique"`
	DodoCustomerID     string             `json:"dodo_customer_id"`
	CurrentPeriodStart time.Time          `json:"current_period_start"`
	CurrentPeriodEnd   time.Time          `json:"current_period_end"`
	CanceledAt         *time.Time         `json:"canceled_at,omitempty"`
	TrialEnd           *time.Time         `json:"trial_end,omitempty"`
	Email              string             `json:"email" gorm:"not null"` // Email used during purchase
	Metadata           datatypes.JSON     `json:"metadata" gorm:"type:jsonb"`
}

// HasFeature checks if a subscription has a specific feature enabled
func (s *Subscription) HasFeature(feature ProductFeature) bool {
	if s.Product == nil {
		return false
	}

	for _, f := range s.Product.Features {
		if f.Feature == feature && f.Enabled {
			return true
		}
	}
	return false
}

// GetFeatureLimit returns the limit for a specific feature
func (s *Subscription) GetFeatureLimit(feature ProductFeature) int {
	if s.Product == nil {
		return 0
	}

	for _, f := range s.Product.Features {
		if f.Feature == feature {
			return f.Limit
		}
	}
	return 0
}
