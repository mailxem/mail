package services

import (
	"context"
	"errors"
	"fmt"
	"kori/internal/models"
	"math"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

// ABTestService handles A/B testing operations
type ABTestService struct {
	base BaseService[models.ABTest]
	db   *gorm.DB
}

// NewABTestService creates a new A/B test service
func NewABTestService(db *gorm.DB) *ABTestService {
	return &ABTestService{
		base: NewBaseService(db, models.ABTest{}),
		db:   db,
	}
}

// Create creates a new A/B test
func (s *ABTestService) Create(ctx context.Context, test *models.ABTest) error {
	return s.base.Create(ctx, test)
}

// Get retrieves an A/B test by ID
func (s *ABTestService) Get(ctx context.Context, id string) (*models.ABTest, error) {
	return s.base.Get(ctx, id, "Variants", "Variants.Result", "WinnerVariant")
}

// List retrieves all A/B tests for a team
func (s *ABTestService) List(ctx context.Context, teamID string, status models.TestStatus, page, limit int) ([]models.ABTest, int64, error) {
	var tests []models.ABTest
	var total int64

	query := s.db.WithContext(ctx).
		Where("team_id = ?", teamID).
		Preload("Variants").
		Preload("WinnerVariant")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Model(&models.ABTest{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&tests).Error

	return tests, total, err
}

// Update updates an A/B test
func (s *ABTestService) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).
		Model(&models.ABTest{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Delete deletes an A/B test
func (s *ABTestService) Delete(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete variant assignments
		if err := tx.Where("test_id = ?", id).Delete(&models.VariantAssignment{}).Error; err != nil {
			return err
		}

		// Delete test results
		if err := tx.Where("test_id = ?", id).Delete(&models.TestResult{}).Error; err != nil {
			return err
		}

		// Delete variants
		if err := tx.Where("test_id = ?", id).Delete(&models.TestVariant{}).Error; err != nil {
			return err
		}

		// Delete test
		return tx.Where("id = ?", id).Delete(&models.ABTest{}).Error
	})
}

// StartTest starts an A/B test
func (s *ABTestService) StartTest(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var test models.ABTest
		if err := tx.Preload("Variants").Where("id = ?", id).First(&test).Error; err != nil {
			return err
		}

		if test.Status != models.TestStatusDraft {
			return errors.New("only draft tests can be started")
		}

		if len(test.Variants) < 2 {
			return errors.New("test must have at least 2 variants")
		}

		// Validate split percentages add up to 100
		totalPercent := 0
		for _, variant := range test.Variants {
			totalPercent += variant.SplitPercentage
		}
		if totalPercent != 100 {
			return fmt.Errorf("variant split percentages must add up to 100, got %d", totalPercent)
		}

		// Create result entries for each variant
		for _, variant := range test.Variants {
			result := models.TestResult{
				TestID:    test.ID,
				VariantID: variant.ID,
			}
			if err := tx.Create(&result).Error; err != nil {
				return err
			}
		}

		now := time.Now()
		return tx.Model(&models.ABTest{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"status":     models.TestStatusRunning,
				"started_at": now,
			}).Error
	})
}

// StopTest stops an A/B test
func (s *ABTestService) StopTest(ctx context.Context, id string) error {
	now := time.Now()
	return s.db.WithContext(ctx).
		Model(&models.ABTest{}).
		Where("id = ? AND status = ?", id, models.TestStatusRunning).
		Updates(map[string]interface{}{
			"status":   models.TestStatusStopped,
			"ended_at": now,
		}).Error
}

