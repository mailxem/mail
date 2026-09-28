// Package sending implements opt-in managed delivery. Its tables are intentionally
// not registered with generic CRUD: tenant policy and secrets require explicit DTOs.
package sending

import (
	"gorm.io/gorm"
	"time"
)

type Account struct {
	SendingMode         string    `json:"sendingMode"`
	OnboardingDismissed bool      `json:"onboardingDismissed"`
	TeamID              string    `gorm:"primaryKey;type:uuid" json:"teamId"`
	Approved            bool      `json:"approved"`
	Suspended           bool      `json:"suspended"`
	Paused              bool      `json:"paused"`
	DailyLimit          int64     `json:"dailyLimit"`
	MonthlyLimit        int64     `json:"monthlyLimit"`
	MonthlyBudgetMicros int64     `json:"monthlyBudgetMicros"`
	Day                 string    `json:"-"`
	Month               string    `json:"-"`
	DailyUsed           int64     `json:"dailyUsed"`
	MonthlyUsed         int64     `json:"monthlyUsed"`
	BudgetUsedMicros    int64     `json:"budgetUsedMicros"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

func (Account) TableName() string { return "managed_accounts" }

type Domain struct {
	CheckID        string     `json:"-"`
	ID             string     `gorm:"primaryKey;type:uuid" json:"id"`
	TeamID         string     `gorm:"index;type:uuid;not null" json:"-"`
	Name           string     `gorm:"uniqueIndex;not null" json:"name"`
	Token          string     `json:"token"`
	Ownership      bool       `json:"ownership"`
	Provisioned    bool       `json:"provisioned"`
	IdentityStatus string     `json:"identityStatus"`
	DKIMStatus     string     `json:"dkimStatus"`
	MAILFROMStatus string     `gorm:"column:mailfrom_status" json:"mailFromStatus"`
	DMARCStatus    string     `json:"dmarcStatus"`
	DKIMTokens     string     `json:"-"`
	Ready          bool       `json:"ready"`
	CheckedAt      *time.Time `json:"checkedAt"`
	SMTPConfigID   string     `json:"smtpConfigId"`
	FromEmail      string     `json:"fromEmail"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func (Domain) TableName() string { return "managed_domains" }

type Credential struct {
	ID        string     `gorm:"primaryKey;type:uuid" json:"id"`
	TeamID    string     `gorm:"index;type:uuid;not null" json:"-"`
	DomainID  string     `gorm:"index;type:uuid;not null" json:"domainId"`
	Name      string     `json:"name"`
	Hash      string     `json:"-"`
	ExpiresAt time.Time  `json:"expiresAt"`
	RevokedAt *time.Time `json:"revokedAt"`
	CreatedAt time.Time  `json:"createdAt"`
}

func (Credential) TableName() string { return "managed_credentials" }

type Message struct {
	IsTest         bool      `json:"isTest"`
	ID             string    `gorm:"primaryKey;type:uuid" json:"id"`
	TeamID         string    `gorm:"index;type:uuid;not null" json:"-"`
	DomainID       string    `gorm:"type:uuid;not null" json:"domainId"`
	CredentialID   string    `json:"-"`
	EmailID        string    `gorm:"index" json:"emailId,omitempty"`
	IdempotencyKey string    `gorm:"uniqueIndex:managed_idempotency" json:"-"`
	PayloadHash    string    `json:"-"`
	From           string    `json:"from"`
	Recipients     string    `json:"recipients"`
	Subject        string    `json:"subject"`
	Raw            []byte    `json:"-"`
	Status         string    `gorm:"index" json:"status"`
	Detail         string    `json:"detail"`
	ProviderID     string    `gorm:"index" json:"providerId,omitempty"`
	CreatedAt      time.Time `gorm:"index" json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (Message) TableName() string { return "managed_messages" }

type Event struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	TeamID    string    `gorm:"index;type:uuid" json:"-"`
	MessageID string    `gorm:"index;type:uuid" json:"messageId"`
	Kind      string    `json:"kind"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"createdAt"`
}

func (Event) TableName() string { return "managed_events" }

type Suppression struct {
	TeamID    string    `gorm:"primaryKey;type:uuid" json:"-"`
	Address   string    `gorm:"primaryKey" json:"address"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"createdAt"`
}

func (Suppression) TableName() string { return "managed_suppressions" }

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&Account{}, &Domain{}, &Credential{}, &Message{}, &Event{}, &Suppression{}, &Audit{}); err != nil {
		return err
	}
	return upgradeStarterAllowance(db)
}

type Audit struct {
	ID        string    `gorm:"primaryKey;type:uuid" json:"id"`
	TeamID    string    `gorm:"index;type:uuid" json:"-"`
	Actor     string    `json:"actor"`
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"createdAt"`
}

func (Audit) TableName() string { return "managed_audits" }
