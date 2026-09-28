package services

import (
	"context"
	"fmt"
	"kori/internal/models"
	"time"

	"gorm.io/gorm"
)

// GoalTrackingService handles goal tracking operations
type GoalTrackingService struct {
	base BaseService[models.AutomationGoal]
	db   *gorm.DB
}

// NewGoalTrackingService creates a new goal tracking service
func NewGoalTrackingService(db *gorm.DB) *GoalTrackingService {
	return &GoalTrackingService{
		base: NewBaseService(db, models.AutomationGoal{}),
		db:   db,
	}
}

// CreateGoal creates a new goal for an automation
func (s *GoalTrackingService) CreateGoal(ctx context.Context, goal *models.AutomationGoal) error {
	return s.base.Create(ctx, goal)
}

// GetGoal retrieves a goal by ID
func (s *GoalTrackingService) GetGoal(ctx context.Context, id string) (*models.AutomationGoal, error) {
	return s.base.Get(ctx, id, "Automation")
}

// ListGoals retrieves all goals for an automation
func (s *GoalTrackingService) ListGoals(ctx context.Context, automationID string, activeOnly bool) ([]models.AutomationGoal, error) {
	var goals []models.AutomationGoal
	query := s.db.WithContext(ctx).
		Where("automation_id = ?", automationID)

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("created_at DESC").Find(&goals).Error
	return goals, err
}

// UpdateGoal updates a goal
func (s *GoalTrackingService) UpdateGoal(ctx context.Context, id string, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).
		Model(&models.AutomationGoal{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// DeleteGoal deletes a goal
func (s *GoalTrackingService) DeleteGoal(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete conversions first
		if err := tx.Where("goal_id = ?", id).Delete(&models.GoalConversion{}).Error; err != nil {
			return err
		}

		// Delete goal
		return tx.Where("id = ?", id).Delete(&models.AutomationGoal{}).Error
	})
}

// RecordConversion records a goal conversion
func (s *GoalTrackingService) RecordConversion(ctx context.Context, conversion *models.GoalConversion) error {
	// Check if conversion already exists (prevent duplicates)
	var existing models.GoalConversion
	err := s.db.WithContext(ctx).
		Where("goal_id = ? AND contact_id = ? AND execution_id = ?",
			conversion.GoalID, conversion.ContactID, conversion.ExecutionID).
		First(&existing).Error

	if err == nil {
		// Already recorded
		return nil
	}

	if err != gorm.ErrRecordNotFound {
		return err
	}

	// Create conversion
	return s.db.WithContext(ctx).Create(conversion).Error
}

// CheckAndRecordGoal checks if an event matches a goal and records it
func (s *GoalTrackingService) CheckAndRecordGoal(ctx context.Context, automationID string, contactID string, executionID string, eventType models.GoalType, value float64, metadata map[string]interface{}) error {
	// Get active goals for this automation
	var goals []models.AutomationGoal
	if err := s.db.WithContext(ctx).
		Where("automation_id = ? AND goal_type = ? AND is_active = ?", automationID, eventType, true).
		Find(&goals).Error; err != nil {
		return err
	}

	// Record conversion for each matching goal
	for _, goal := range goals {
		conversion := &models.GoalConversion{
			GoalID:          goal.ID,
			AutomationID:    automationID,
			ContactID:       contactID,
			ExecutionID:     &executionID,
			ConversionValue: value,
		}

		if err := s.RecordConversion(ctx, conversion); err != nil {
			// Log error but continue processing other goals
			continue
		}
	}

	return nil
}

