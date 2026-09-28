package services

import (
	"context"
	"fmt"
	"kori/internal/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

// BounceManagementService handles bounce tracking and suppression
type BounceManagementService struct {
	db *gorm.DB
}

// NewBounceManagementService creates a new bounce management service
func NewBounceManagementService(db *gorm.DB) *BounceManagementService {
	return &BounceManagementService{db: db}
}

// ProcessBounce processes a bounce event and handles auto-suppression
func (s *BounceManagementService) ProcessBounce(ctx context.Context, bounce *models.EmailBounce) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Classify bounce if not already classified
		if bounce.BounceType == "" {
			bounceType, category := models.ClassifyBounce(bounce.BounceCode, bounce.BounceMessage)
			bounce.BounceType = bounceType
			bounce.BounceCategory = category
		}

		// Determine if should suppress
		bounce.ShouldSuppress = bounce.IsPermanent()

		// Save bounce record
		if err := tx.Create(bounce).Error; err != nil {
			return err
		}

		// Check bounce rules and apply suppression if needed
		if err := s.checkBounceRules(ctx, tx, bounce); err != nil {
			return err
		}

		// Mark as processed
		now := time.Now()
		bounce.ProcessedAt = &now

		return tx.Save(bounce).Error
	})
}

// checkBounceRules checks if bounce triggers any suppression rules
func (s *BounceManagementService) checkBounceRules(ctx context.Context, tx *gorm.DB, bounce *models.EmailBounce) error {
	// Get active rules for this team and bounce type
	var rules []models.BounceRule
	if err := tx.Where("team_id = ? AND bounce_type = ? AND is_active = ?",
		bounce.TeamID, bounce.BounceType, true).
		Find(&rules).Error; err != nil {
		return err
	}

	for _, rule := range rules {
		if !rule.AutoSuppress {
			continue
		}

		// Count recent bounces within threshold period
		cutoffDate := time.Now().AddDate(0, 0, -rule.ThresholdPeriodDays)
		var count int64
		if err := tx.Model(&models.EmailBounce{}).
			Where("team_id = ? AND email_address = ? AND bounce_type = ? AND created_at >= ?",
				bounce.TeamID, bounce.EmailAddress, bounce.BounceType, cutoffDate).
			Count(&count).Error; err != nil {
			return err
		}

		// If threshold exceeded, add to suppression list
		if int(count) >= rule.ThresholdCount {
			suppression := &models.SuppressionList{
				TeamID:       bounce.TeamID,
				EmailAddress: bounce.EmailAddress,
				ContactID:    &bounce.ContactID,
				Reason:       s.getReason(bounce.BounceType),
				Source:       "bounce_rule",
				Description:  fmt.Sprintf("Auto-suppressed after %d %s bounces", count, bounce.BounceType),
				IsActive:     true,
			}

			// Set expiration for temporary suppressions
			if rule.SuppressionDuration != nil {
				expires := time.Now().AddDate(0, 0, *rule.SuppressionDuration)
				suppression.ExpiresAt = &expires
			}

			// Check if already suppressed
			var existing models.SuppressionList
			err := tx.Where("team_id = ? AND email_address = ? AND is_active = ?",
				bounce.TeamID, bounce.EmailAddress, true).
				First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				// Create new suppression
				if err := tx.Create(suppression).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
			// If already suppressed, no need to create duplicate
		}
	}

	return nil
}

// getReason converts bounce type to suppression reason
func (s *BounceManagementService) getReason(bounceType models.BounceType) models.SuppressionReason {
	switch bounceType {
	case models.BounceTypeHard:
		return models.SuppressionReasonHardBounce
	case models.BounceTypeSoft:
		return models.SuppressionReasonMultipleSoftBounces
	default:
		return models.SuppressionReasonHardBounce
	}
}

