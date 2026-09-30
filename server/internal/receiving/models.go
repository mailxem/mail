package receiving

import "time"

type DomainState struct {
	DomainID       string     `gorm:"primaryKey;type:uuid" json:"domainId"`
	TeamID         string     `gorm:"index;type:uuid;not null" json:"-"`
	Status         string     `gorm:"index;not null" json:"status"`
	Detail         string     `json:"detail,omitempty"`
	ExistingMXJSON []byte     `gorm:"type:jsonb" json:"-"`
	CheckedAt      *time.Time `json:"checkedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func (DomainState) TableName() string { return "managed_receiving_domains" }

type Mailbox struct {
	ID           string    `gorm:"primaryKey;type:uuid" json:"id"`
	TeamID       string    `gorm:"index;type:uuid;not null" json:"-"`
	DomainID     string    `gorm:"index;type:uuid;not null" json:"domainId"`
	Address      string    `gorm:"uniqueIndex;not null" json:"address"`
	DisplayName  string    `json:"displayName"`
	Active       bool      `gorm:"index;not null" json:"active"`
	SMTPConfigID string    `gorm:"uniqueIndex;type:uuid" json:"smtpConfigId,omitempty"`
	Status       string    `gorm:"index;not null" json:"status"`
	UIDValidity  uint32    `gorm:"not null" json:"-"`
	UsageBytes   int64     `gorm:"not null;default:0" json:"usageBytes"`
	QuotaBytes   int64     `gorm:"not null" json:"quotaBytes"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"-"`
}

func (Mailbox) TableName() string { return "managed_receiving_mailboxes" }

type Blob struct {
	ProviderID string    `gorm:"primaryKey"`
	S3Key      string    `gorm:"uniqueIndex;not null"`
	Size       int64     `gorm:"not null"`
	References int64     `gorm:"not null"`
	CreatedAt  time.Time `gorm:"index"`
}

func (Blob) TableName() string { return "managed_receiving_blobs" }

type Message struct {
	ID              string `gorm:"primaryKey;type:uuid"`
	MailboxID       string `gorm:"uniqueIndex:managed_mailbox_provider;uniqueIndex:managed_mailbox_uid;index;type:uuid;not null"`
	TeamID          string `gorm:"index;type:uuid;not null"`
	ProviderID      string `gorm:"uniqueIndex:managed_mailbox_provider;index;not null"`
	UID             uint32 `gorm:"uniqueIndex:managed_mailbox_uid;not null"`
	UIDValidity     uint32 `gorm:"not null"`
	Folder          string `gorm:"index;not null"`
	FromAddress     string
	ToAddress       string
	CcAddress       string
	BccAddress      string
	ReplyTo         string
	Subject         string
	RFCMessageID    string
	SentAt          time.Time
	Snippet         string
	RawKey          string    `gorm:"not null"`
	RawSize         int64     `gorm:"not null"`
	Seen            bool      `gorm:"index;not null"`
	Starred         bool      `gorm:"index;not null"`
	AttachmentsJSON []byte    `gorm:"type:jsonb"`
	CreatedAt       time.Time `gorm:"index"`
	UpdatedAt       time.Time
}

func (Message) TableName() string { return "managed_receiving_messages" }

type Rejection struct {
	ID               string `gorm:"primaryKey;type:uuid"`
	ProviderID       string `gorm:"uniqueIndex:managed_rejection"`
	S3Key            string
	Reason           string    `gorm:"uniqueIndex:managed_rejection"`
	RecipientSetHash string    `gorm:"uniqueIndex:managed_rejection"`
	RecipientsJSON   []byte    `gorm:"type:jsonb"`
	CreatedAt        time.Time `gorm:"index"`
}

func (Rejection) TableName() string { return "managed_receiving_rejections" }

func Migrate(db interface{ AutoMigrate(...interface{}) error }) error {
	return db.AutoMigrate(&DomainState{}, &Mailbox{}, &Blob{}, &Message{}, &Rejection{})
}
