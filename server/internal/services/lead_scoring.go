package services

import (
	"context"
	"fmt"
	"kori/internal/models"
	"time"

	"gorm.io/gorm"
)

// LeadScoringService handles lead scoring operations
type LeadScoringService struct {
	base BaseService[models.LeadScore]
	db   *gorm.DB
}

// NewLeadScoringService creates a new lead scoring service
func NewLeadScoringService(db *gorm.DB) *LeadScoringService {
	return &LeadScoringService{
		base: NewBaseService(db, models.LeadScore{}),
		db:   db,
	}
}

// GetContactScore retrieves the score for a contact, creating it if it doesn't exist
func (s *LeadScoringService) GetContactScore(ctx context.Context, contactID string) (*models.LeadScore, error) {
	var score models.LeadScore
	err := s.db.WithContext(ctx).
		Preload("Contact").
		Where("contact_id = ?", contactID).
		First(&score).Error

	if err == gorm.ErrRecordNotFound {
		// Create initial score
		score = models.LeadScore{
			ContactID: contactID,
			Score:     0,
			Grade:     models.GradeF,
		}
		if err := s.db.WithContext(ctx).Create(&score).Error; err != nil {
			return nil, err
		}
		return &score, nil
	}

	return &score, err
}

// AddPoints adds points to a contact's score and logs the activity
func (s *LeadScoringService) AddPoints(ctx context.Context, contactID string, activityType string, points int, description string, metadata map[string]interface{}) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get or create contact score
		var score models.LeadScore
		err := tx.Where("contact_id = ?", contactID).First(&score).Error
		if err == gorm.ErrRecordNotFound {
			score = models.LeadScore{
				ContactID: contactID,
				Score:     0,
			}
		} else if err != nil {
			return err
		}

		// Calculate old grade
		oldGrade := score.Grade

		// Update score
		score.Score += points
		if score.Score < 0 {
			score.Score = 0 // Prevent negative scores
		}
		now := time.Now()
		score.LastActivityAt = &now

		// Save score (BeforeSave hook will calculate grade)
		if err := tx.Save(&score).Error; err != nil {
			return err
		}

		// Log activity
		activity := models.ScoreActivity{
			ContactID:    contactID,
			ActivityType: activityType,
			Points:       points,
			Description:  description,
		}

		// Extract metadata if provided
		if automationID, ok := metadata["automation_id"].(string); ok {
			activity.AutomationID = &automationID
		}
		if campaignID, ok := metadata["campaign_id"].(string); ok {
			activity.CampaignID = &campaignID
		}
		if emailID, ok := metadata["email_id"].(string); ok {
			activity.EmailID = &emailID
		}

		if err := tx.Create(&activity).Error; err != nil {
			return err
		}

		// Emit grade change event if grade changed
		if oldGrade != score.Grade {
			// TODO: Emit event for grade change
			// events.Emit("score.grade_changed", map[string]interface{}{
			// 	"contact_id": contactID,
			// 	"old_grade":  oldGrade,
			// 	"new_grade":  score.Grade,
			// 	"score":      score.Score,
			// })
		}

		return nil
	})
}

// ProcessActivity processes an activity and adds appropriate points based on score rules
func (s *LeadScoringService) ProcessActivity(ctx context.Context, teamID string, contactID string, activityType string, description string, metadata map[string]interface{}) error {
	// Get score rule for this activity
	var rule models.ScoreRule
	err := s.db.WithContext(ctx).
		Where("team_id = ? AND activity_type = ? AND is_active = ?", teamID, activityType, true).
		First(&rule).Error

	if err == gorm.ErrRecordNotFound {
		// No rule found, skip scoring
		return nil
	}
	if err != nil {
		return err
	}

	// Add points
	return s.AddPoints(ctx, contactID, activityType, rule.Points, description, metadata)
}

// GetScoreHistory retrieves the score activity history for a contact
func (s *LeadScoringService) GetScoreHistory(ctx context.Context, contactID string, limit int, offset int) ([]models.ScoreActivity, int64, error) {
	var activities []models.ScoreActivity
	var total int64

	query := s.db.WithContext(ctx).
		Model(&models.ScoreActivity{}).
		Where("contact_id = ?", contactID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&activities).Error

	return activities, total, err
}

// GetTopScoredContacts retrieves contacts with highest scores
func (s *LeadScoringService) GetTopScoredContacts(ctx context.Context, teamID string, limit int) ([]models.LeadScore, error) {
	var scores []models.LeadScore
	err := s.db.WithContext(ctx).
		Joins("JOIN contacts ON contacts.id = lead_scores.contact_id").
		Where("contacts.team_id = ?", teamID).
		Preload("Contact").
		Order("score DESC").
		Limit(limit).
		Find(&scores).Error

	return scores, err
}