// ProcessComplaint processes a spam complaint
func (s *BounceManagementService) ProcessComplaint(ctx context.Context, complaint *models.ComplaintReport) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Save complaint
		if err := tx.Create(complaint).Error; err != nil {
			return err
		}

		// Auto-suppress if enabled
		if complaint.AutoSuppressed {
			suppression := &models.SuppressionList{
				TeamID:       complaint.TeamID,
				EmailAddress: complaint.EmailAddress,
				ContactID:    complaint.ContactID,
				Reason:       models.SuppressionReasonComplaint,
				Source:       "complaint",
				Description:  fmt.Sprintf("Spam complaint: %s", complaint.ComplaintType),
				IsActive:     true,
			}

			// Check if already suppressed
			var existing models.SuppressionList
			err := tx.Where("team_id = ? AND email_address = ? AND is_active = ?",
				complaint.TeamID, complaint.EmailAddress, true).
				First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				if err := tx.Create(suppression).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}

		// Mark as processed
		now := time.Now()
		complaint.ProcessedAt = &now

		return tx.Save(complaint).Error
	})
}

// IsEmailSuppressed checks if an email address is on the suppression list
func (s *BounceManagementService) IsEmailSuppressed(ctx context.Context, teamID string, emailAddress string) (bool, *models.SuppressionList, error) {
	var suppression models.SuppressionList
	err := s.db.WithContext(ctx).
		Where("team_id = ? AND email_address = ? AND is_active = ?", teamID, strings.ToLower(emailAddress), true).
		First(&suppression).Error

	if err == gorm.ErrRecordNotFound {
		return false, nil, nil
	}

	if err != nil {
		return false, nil, err
	}

	// Check if expired
	if !suppression.IsActiveNow() {
		return false, &suppression, nil
	}

	return true, &suppression, nil
}

// AddToSuppressionList manually adds an email to suppression list
func (s *BounceManagementService) AddToSuppressionList(ctx context.Context, suppression *models.SuppressionList) error {
	// Normalize email
	suppression.EmailAddress = strings.ToLower(suppression.EmailAddress)

	// Check if already suppressed
	var existing models.SuppressionList
	err := s.db.WithContext(ctx).
		Where("team_id = ? AND email_address = ?", suppression.TeamID, suppression.EmailAddress).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new
		return s.db.WithContext(ctx).Create(suppression).Error
	}

	if err != nil {
		return err
	}

	// Update existing
	return s.db.WithContext(ctx).
		Model(&existing).
		Updates(map[string]interface{}{
			"reason":      suppression.Reason,
			"description": suppression.Description,
			"is_active":   true,
			"expires_at":  suppression.ExpiresAt,
		}).Error
}

// RemoveFromSuppressionList removes an email from suppression list
func (s *BounceManagementService) RemoveFromSuppressionList(ctx context.Context, teamID string, emailAddress string) error {
	now := time.Now()
	return s.db.WithContext(ctx).
		Model(&models.SuppressionList{}).
		Where("team_id = ? AND email_address = ?", teamID, strings.ToLower(emailAddress)).
		Updates(map[string]interface{}{
			"is_active":      false,
			"reactivated_at": now,
		}).Error
}

