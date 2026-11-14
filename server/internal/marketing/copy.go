package marketing

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/models"
)

// ErrNotSent is returned for a follow-up audience on a campaign that has not
// reached anyone yet.
var ErrNotSent = errors.New("this campaign hasn't reached anyone yet, so there is no one to follow up with")

// CopyCampaign returns a new one-off draft with src's content, audience,
// sender and settings. Status, schedule, counters and staging start fresh.
func CopyCampaign(src models.Campaign) models.Campaign {
	return models.Campaign{
		TeamID:              src.TeamID,
		Name:                src.Name + " (copy)",
		Description:         src.Description,
		TemplateID:          src.TemplateID,
		ListID:              src.ListID,
		SMTPConfigID:        src.SMTPConfigID,
		Subject:             src.Subject,
		HTMLBody:            src.HTMLBody,
		ReplyTo:             src.ReplyTo,
		PostalAddress:       src.PostalAddress,
		TrackEngagement:     src.TrackEngagement,
		ArchiveNewsletterID: src.ArchiveNewsletterID,
		Data:                src.Data,
		BatchSize:           src.BatchSize,
		BatchDelay:          src.BatchDelay,
		Timezone:            src.Timezone,
		Status:              models.CampaignStatusDraft,
		Schedule:            models.CampaignScheduleOneTime,
	}
}

// DuplicateCampaign saves a draft copy of the team's campaign id. A non-empty
// audience sends the copy only to readers of id who did not open or click it.
// The draft never sends on its own.
func (s *Service) DuplicateCampaign(ctx context.Context, team, id string, audience models.AudienceFilter) (models.Campaign, error) {
	db := s.DB.WithContext(ctx)
	var src models.Campaign
	if err := Scope(db, team).First(&src, "id = ?", id).Error; err != nil {
		return src, err
	}
	c := CopyCampaign(src)
	if audience != "" {
		if audience != models.AudienceNotOpened && audience != models.AudienceNotClicked {
			return c, errors.New("audience must be not_opened or not_clicked")
		}
		var reached int64
		if err := db.Model(&models.Email{}).Where("campaign_id = ? AND test = ? AND status IN ?", src.ID, false, reachedStatuses).Count(&reached).Error; err != nil {
			return c, err
		}
		if reached == 0 {
			return c, ErrNotSent
		}
		c.AudienceCampaignID, c.AudienceFilter = &src.ID, audience
	}
	now := time.Now().UTC()
	c.Base = models.Base{ID: uuid.NewString(), CreatedAt: now, UpdatedAt: now}
	// SkipHooks: creating a campaign normally queues it to send.
	return c, db.Session(&gorm.Session{SkipHooks: true}).Omit(clause.Associations).Create(&c).Error
}
