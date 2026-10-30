package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

// MailConnection credentials are never exposed through generic model CRUD.
// Keep disconnected rows as tombstones so a linked config cannot fall back to
// password authentication after its OAuth credentials have been removed.
type MailConnection struct {
	Base
	TeamID       string  `gorm:"type:uuid;not null;index" json:"-"`
	Provider     string  `gorm:"not null" json:"provider"`
	Address      string  `gorm:"not null" json:"address"`
	AccountID    string  `json:"accountId,omitempty"`
	SMTPConfigID string  `gorm:"type:uuid;not null;uniqueIndex" json:"smtpConfigId"`
	IMAPConfigID *string `gorm:"type:uuid;uniqueIndex" json:"imapConfigId,omitempty"`
	Active       bool    `json:"active"`
	Secret       string  `gorm:"type:text" json:"-"`
}

// A state is single-use, short-lived, and bound to the authenticated user/team.
type MailOAuthState struct {
	Hash      string    `gorm:"primaryKey" json:"-"`
	TeamID    string    `gorm:"type:uuid;index" json:"-"`
	UserID    string    `gorm:"type:uuid" json:"-"`
	Verifier  string    `json:"-"`
	ExpiresAt time.Time `gorm:"index" json:"-"`
}

// Guard the database enqueue boundary too: campaigns have several producers.
func rejectCloudflareMarketing(tx *gorm.DB, team, sender string, campaign bool, category string) error {
	if sender == "" {
		return nil
	}
	var config struct{ Provider string }
	if err := tx.Table("smtp_configs").Select("provider").Where("id = ? AND team_id = ?", sender, team).Scan(&config).Error; err != nil {
		return err
	}
	if config.Provider != "CLOUDFLARE" {
		return nil
	}
	if campaign {
		return errors.New("Cloudflare Email Sending does not support campaigns or newsletters")
	}
	var kind struct{ Name string }
	if err := tx.Table("email_categories").Select("name").Where("id = ? AND team_id = ?", category, team).Scan(&kind).Error; err != nil {
		return err
	}
	if kind.Name != "Transactional" {
		return errors.New("Cloudflare Email Sending only supports transactional email")
	}
	return nil
}

func (e *Email) BeforeCreate(tx *gorm.DB) error {
	if err := e.Base.BeforeCreate(tx); err != nil {
		return err
	}
	if err := ValidateMailAttachments(e.Attachments); err != nil {
		return err
	}
	return rejectCloudflareMarketing(tx, e.TeamID, e.SMTPConfigID, e.CampaignID != "" || e.UnsubscribeURL != "", e.CategoryID)
}

func (c *Campaign) BeforeCreate(tx *gorm.DB) error {
	if err := c.Base.BeforeCreate(tx); err != nil {
		return err
	}
	if err := c.checkArchiveNewsletter(tx, true); err != nil {
		return err
	}
	return rejectCloudflareMarketing(tx, c.TeamID, c.SMTPConfigID, true, "")
}

// BeforeUpdate keeps the archive in the campaign's workspace. A newsletter
// deleted later still passes, so status updates on old campaigns don't fail.
func (c *Campaign) BeforeUpdate(tx *gorm.DB) error {
	return c.checkArchiveNewsletter(tx, false)
}

// checkArchiveNewsletter keeps a campaign's archive inside its own workspace.
func (c *Campaign) checkArchiveNewsletter(tx *gorm.DB, mustExist bool) error {
	if c.ArchiveNewsletterID == nil || *c.ArchiveNewsletterID == "" {
		c.ArchiveNewsletterID = nil
		return nil
	}
	var n int64
	q := tx.Session(&gorm.Session{NewDB: true}).Model(&Newsletter{}).Where("id = ? AND team_id = ?", *c.ArchiveNewsletterID, c.TeamID)
	if mustExist {
		q = q.Where("is_deleted = ?", false)
	}
	if err := q.Count(&n).Error; err != nil {
		return err
	}
	if n == 0 {
		return errors.New("archive newsletter not found")
	}
	return nil
}
