package services

import (
	"context"
	"encoding/json"
	"fmt"
	"kori/internal/models"
	"time"

	"gorm.io/gorm"
)

// SendTimeOptimizationService handles send time optimization
type SendTimeOptimizationService struct {
	db *gorm.DB
}

// NewSendTimeOptimizationService creates a new send time optimization service
func NewSendTimeOptimizationService(db *gorm.DB) *SendTimeOptimizationService {
	return &SendTimeOptimizationService{db: db}
}

// RecordEmailOpen records an email open event for learning
func (s *SendTimeOptimizationService) RecordEmailOpen(ctx context.Context, event *models.EmailOpenEvent) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Save open event
		if err := tx.Create(event).Error; err != nil {
			return err
		}

		// Update engagement pattern
		return s.updateEngagementPattern(ctx, tx, event)
	})
}

// updateEngagementPattern updates the contact's engagement pattern
func (s *SendTimeOptimizationService) updateEngagementPattern(ctx context.Context, tx *gorm.DB, event *models.EmailOpenEvent) error {
	// Get or create pattern
	var pattern models.ContactEngagementPattern
	err := tx.Where("contact_id = ?", event.ContactID).First(&pattern).Error

	if err == gorm.ErrRecordNotFound {
		// Create new pattern
		pattern = models.ContactEngagementPattern{
			ContactID:        event.ContactID,
			Timezone:         event.OpenedTimezone,
			TimezoneDetected: event.OpenedTimezone != "",
		}
	} else if err != nil {
		return err
	}

	// Update hourly engagement
	var hourlyData map[string]int
	if pattern.HourlyEngagement != nil {
		json.Unmarshal(pattern.HourlyEngagement, &hourlyData)
	} else {
		hourlyData = make(map[string]int)
	}
	hourKey := fmt.Sprintf("%d", event.OpenedHour)
	hourlyData[hourKey]++
	hourlyJSON, _ := json.Marshal(hourlyData)
	pattern.HourlyEngagement = hourlyJSON

	// Update daily engagement
	var dailyData map[string]int
	if pattern.DailyEngagement != nil {
		json.Unmarshal(pattern.DailyEngagement, &dailyData)
	} else {
		dailyData = make(map[string]int)
	}
	dayKey := fmt.Sprintf("%d", event.OpenedDayOfWeek)
	dailyData[dayKey]++
	dailyJSON, _ := json.Marshal(dailyData)
	pattern.DailyEngagement = dailyJSON

	// Update heatmap (day+hour combination)
	var heatmapData map[string]int
	if pattern.HeatmapData != nil {
		json.Unmarshal(pattern.HeatmapData, &heatmapData)
	} else {
		heatmapData = make(map[string]int)
	}
	heatmapKey := fmt.Sprintf("%d_%d", event.OpenedDayOfWeek, event.OpenedHour)
	heatmapData[heatmapKey]++
	heatmapJSON, _ := json.Marshal(heatmapData)
	pattern.HeatmapData = heatmapJSON

	// Update stats
	pattern.TotalOpens++
	pattern.LastOpenAt = &event.OpenedAt

	// Recalculate optimal time
	s.calculateOptimalTime(&pattern, hourlyData, dailyData)

	// Calculate data quality
	pattern.DataQuality = pattern.CalculateDataQuality()

	now := time.Now()
	pattern.LastCalculatedAt = &now

	// Save pattern
	return tx.Save(&pattern).Error
}

// calculateOptimalTime finds the hour/day with most opens
func (s *SendTimeOptimizationService) calculateOptimalTime(pattern *models.ContactEngagementPattern, hourlyData map[string]int, dailyData map[string]int) {
	// Find hour with most opens
	maxHourOpens := 0
	optimalHour := 10 // Default 10 AM

	for hourStr, count := range hourlyData {
		if count > maxHourOpens {
			maxHourOpens = count
			fmt.Sscanf(hourStr, "%d", &optimalHour)
		}
	}

	// Find day with most opens
	maxDayOpens := 0
	optimalDay := 2 // Default Tuesday

	for dayStr, count := range dailyData {
		if count > maxDayOpens {
			maxDayOpens = count
			fmt.Sscanf(dayStr, "%d", &optimalDay)
		}
	}

	pattern.OptimalHour = &optimalHour
	pattern.OptimalDayOfWeek = &optimalDay

	// Calculate confidence score based on data spread
	totalOpens := pattern.TotalOpens
	if totalOpens > 0 {
		topHourPercentage := float64(maxHourOpens) / float64(totalOpens)
		pattern.OptimalScore = topHourPercentage * 100
	}
}

