package models

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SummaryNoticeKind is the daily or weekly sending summary. Its ReferenceID
// names the period, so each recipient gets at most one email per period.
const SummaryNoticeKind = "summary"

func QueueSummaryNotice(tx *gorm.DB, user User, period string, now time.Time) error {
	n := ServiceNotice{Key: noticeKey(SummaryNoticeKind, user.ID, period), Kind: SummaryNoticeKind, TeamID: user.TeamID, UserID: user.ID, AccountEmail: user.Email, ReferenceID: period, Status: "QUEUED", NextAttemptAt: now, CreatedAt: now, UpdatedAt: now}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&n).Error
}
