package models

import "time"

// Affiliate is a Xem user who refers paying workspaces for a share of what they pay.
type Affiliate struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    string    `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	Code      string    `gorm:"uniqueIndex;not null" json:"code"`
	CreatedAt time.Time `json:"createdAt"`
}

// AffiliateClick is one visit through an affiliate link. Nothing about the visitor is kept.
type AffiliateClick struct {
	ID          uint      `gorm:"primaryKey"`
	AffiliateID string    `gorm:"type:uuid;index;not null"`
	CreatedAt   time.Time `gorm:"index"`
}

// AffiliateReferral ties a workspace to the affiliate whose link it signed up through.
type AffiliateReferral struct {
	TeamID      string `gorm:"type:uuid;primaryKey"`
	AffiliateID string `gorm:"type:uuid;index;not null"`
	CreatedAt   time.Time
}

// AffiliateCommission is the affiliate's share of one plan payment, in the
// settlement currency's smallest unit.
type AffiliateCommission struct {
	ID          string `gorm:"type:uuid;primaryKey"`
	AffiliateID string `gorm:"type:uuid;index;not null"`
	TeamID      string `gorm:"type:uuid;index;not null"`
	PaymentID   string `gorm:"uniqueIndex;not null"`
	Base        int64  `gorm:"not null"` // payment after discounts, without tax
	Amount      int64  `gorm:"not null"` // the affiliate's share
	Currency    string `gorm:"not null"`
	PaidAt      time.Time
	UnlocksAt   time.Time `gorm:"index"`
	ReversedAt  *time.Time
	PayoutID    *string `gorm:"type:uuid;index"`
	// Set when a commission already paid out is reversed and taken back from a later payout.
	ClawbackPayoutID *string `gorm:"type:uuid"`
	CreatedAt        time.Time
}

// AffiliatePayout records a bank transfer to an affiliate.
type AffiliatePayout struct {
	ID          string `gorm:"type:uuid;primaryKey"`
	AffiliateID string `gorm:"type:uuid;index;not null"`
	Amount      int64  `gorm:"not null"`
	Currency    string `gorm:"not null"`
	CreatedAt   time.Time
}