// AssignVariant assigns a contact to a variant
func (s *ABTestService) AssignVariant(ctx context.Context, testID string, contactID string) (*models.TestVariant, error) {
	// Check if already assigned
	var existing models.VariantAssignment
	err := s.db.WithContext(ctx).
		Where("test_id = ? AND contact_id = ?", testID, contactID).
		First(&existing).Error

	if err == nil {
		// Already assigned, return the variant
		var variant models.TestVariant
		if err := s.db.WithContext(ctx).Where("id = ?", existing.VariantID).First(&variant).Error; err != nil {
			return nil, err
		}
		return &variant, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Get test and variants
	var test models.ABTest
	if err := s.db.WithContext(ctx).
		Preload("Variants").
		Where("id = ?", testID).
		First(&test).Error; err != nil {
		return nil, err
	}

	if test.Status != models.TestStatusRunning {
		return nil, errors.New("test is not running")
	}

	// Select variant based on split percentages
	variant := s.selectVariant(test.Variants)

	// Create assignment
	assignment := models.VariantAssignment{
		TestID:    testID,
		ContactID: contactID,
		VariantID: variant.ID,
	}

	if err := s.db.WithContext(ctx).Create(&assignment).Error; err != nil {
		return nil, err
	}

	return &variant, nil
}

// selectVariant randomly selects a variant based on split percentages
func (s *ABTestService) selectVariant(variants []models.TestVariant) models.TestVariant {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(100) + 1 // 1-100

	cumulative := 0
	for _, variant := range variants {
		cumulative += variant.SplitPercentage
		if r <= cumulative {
			return variant
		}
	}

	// Fallback to first variant (should never reach here)
	return variants[0]
}

// RecordSend records that an email was sent for a variant
func (s *ABTestService) RecordSend(ctx context.Context, testID string, variantID string) error {
	return s.updateResult(ctx, variantID, map[string]interface{}{
		"sends": gorm.Expr("sends + ?", 1),
	})
}

// RecordOpen records an email open for a variant
func (s *ABTestService) RecordOpen(ctx context.Context, testID string, contactID string) error {
	variantID, err := s.getAssignedVariant(ctx, testID, contactID)
	if err != nil {
		return err
	}

	return s.updateResult(ctx, variantID, map[string]interface{}{
		"opens": gorm.Expr("opens + ?", 1),
	})
}

// RecordClick records a click for a variant
func (s *ABTestService) RecordClick(ctx context.Context, testID string, contactID string) error {
	variantID, err := s.getAssignedVariant(ctx, testID, contactID)
	if err != nil {
		return err
	}

	return s.updateResult(ctx, variantID, map[string]interface{}{
		"clicks": gorm.Expr("clicks + ?", 1),
	})
}

// RecordConversion records a conversion for a variant
func (s *ABTestService) RecordConversion(ctx context.Context, testID string, contactID string, value float64) error {
	variantID, err := s.getAssignedVariant(ctx, testID, contactID)
	if err != nil {
		return err
	}

	return s.updateResult(ctx, variantID, map[string]interface{}{
		"conversions": gorm.Expr("conversions + ?", 1),
		"revenue":     gorm.Expr("revenue + ?", value),
	})
}

// getAssignedVariant gets the variant assigned to a contact
func (s *ABTestService) getAssignedVariant(ctx context.Context, testID string, contactID string) (string, error) {
	var assignment models.VariantAssignment
	err := s.db.WithContext(ctx).
		Where("test_id = ? AND contact_id = ?", testID, contactID).
		First(&assignment).Error

	if err != nil {
		return "", err
	}

	return assignment.VariantID, nil
}

// updateResult updates result metrics and recalculates rates
func (s *ABTestService) updateResult(ctx context.Context, variantID string, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update counts
		if err := tx.Model(&models.TestResult{}).
			Where("variant_id = ?", variantID).
			Updates(updates).Error; err != nil {
			return err
		}

		// Recalculate rates
		var result models.TestResult
		if err := tx.Where("variant_id = ?", variantID).First(&result).Error; err != nil {
			return err
		}

		result.CalculateRates()

		return tx.Model(&models.TestResult{}).
			Where("variant_id = ?", variantID).
			Updates(map[string]interface{}{
				"open_rate":       result.OpenRate,
				"click_rate":      result.ClickRate,
				"conversion_rate": result.ConversionRate,
			}).Error
	})
}

// GetResults retrieves results for a test
func (s *ABTestService) GetResults(ctx context.Context, testID string) ([]models.TestResult, error) {
	var results []models.TestResult
	err := s.db.WithContext(ctx).
		Preload("Variant").
		Where("test_id = ?", testID).
		Find(&results).Error

	return results, err
}

// AnalyzeTest performs statistical analysis on test results
func (s *ABTestService) AnalyzeTest(ctx context.Context, testID string) (*TestAnalysis, error) {
	var test models.ABTest
	if err := s.db.WithContext(ctx).
		Preload("Variants.Result").
		Where("id = ?", testID).
		First(&test).Error; err != nil {
		return nil, err
	}

	results, err := s.GetResults(ctx, testID)
	if err != nil {
		return nil, err
	}

	if len(results) < 2 {
		return nil, errors.New("need at least 2 variants to analyze")
	}

	analysis := &TestAnalysis{
		TestID:          testID,
		IsSignificant:   false,
		WinnerVariantID: nil,
		Comparisons:     []VariantComparison{},
	}

	// Check if minimum sample size is met
	allMeetMinimum := true
	for _, result := range results {
		if result.Sends < test.MinimumSampleSize {
			allMeetMinimum = false
			break
		}
	}

	if !allMeetMinimum {
		analysis.Message = "Minimum sample size not yet reached for all variants"
		return analysis, nil
	}

	// Find best performing variant
	var bestVariant *models.TestResult
	var secondBestVariant *models.TestResult

	for i := range results {
		rate := results[i].GetRate(test.WinnerMetric)
		if bestVariant == nil || rate > bestVariant.GetRate(test.WinnerMetric) {
			secondBestVariant = bestVariant
			bestVariant = &results[i]
		} else if secondBestVariant == nil || rate > secondBestVariant.GetRate(test.WinnerMetric) {
			secondBestVariant = &results[i]
		}
	}

	// Calculate statistical significance between best and second best
	if bestVariant != nil && secondBestVariant != nil {
		pValue := s.calculatePValue(bestVariant, secondBestVariant, test.WinnerMetric)
		confidence := 1.0 - pValue

		comparison := VariantComparison{
			VariantAID:   bestVariant.VariantID,
			VariantBID:   secondBestVariant.VariantID,
			PValue:       pValue,
			Confidence:   confidence,
			IsSignificant: confidence >= test.ConfidenceThreshold,
		}

		analysis.Comparisons = append(analysis.Comparisons, comparison)

		if comparison.IsSignificant {
			analysis.IsSignificant = true
			analysis.WinnerVariantID = &bestVariant.VariantID
			analysis.Confidence = confidence
			analysis.Message = fmt.Sprintf("Variant has %.1f%% confidence of being better", confidence*100)
		} else {
			analysis.Message = fmt.Sprintf("Not yet significant (%.1f%% confidence, need %.1f%%)",
				confidence*100, test.ConfidenceThreshold*100)
		}
	}

	return analysis, nil
}

