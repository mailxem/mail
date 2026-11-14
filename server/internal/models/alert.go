package models

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Alert is a workspace's in-app copy of a service alert, or news such as a
// finished import. Whether its problem has cleared is worked out when it is
// listed, from the campaign, domain or import it is about.
type Alert struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	TeamID      string     `gorm:"index:alert_team_recent,priority:1;not null" json:"-"`
	Kind        string     `gorm:"not null" json:"kind"`
	ReferenceID string     `json:"-"`
	Reason      string     `json:"reason"` // sanitized like ServiceNoticeEvent.Reason
	EventCount  int64      `json:"count"`
	ReadAt      *time.Time `json:"readAt"`
	CreatedAt   time.Time  `gorm:"index:alert_team_recent,priority:2" json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// RecordAlert keeps one alert per workspace, kind and group. A repeat in the
// same group only refreshes its count, so it doesn't turn unread again.
func RecordAlert(tx *gorm.DB, team, kind, group, reference, reason string, count int64, now time.Time) error {
	a := Alert{ID: noticeKey("alert", team, kind, group), TeamID: team, Kind: kind, ReferenceID: reference, Reason: SanitizeReason(reason), EventCount: count, CreatedAt: now, UpdatedAt: now}
	return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.Assignments(map[string]any{"event_count": count, "updated_at": now})}).Create(&a).Error
}
