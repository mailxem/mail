package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// Workspace tracking consent modes.
const (
	TrackingConsentNone   = "none"    // follow each email's switch
	TrackingConsentOptOut = "opt_out" // track unless the reader objected
	TrackingConsentOptIn  = "opt_in"  // track only readers who agreed
)

// TrackingConsent is a reader's latest choice about open and click tracking in one workspace.
type TrackingConsent struct {
	Base
	TeamID    string    `gorm:"type:uuid;not null;uniqueIndex:idx_tracking_consent_address" json:"teamId"`
	Email     string    `gorm:"not null;uniqueIndex:idx_tracking_consent_address" json:"email"`
	Allowed   bool      `gorm:"not null" json:"allowed"`
	Source    string    `gorm:"not null" json:"source"`
	ChangedAt time.Time `gorm:"not null" json:"changedAt"`
}

func consentAddress(address string) string { return strings.ToLower(strings.TrimSpace(address)) }

// ShouldTrack decides whether one send gets the open pixel and rewritten links.
// The per-email switch (nil = on) always applies. The workspace consent mode
// applies only to marketing email and fails closed: a lookup error sends untracked.
func ShouldTrack(db *gorm.DB, teamID, address string, enabled *bool, marketing bool) bool {
	if enabled != nil && !*enabled {
		return false
	}
	if !marketing {
		return true
	}
	var row struct {
		Mode    string
		Allowed *bool
	}
	// A savepoint keeps a failed lookup from aborting the caller's transaction.
	err := db.Transaction(func(tx *gorm.DB) error {
		return tx.Raw(`SELECT t.tracking_consent_mode AS mode, c.allowed AS allowed FROM teams t
 LEFT JOIN tracking_consents c ON c.team_id=t.id AND c.email=? AND c.is_deleted=false
 WHERE t.id=?`, consentAddress(address), teamID).Scan(&row).Error
	})
	if err != nil {
		return false
	}
	switch row.Mode {
	case TrackingConsentOptIn:
		return row.Allowed != nil && *row.Allowed
	case TrackingConsentOptOut:
		return row.Allowed == nil || *row.Allowed
	}
	return true
}

// TrackingConsentMode returns the workspace mode, or none when unknown.
func TrackingConsentMode(db *gorm.DB, teamID string) (string, error) {
	var mode string
	err := db.Model(&Team{}).Select("tracking_consent_mode").Where("id = ?", teamID).Scan(&mode).Error
	if mode == "" {
		mode = TrackingConsentNone
	}
	return mode, err
}

// TrackingAllowed returns the reader's recorded choice, if any.
func TrackingAllowed(db *gorm.DB, teamID, address string) (*bool, error) {
	var rows []TrackingConsent
	if err := db.Where("team_id = ? AND email = ? AND is_deleted = false", teamID, consentAddress(address)).Limit(1).Find(&rows).Error; err != nil || len(rows) == 0 {
		return nil, err
	}
	return &rows[0].Allowed, nil
}

// SetTrackingConsent records a reader's choice with when and where it was made.
func SetTrackingConsent(db *gorm.DB, teamID, address string, allowed bool, source string) error {
	address = consentAddress(address)
	now := time.Now().UTC()
	var row TrackingConsent
	err := db.Where("team_id = ? AND email = ?", teamID, address).Limit(1).Find(&row).Error
	if err != nil {
		return err
	}
	if row.ID == "" {
		return db.Create(&TrackingConsent{TeamID: teamID, Email: address, Allowed: allowed, Source: source, ChangedAt: now}).Error
	}
	return db.Model(&row).Updates(map[string]any{"allowed": allowed, "source": source, "changed_at": now, "is_deleted": false}).Error
}