// GetGoalStats retrieves statistics for a specific goal
func (s *GoalTrackingService) GetGoalStats(ctx context.Context, goalID string) (*models.GoalStats, error) {
	var goal models.AutomationGoal
	if err := s.db.WithContext(ctx).Where("id = ?", goalID).First(&goal).Error; err != nil {
		return nil, err
	}

	stats := &models.GoalStats{
		GoalID:   goalID,
		GoalName: goal.Name,
	}

	// Total conversions
	if err := s.db.WithContext(ctx).
		Model(&models.GoalConversion{}).
		Where("goal_id = ?", goalID).
		Count(&stats.TotalConversions).Error; err != nil {
		return nil, err
	}

	// Unique contacts
	if err := s.db.WithContext(ctx).
		Model(&models.GoalConversion{}).
		Where("goal_id = ?", goalID).
		Distinct("contact_id").
		Count(&stats.UniqueContacts).Error; err != nil {
		return nil, err
	}

	// Total value
	if err := s.db.WithContext(ctx).
		Model(&models.GoalConversion{}).
		Where("goal_id = ?", goalID).
		Select("COALESCE(SUM(conversion_value), 0)").
		Scan(&stats.TotalValue).Error; err != nil {
		return nil, err
	}

	// First and last conversion
	var firstConv, lastConv models.GoalConversion
	if err := s.db.WithContext(ctx).
		Where("goal_id = ?", goalID).
		Order("converted_at ASC").
		First(&firstConv).Error; err == nil {
		stats.FirstConversion = &firstConv.ConvertedAt
	}

	if err := s.db.WithContext(ctx).
		Where("goal_id = ?", goalID).
		Order("converted_at DESC").
		First(&lastConv).Error; err == nil {
		stats.LastConversion = &lastConv.ConvertedAt
	}

	// Calculate conversion rate
	var totalExecutions int64
	if err := s.db.WithContext(ctx).
		Model(&models.AutomationExecution{}).
		Where("automation_id = ?", goal.AutomationID).
		Count(&totalExecutions).Error; err == nil && totalExecutions > 0 {
		stats.ConversionRate = float64(stats.TotalConversions) / float64(totalExecutions)
	}

	// Calculate average time to goal
	type TimeResult struct {
		AvgHours float64
	}
	var timeResult TimeResult
	if err := s.db.WithContext(ctx).Raw(`
		SELECT AVG(EXTRACT(EPOCH FROM (gc.converted_at - ae.started_at)) / 3600) as avg_hours
		FROM goal_conversions gc
		JOIN automation_executions ae ON gc.execution_id = ae.id
		WHERE gc.goal_id = ?
	`, goalID).Scan(&timeResult).Error; err == nil {
		stats.AverageTimeToGoal = timeResult.AvgHours
	}

	return stats, nil
}

// GetAutomationGoalSummary retrieves a summary of all goals for an automation
func (s *GoalTrackingService) GetAutomationGoalSummary(ctx context.Context, automationID string) (*models.AutomationGoalSummary, error) {
	summary := &models.AutomationGoalSummary{
		AutomationID: automationID,
	}

	// Total executions
	if err := s.db.WithContext(ctx).
		Model(&models.AutomationExecution{}).
		Where("automation_id = ?", automationID).
		Count(&summary.TotalExecutions).Error; err != nil {
		return nil, err
	}

	// Total conversions across all goals
	if err := s.db.WithContext(ctx).
		Model(&models.GoalConversion{}).
		Where("automation_id = ?", automationID).
		Count(&summary.TotalConversions).Error; err != nil {
		return nil, err
	}

	// Total revenue
	if err := s.db.WithContext(ctx).
		Model(&models.GoalConversion{}).
		Where("automation_id = ?", automationID).
		Select("COALESCE(SUM(conversion_value), 0)").
		Scan(&summary.TotalRevenue).Error; err != nil {
		return nil, err
	}

	// Conversion rate
	if summary.TotalExecutions > 0 {
		summary.ConversionRate = float64(summary.TotalConversions) / float64(summary.TotalExecutions)
	}

	// Get stats for each goal
	var goals []models.AutomationGoal
	if err := s.db.WithContext(ctx).
		Where("automation_id = ? AND is_active = ?", automationID, true).
		Find(&goals).Error; err != nil {
		return nil, err
	}

	summary.Goals = []models.GoalStats{}
	for _, goal := range goals {
		stats, err := s.GetGoalStats(ctx, goal.ID)
		if err != nil {
			continue
		}
		summary.Goals = append(summary.Goals, *stats)
	}

	return summary, nil
}