// GetOptimalSendTime calculates the optimal send time for a contact
func (s *SendTimeOptimizationService) GetOptimalSendTime(ctx context.Context, contactID string, requestedTime time.Time, teamID string) (time.Time, string, error) {
	// Get team defaults
	var teamDefaults models.TeamSendTimeDefaults
	err := s.db.WithContext(ctx).Where("team_id = ?", teamID).First(&teamDefaults).Error
	if err == gorm.ErrRecordNotFound {
		// No team defaults, use requested time
		return requestedTime, "fixed_time", nil
	} else if err != nil {
		return requestedTime, "", err
	}

	// Get contact pattern
	var pattern models.ContactEngagementPattern
	err = s.db.WithContext(ctx).Where("contact_id = ?", contactID).First(&pattern).Error

	if err == gorm.ErrRecordNotFound || !pattern.HasSufficientData(teamDefaults.MinimumOpensRequired) {
		// Not enough data, use team default
		return s.getTeamDefaultTime(requestedTime, &teamDefaults), "team_default", nil
	}

	if err != nil {
		return requestedTime, "", err
	}

	// Use contact's optimal time
	optimalTime := pattern.GetOptimalSendTime(requestedTime)

	// Ensure within team send window
	if !teamDefaults.IsWithinSendWindow(optimalTime) {
		optimalTime = teamDefaults.GetNextAllowedSendTime(optimalTime)
	}

	return optimalTime, "contact_pattern", nil
}

// getTeamDefaultTime returns the default send time for the team
func (s *SendTimeOptimizationService) getTeamDefaultTime(requested time.Time, defaults *models.TeamSendTimeDefaults) time.Time {
	loc, _ := time.LoadLocation(defaults.DefaultTimezone)
	t := requested.In(loc)

	// If team has calculated average, use it
	if defaults.TeamAverageOptimalHour != nil {
		return time.Date(
			t.Year(), t.Month(), t.Day(),
			*defaults.TeamAverageOptimalHour,
			0, 0, 0, loc,
		)
	}

	// Use default send hour
	return time.Date(
		t.Year(), t.Month(), t.Day(),
		defaults.DefaultSendHour,
		0, 0, 0, loc,
	)
}

// QueueEmailWithOptimization queues an email with send time optimization
func (s *SendTimeOptimizationService) QueueEmailWithOptimization(ctx context.Context, queueItem *models.SendTimeQueue) error {
	// Calculate optimal send time
	optimalTime, strategy, err := s.GetOptimalSendTime(
		ctx,
		queueItem.ContactID,
		queueItem.RequestedSendAt,
		queueItem.TeamID,
	)
	if err != nil {
		return err
	}

	queueItem.OptimizedSendAt = optimalTime
	queueItem.OptimizationUsed = strategy
	queueItem.Status = "QUEUED"

	return s.db.WithContext(ctx).Create(queueItem).Error
}

// ProcessSendQueue processes queued emails ready to send
func (s *SendTimeOptimizationService) ProcessSendQueue(ctx context.Context, batchSize int) error {
	now := time.Now()

	// Get emails ready to send
	var queueItems []models.SendTimeQueue
	if err := s.db.WithContext(ctx).
		Where("status = ? AND optimized_send_at <= ?", "QUEUED", now).
		Limit(batchSize).
		Find(&queueItems).Error; err != nil {
		return err
	}

	for _, item := range queueItems {
		if err := s.processSendQueueItem(ctx, &item); err != nil {
			// Log error but continue processing
			continue
		}
	}

	return nil
}

// processSendQueueItem processes a single queue item
func (s *SendTimeOptimizationService) processSendQueueItem(ctx context.Context, item *models.SendTimeQueue) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Here you would actually send the email
		// For now, just mark as sent

		now := time.Now()
		item.Status = "SENT"
		item.SentAt = &now
		item.ProcessedAt = &now

		return tx.Save(item).Error
	})
}

// GetContactEngagementPattern retrieves engagement pattern for a contact
func (s *SendTimeOptimizationService) GetContactEngagementPattern(ctx context.Context, contactID string) (*models.ContactEngagementPattern, error) {
	var pattern models.ContactEngagementPattern
	err := s.db.WithContext(ctx).
		Where("contact_id = ?", contactID).
		First(&pattern).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &pattern, err
}