// calculatePValue performs chi-squared test for statistical significance
func (s *ABTestService) calculatePValue(a, b *models.TestResult, metric models.WinnerMetric) float64 {
	// Simplified chi-squared test for proportions
	// In production, use a proper statistics library

	var successA, totalA, successB, totalB float64

	switch metric {
	case models.MetricOpenRate:
		successA, totalA = float64(a.Opens), float64(a.Sends)
		successB, totalB = float64(b.Opens), float64(b.Sends)
	case models.MetricClickRate:
		successA, totalA = float64(a.Clicks), float64(a.Sends)
		successB, totalB = float64(b.Clicks), float64(b.Sends)
	case models.MetricConversionRate:
		successA, totalA = float64(a.Conversions), float64(a.Sends)
		successB, totalB = float64(b.Conversions), float64(b.Sends)
	default:
		return 1.0 // Not significant
	}

	if totalA == 0 || totalB == 0 {
		return 1.0
	}

	// Calculate pooled proportion
	pooledP := (successA + successB) / (totalA + totalB)
	expectedA := totalA * pooledP
	expectedB := totalB * pooledP

	// Chi-squared statistic
	chiSquared := math.Pow(successA-expectedA, 2)/expectedA +
		math.Pow(totalA-successA-(totalA-expectedA), 2)/(totalA-expectedA) +
		math.Pow(successB-expectedB, 2)/expectedB +
		math.Pow(totalB-successB-(totalB-expectedB), 2)/(totalB-expectedB)

	// Simplified p-value calculation (degrees of freedom = 1)
	// For chi-squared = 3.84, p = 0.05 (95% confidence)
	// For chi-squared = 6.63, p = 0.01 (99% confidence)

	if chiSquared >= 6.63 {
		return 0.01
	} else if chiSquared >= 3.84 {
		return 0.05
	} else {
		return 0.10
	}
}

// DeclareWinner manually declares a winner
func (s *ABTestService) DeclareWinner(ctx context.Context, testID string, variantID string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		return tx.Model(&models.ABTest{}).
			Where("id = ?", testID).
			Updates(map[string]interface{}{
				"winner_variant_id": variantID,
				"status":            models.TestStatusCompleted,
				"ended_at":          now,
			}).Error
	})
}

// AutoDeclareWinners checks all running tests and auto-declares winners if thresholds are met
func (s *ABTestService) AutoDeclareWinners(ctx context.Context) error {
	var tests []models.ABTest
	if err := s.db.WithContext(ctx).
		Where("status = ? AND auto_declare_winner = ?", models.TestStatusRunning, true).
		Find(&tests).Error; err != nil {
		return err
	}

	for _, test := range tests {
		analysis, err := s.AnalyzeTest(ctx, test.ID)
		if err != nil {
			continue
		}

		if analysis.IsSignificant && analysis.WinnerVariantID != nil {
			if err := s.DeclareWinner(ctx, test.ID, *analysis.WinnerVariantID); err != nil {
				continue
			}
		}
	}

	return nil
}

// TestAnalysis represents the statistical analysis of a test
type TestAnalysis struct {
	TestID          string              `json:"testId"`
	IsSignificant   bool                `json:"isSignificant"`
	WinnerVariantID *string             `json:"winnerVariantId"`
	Confidence      float64             `json:"confidence"`
	Message         string              `json:"message"`
	Comparisons     []VariantComparison `json:"comparisons"`
}

// VariantComparison represents a statistical comparison between two variants
type VariantComparison struct {
	VariantAID    string  `json:"variantAId"`
	VariantBID    string  `json:"variantBId"`
	PValue        float64 `json:"pValue"`
	Confidence    float64 `json:"confidence"`
	IsSignificant bool    `json:"isSignificant"`
}