// GetSuppressionList retrieves the suppression list for a team
func (s *BounceManagementService) GetSuppressionList(ctx context.Context, teamID string, activeOnly bool, limit int, offset int) ([]models.SuppressionList, int64, error) {
	var suppressions []models.SuppressionList
	var total int64

	query := s.db.WithContext(ctx).
		Where("team_id = ?", teamID).
		Preload("Contact")

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Model(&models.SuppressionList{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("added_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&suppressions).Error

	return suppressions, total, err
}

// GetBounceHistory retrieves bounce history for an email address
func (s *BounceManagementService) GetBounceHistory(ctx context.Context, teamID string, emailAddress string, limit int) ([]models.EmailBounce, error) {
	var bounces []models.EmailBounce
	err := s.db.WithContext(ctx).
		Where("team_id = ? AND email_address = ?", teamID, strings.ToLower(emailAddress)).
		Preload("Email").
		Preload("Contact").
		Order("created_at DESC").
		Limit(limit).
		Find(&bounces).Error

	return bounces, err
}

// GetBounceStats retrieves bounce statistics for a team
func (s *BounceManagementService) GetBounceStats(ctx context.Context, teamID string, startDate, endDate time.Time) (*BounceStats, error) {
	stats := &BounceStats{
		TeamID:    teamID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Total bounces
	if err := s.db.WithContext(ctx).
		Model(&models.EmailBounce{}).
		Where("team_id = ? AND created_at BETWEEN ? AND ?", teamID, startDate, endDate).
		Count(&stats.TotalBounces).Error; err != nil {
		return nil, err
	}

	// Hard bounces
	if err := s.db.WithContext(ctx).
		Model(&models.EmailBounce{}).
		Where("team_id = ? AND bounce_type = ? AND created_at BETWEEN ? AND ?",
			teamID, models.BounceTypeHard, startDate, endDate).
		Count(&stats.HardBounces).Error; err != nil {
		return nil, err
	}

	// Soft bounces
	if err := s.db.WithContext(ctx).
		Model(&models.EmailBounce{}).
		Where("team_id = ? AND bounce_type = ? AND created_at BETWEEN ? AND ?",
			teamID, models.BounceTypeSoft, startDate, endDate).
		Count(&stats.SoftBounces).Error; err != nil {
		return nil, err
	}

	// Total emails sent (from email_trackings or deliveries)
	var totalSent int64
	if err := s.db.WithContext(ctx).
		Model(&models.Email{}).
		Where("team_id = ? AND created_at BETWEEN ? AND ?", teamID, startDate, endDate).
		Count(&totalSent).Error; err != nil {
		return nil, err
	}

	stats.TotalSent = totalSent

	// Calculate rates
	if totalSent > 0 {
		stats.BounceRate = float64(stats.TotalBounces) / float64(totalSent)
		stats.HardBounceRate = float64(stats.HardBounces) / float64(totalSent)
		stats.SoftBounceRate = float64(stats.SoftBounces) / float64(totalSent)
	}

	// Active suppressions
	if err := s.db.WithContext(ctx).
		Model(&models.SuppressionList{}).
		Where("team_id = ? AND is_active = ?", teamID, true).
		Count(&stats.ActiveSuppressions).Error; err != nil {
		return nil, err
	}

	// Complaints
	if err := s.db.WithContext(ctx).
		Model(&models.ComplaintReport{}).
		Where("team_id = ? AND created_at BETWEEN ? AND ?", teamID, startDate, endDate).
		Count(&stats.Complaints).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// InitializeBounceRules initializes default bounce rules for a team
func (s *BounceManagementService) InitializeBounceRules(ctx context.Context, teamID string) error {
	defaultRules := models.DefaultBounceRules(teamID)

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, rule := range defaultRules {
			// Check if rule already exists
			var existing models.BounceRule
			err := tx.Where("team_id = ? AND name = ?", teamID, rule.Name).
				First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				if err := tx.Create(&rule).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// BounceStats represents bounce statistics
type BounceStats struct {
	TeamID              string    `json:"teamId"`
	StartDate           time.Time `json:"startDate"`
	EndDate             time.Time `json:"endDate"`
	TotalSent           int64     `json:"totalSent"`
	TotalBounces        int64     `json:"totalBounces"`
	HardBounces         int64     `json:"hardBounces"`
	SoftBounces         int64     `json:"softBounces"`
	BounceRate          float64   `json:"bounceRate"`
	HardBounceRate      float64   `json:"hardBounceRate"`
	SoftBounceRate      float64   `json:"softBounceRate"`
	ActiveSuppressions  int64     `json:"activeSuppressions"`
	Complaints          int64     `json:"complaints"`
}