// InitializeTeamDefaults initializes default send time settings for a team
func (s *SendTimeOptimizationService) InitializeTeamDefaults(ctx context.Context, teamID string) error {
	var existing models.TeamSendTimeDefaults
	err := s.db.WithContext(ctx).Where("team_id = ?", teamID).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create defaults
		defaults := models.TeamSendTimeDefaults{
			TeamID:               teamID,
			DefaultTimezone:      "UTC",
			DefaultSendHour:      10,
			EarliestSendHour:     8,
			LatestSendHour:       20,
			SendMonday:           true,
			SendTuesday:          true,
			SendWednesday:        true,
			SendThursday:         true,
			SendFriday:           true,
			SendSaturday:         false,
			SendSunday:           false,
			UseSTOByDefault:      false,
			MinimumOpensRequired: 5,
			FallbackStrategy:     "team_average",
		}

		return s.db.WithContext(ctx).Create(&defaults).Error
	}

	return err
}

// UpdateTeamDefaults updates team send time defaults
func (s *SendTimeOptimizationService) UpdateTeamDefaults(ctx context.Context, teamID string, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).
		Model(&models.TeamSendTimeDefaults{}).
		Where("team_id = ?", teamID).
		Updates(updates).Error
}

// CalculateTeamAverageOptimalTime calculates average optimal time across all contacts
func (s *SendTimeOptimizationService) CalculateTeamAverageOptimalTime(ctx context.Context, teamID string) error {
	// Get all contact patterns for contacts in this team
	var patterns []models.ContactEngagementPattern
	if err := s.db.WithContext(ctx).
		Joins("JOIN contacts ON contacts.id = contact_engagement_patterns.contact_id").
		Where("contacts.team_id = ? AND contact_engagement_patterns.total_opens >= ?", teamID, 5).
		Find(&patterns).Error; err != nil {
		return err
	}

	if len(patterns) == 0 {
		return nil
	}

	// Calculate average optimal hour
	totalHour := 0
	totalDay := 0
	count := 0

	for _, pattern := range patterns {
		if pattern.OptimalHour != nil {
			totalHour += *pattern.OptimalHour
			count++
		}
		if pattern.OptimalDayOfWeek != nil {
			totalDay += *pattern.OptimalDayOfWeek
		}
	}

	if count == 0 {
		return nil
	}

	avgHour := totalHour / count
	avgDay := totalDay / count

	// Update team defaults
	return s.db.WithContext(ctx).
		Model(&models.TeamSendTimeDefaults{}).
		Where("team_id = ?", teamID).
		Updates(map[string]interface{}{
			"team_average_optimal_hour": avgHour,
			"team_average_optimal_day":  avgDay,
		}).Error
}

// GetEngagementStats retrieves engagement statistics
func (s *SendTimeOptimizationService) GetEngagementStats(ctx context.Context, teamID string) (*EngagementStats, error) {
	stats := &EngagementStats{
		TeamID: teamID,
	}

	// Total contacts with patterns
	if err := s.db.WithContext(ctx).
		Model(&models.ContactEngagementPattern{}).
		Joins("JOIN contacts ON contacts.id = contact_engagement_patterns.contact_id").
		Where("contacts.team_id = ?", teamID).
		Count(&stats.TotalContactsWithData).Error; err != nil {
		return nil, err
	}

	// Contacts with sufficient data
	if err := s.db.WithContext(ctx).
		Model(&models.ContactEngagementPattern{}).
		Joins("JOIN contacts ON contacts.id = contact_engagement_patterns.contact_id").
		Where("contacts.team_id = ? AND contact_engagement_patterns.total_opens >= ?", teamID, 5).
		Count(&stats.ContactsWithSufficientData).Error; err != nil {
		return nil, err
	}

	// Average opens per contact
	var result struct {
		AvgOpens float64
	}
	if err := s.db.WithContext(ctx).
		Model(&models.ContactEngagementPattern{}).
		Select("AVG(total_opens) as avg_opens").
		Joins("JOIN contacts ON contacts.id = contact_engagement_patterns.contact_id").
		Where("contacts.team_id = ?", teamID).
		Scan(&result).Error; err != nil {
		return nil, err
	}
	stats.AverageOpensPerContact = result.AvgOpens

	// Queued emails
	if err := s.db.WithContext(ctx).
		Model(&models.SendTimeQueue{}).
		Where("team_id = ? AND status = ?", teamID, "QUEUED").
		Count(&stats.EmailsInQueue).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// EngagementStats represents engagement statistics
type EngagementStats struct {
	TeamID                     string  `json:"teamId"`
	TotalContactsWithData      int64   `json:"totalContactsWithData"`
	ContactsWithSufficientData int64   `json:"contactsWithSufficientData"`
	AverageOpensPerContact     float64 `json:"averageOpensPerContact"`
	EmailsInQueue              int64   `json:"emailsInQueue"`
}