// GetConversions retrieves conversions for a goal
func (s *GoalTrackingService) GetConversions(ctx context.Context, goalID string, limit int, offset int) ([]models.GoalConversion, int64, error) {
	var conversions []models.GoalConversion
	var total int64

	query := s.db.WithContext(ctx).
		Where("goal_id = ?", goalID).
		Preload("Contact")

	if err := query.Model(&models.GoalConversion{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("converted_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&conversions).Error

	return conversions, total, err
}

// GetContactGoalHistory retrieves all goal conversions for a contact
func (s *GoalTrackingService) GetContactGoalHistory(ctx context.Context, contactID string, limit int, offset int) ([]models.GoalConversion, int64, error) {
	var conversions []models.GoalConversion
	var total int64

	query := s.db.WithContext(ctx).
		Where("contact_id = ?", contactID).
		Preload("Goal").
		Preload("Automation")

	if err := query.Model(&models.GoalConversion{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("converted_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&conversions).Error

	return conversions, total, err
}

// GetGoalsByDateRange retrieves goal conversions within a date range
func (s *GoalTrackingService) GetGoalsByDateRange(ctx context.Context, automationID string, startDate time.Time, endDate time.Time) ([]models.GoalConversion, error) {
	var conversions []models.GoalConversion
	err := s.db.WithContext(ctx).
		Where("automation_id = ? AND converted_at BETWEEN ? AND ?", automationID, startDate, endDate).
		Preload("Goal").
		Preload("Contact").
		Order("converted_at DESC").
		Find(&conversions).Error

	return conversions, err
}

// GetTopConvertingContacts retrieves contacts with the most conversions
func (s *GoalTrackingService) GetTopConvertingContacts(ctx context.Context, automationID string, limit int) ([]ContactConversionStats, error) {
	type Result struct {
		ContactID        string
		ConversionCount  int64
		TotalValue       float64
		LastConvertedAt  time.Time
	}

	var results []Result
	err := s.db.WithContext(ctx).
		Model(&models.GoalConversion{}).
		Select(`
			contact_id,
			COUNT(*) as conversion_count,
			COALESCE(SUM(conversion_value), 0) as total_value,
			MAX(converted_at) as last_converted_at
		`).
		Where("automation_id = ?", automationID).
		Group("contact_id").
		Order("conversion_count DESC").
		Limit(limit).
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	// Load contact details
	stats := make([]ContactConversionStats, len(results))
	for i, result := range results {
		var contact models.Contact
		if err := s.db.WithContext(ctx).Where("id = ?", result.ContactID).First(&contact).Error; err == nil {
			stats[i] = ContactConversionStats{
				Contact:         &contact,
				ConversionCount: result.ConversionCount,
				TotalValue:      result.TotalValue,
				LastConvertedAt: result.LastConvertedAt,
			}
		}
	}

	return stats, nil
}

// CalculateGoalProgress calculates progress towards a target
func (s *GoalTrackingService) CalculateGoalProgress(ctx context.Context, goalID string) (*GoalProgress, error) {
	goal, err := s.GetGoal(ctx, goalID)
	if err != nil {
		return nil, err
	}

	if !goal.HasTarget() {
		return nil, fmt.Errorf("goal does not have a target value")
	}

	stats, err := s.GetGoalStats(ctx, goalID)
	if err != nil {
		return nil, err
	}

	progress := &GoalProgress{
		GoalID:      goalID,
		GoalName:    goal.Name,
		TargetValue: *goal.TargetValue,
	}

	if goal.IsRevenueGoal() {
		progress.CurrentValue = stats.TotalValue
	} else {
		progress.CurrentValue = float64(stats.TotalConversions)
	}

	progress.ProgressPercent = (progress.CurrentValue / progress.TargetValue) * 100
	progress.IsAchieved = progress.CurrentValue >= progress.TargetValue

	return progress, nil
}

// ContactConversionStats represents conversion statistics for a contact
type ContactConversionStats struct {
	Contact         *models.Contact `json:"contact"`
	ConversionCount int64           `json:"conversionCount"`
	TotalValue      float64         `json:"totalValue"`
	LastConvertedAt time.Time       `json:"lastConvertedAt"`
}

// GoalProgress represents progress towards a goal target
type GoalProgress struct {
	GoalID          string  `json:"goalId"`
	GoalName        string  `json:"goalName"`
	CurrentValue    float64 `json:"currentValue"`
	TargetValue     float64 `json:"targetValue"`
	ProgressPercent float64 `json:"progressPercent"`
	IsAchieved      bool    `json:"isAchieved"`
}