// GetContactsByGrade retrieves all contacts with a specific grade
func (s *LeadScoringService) GetContactsByGrade(ctx context.Context, teamID string, grade models.ScoreGrade, limit int, offset int) ([]models.LeadScore, int64, error) {
	var scores []models.LeadScore
	var total int64

	query := s.db.WithContext(ctx).
		Joins("JOIN contacts ON contacts.id = lead_scores.contact_id").
		Where("contacts.team_id = ? AND lead_scores.grade = ?", teamID, grade)

	if err := query.Model(&models.LeadScore{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Contact").
		Order("score DESC").
		Limit(limit).
		Offset(offset).
		Find(&scores).Error

	return scores, total, err
}

// GetScoreDistribution returns the count of contacts per grade
func (s *LeadScoringService) GetScoreDistribution(ctx context.Context, teamID string) (map[models.ScoreGrade]int64, error) {
	type Result struct {
		Grade models.ScoreGrade
		Count int64
	}

	var results []Result
	err := s.db.WithContext(ctx).
		Model(&models.LeadScore{}).
		Select("grade, COUNT(*) as count").
		Joins("JOIN contacts ON contacts.id = lead_scores.contact_id").
		Where("contacts.team_id = ?", teamID).
		Group("grade").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	distribution := make(map[models.ScoreGrade]int64)
	for _, result := range results {
		distribution[result.Grade] = result.Count
	}

	return distribution, nil
}

// DecayScores applies decay to scores based on inactivity
func (s *LeadScoringService) DecayScores(ctx context.Context) error {
	// Get all score rules with decay configured
	var rules []models.ScoreRule
	err := s.db.WithContext(ctx).
		Where("decay_days IS NOT NULL AND decay_days > 0 AND is_active = ?", true).
		Find(&rules).Error

	if err != nil {
		return err
	}

	for _, rule := range rules {
		// Find activities older than decay_days
		cutoffDate := time.Now().AddDate(0, 0, -*rule.DecayDays)

		var activities []models.ScoreActivity
		err := s.db.WithContext(ctx).
			Where("activity_type = ? AND created_at < ? AND created_at >= ?",
				rule.ActivityType,
				cutoffDate,
				cutoffDate.AddDate(0, 0, -1), // Only process activities from one day
			).
			Find(&activities).Error

		if err != nil {
			continue
		}

		// Decay points for each activity
		for _, activity := range activities {
			if err := s.AddPoints(ctx, activity.ContactID, "score_decay", -rule.Points,
				fmt.Sprintf("Decay for %s activity older than %d days", rule.ActivityType, *rule.DecayDays),
				map[string]interface{}{
					"original_activity_id": activity.ID,
					"decay_rule_id":        rule.ID,
				}); err != nil {
				// Log error but continue processing
				continue
			}
		}
	}

	return nil
}

// ScoreRuleService handles score rule operations
type ScoreRuleService struct {
	base BaseService[models.ScoreRule]
	db   *gorm.DB
}

// NewScoreRuleService creates a new score rule service
func NewScoreRuleService(db *gorm.DB) *ScoreRuleService {
	return &ScoreRuleService{
		base: NewBaseService(db, models.ScoreRule{}),
		db:   db,
	}
}

// Create creates a new score rule
func (s *ScoreRuleService) Create(ctx context.Context, rule *models.ScoreRule) error {
	return s.base.Create(ctx, rule)
}

// GetByID retrieves a score rule by ID
func (s *ScoreRuleService) GetByID(ctx context.Context, id string) (*models.ScoreRule, error) {
	return s.base.Get(ctx, id)
}

// GetTeamRules retrieves all score rules for a team
func (s *ScoreRuleService) GetTeamRules(ctx context.Context, teamID string, activeOnly bool) ([]models.ScoreRule, error) {
	var rules []models.ScoreRule
	query := s.db.WithContext(ctx).Where("team_id = ?", teamID)

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("activity_type ASC").Find(&rules).Error
	return rules, err
}

// GetRuleByActivity retrieves a specific rule for an activity type
func (s *ScoreRuleService) GetRuleByActivity(ctx context.Context, teamID string, activityType string) (*models.ScoreRule, error) {
	var rule models.ScoreRule
	err := s.db.WithContext(ctx).
		Where("team_id = ? AND activity_type = ?", teamID, activityType).
		First(&rule).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &rule, err
}

// InitializeDefaultRules creates default score rules for a new team
func (s *ScoreRuleService) InitializeDefaultRules(ctx context.Context, teamID string) error {
	defaultRules := models.DefaultScoreRules(teamID)

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, rule := range defaultRules {
			// Check if rule already exists
			var existing models.ScoreRule
			err := tx.Where("team_id = ? AND activity_type = ?", teamID, rule.ActivityType).
				First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				// Create new rule
				if err := tx.Create(&rule).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// UpdateRule updates a score rule
func (s *ScoreRuleService) UpdateRule(ctx context.Context, id string, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).
		Model(&models.ScoreRule{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// DeleteRule soft deletes a score rule by setting is_active to false
func (s *ScoreRuleService) DeleteRule(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).
		Model(&models.ScoreRule{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}
