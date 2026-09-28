package services

import (
	"context"
	"encoding/json"
	"fmt"
	"kori/internal/models"
	"kori/internal/utils/logger"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

var segmentLog = logger.New("SEGMENT_SERVICE")

type SegmentService struct {
	base BaseService[models.Segment]
	db   *gorm.DB
}

func NewSegmentService(db *gorm.DB) *SegmentService {
	return &SegmentService{
		base: NewBaseService(db, models.Segment{}),
		db:   db,
	}
}

// Create creates a new segment
func (s *SegmentService) Create(ctx context.Context, segment *models.Segment, includes ...string) error {
	// Validate rules
	if err := s.ValidateRules(segment); err != nil {
		return fmt.Errorf("invalid segment rules: %w", err)
	}

	if err := s.base.Create(ctx, segment, includes...); err != nil {
		return err
	}

	// If dynamic, calculate initial contact count
	if segment.IsDynamic {
		if err := s.RefreshSegment(ctx, segment.ID); err != nil {
			segmentLog.Warn("Failed to refresh new segment %s: %v", segment.ID, err)
		}
	}

	return nil
}

// Get retrieves a segment by ID
func (s *SegmentService) Get(ctx context.Context, id string, includes ...string) (*models.Segment, error) {
	return s.base.Get(ctx, id, includes...)
}

// List retrieves segments with filters
func (s *SegmentService) List(ctx context.Context, page, limit int, filters map[string]interface{}, excludeFields map[string]bool, sortFields []string, order string, includes ...string) ([]models.Segment, int64, error) {
	return s.base.List(ctx, page, limit, filters, excludeFields, sortFields, order, includes...)
}

// Update updates a segment
func (s *SegmentService) Update(ctx context.Context, id string, segment *models.Segment, includes ...string) error {
	// Validate rules if they changed
	if err := s.ValidateRules(segment); err != nil {
		return fmt.Errorf("invalid segment rules: %w", err)
	}

	if err := s.base.Update(ctx, id, segment, includes...); err != nil {
		return err
	}

	// If dynamic, refresh membership
	if segment.IsDynamic {
		if err := s.RefreshSegment(ctx, id); err != nil {
			segmentLog.Warn("Failed to refresh segment %s after update: %v", id, err)
		}
	}

	return nil
}

// Delete deletes a segment
func (s *SegmentService) Delete(ctx context.Context, id string) error {
	// Delete segment contacts first
	if err := s.db.WithContext(ctx).Where("segment_id = ?", id).Delete(&models.SegmentContact{}).Error; err != nil {
		return fmt.Errorf("failed to delete segment contacts: %w", err)
	}

	return s.base.Delete(ctx, id)
}

// ValidateRules validates segment rules
func (s *SegmentService) ValidateRules(segment *models.Segment) error {
	var rules models.SegmentRules
	if err := json.Unmarshal(segment.Rules, &rules); err != nil {
		return fmt.Errorf("failed to parse rules: %w", err)
	}

	if len(rules.Rules) == 0 {
		return fmt.Errorf("segment must have at least one rule")
	}

	// Validate each rule
	for i, rule := range rules.Rules {
		if rule.Field == "" {
			return fmt.Errorf("rule %d: field is required", i)
		}
		if rule.Operator == "" {
			return fmt.Errorf("rule %d: operator is required", i)
		}
		// Value can be nil for operators like is_set, is_not_set, is_empty, is_not_empty
		if !s.isNullableOperator(rule.Operator) && rule.Value == nil {
			return fmt.Errorf("rule %d: value is required for operator %s", i, rule.Operator)
		}
	}

	// Validate match type
	if rules.MatchType != models.MatchTypeAll && rules.MatchType != models.MatchTypeAny {
		return fmt.Errorf("match type must be ALL or ANY")
	}

	return nil
}

// isNullableOperator checks if an operator doesn't require a value
func (s *SegmentService) isNullableOperator(op models.RuleOperator) bool {
	nullableOps := []models.RuleOperator{
		models.OpIsSet,
		models.OpIsNotSet,
		models.OpIsEmpty,
		models.OpIsNotEmpty,
		models.OpIsTrue,
		models.OpIsFalse,
	}
	for _, nullableOp := range nullableOps {
		if op == nullableOp {
			return true
		}
	}
	return false
}

// RefreshSegment recalculates segment membership and updates contact count
func (s *SegmentService) RefreshSegment(ctx context.Context, segmentID string) error {
	segment, err := s.Get(ctx, segmentID)
	if err != nil {
		return fmt.Errorf("failed to load segment: %w", err)
	}

	if !segment.IsDynamic {
		return fmt.Errorf("cannot refresh static segment")
	}

	// Get matching contacts
	contacts, err := s.GetMatchingContacts(ctx, segment)
	if err != nil {
		return fmt.Errorf("failed to get matching contacts: %w", err)
	}

	// Update contact count
	now := time.Now()
	if err := s.db.WithContext(ctx).Model(&models.Segment{}).
		Where("id = ?", segmentID).
		Updates(map[string]interface{}{
			"contact_count": len(contacts),
			"last_refresh":  now,
		}).Error; err != nil {
		return fmt.Errorf("failed to update contact count: %w", err)
	}

	// Clear existing memberships
	if err := s.db.WithContext(ctx).Where("segment_id = ?", segmentID).Delete(&models.SegmentContact{}).Error; err != nil {
		return fmt.Errorf("failed to clear segment contacts: %w", err)
	}

	// Add new memberships
	if len(contacts) > 0 {
		segmentContacts := make([]models.SegmentContact, len(contacts))
		for i, contact := range contacts {
			segmentContacts[i] = models.SegmentContact{
				SegmentID: segmentID,
				ContactID: contact.ID,
				AddedAt:   now,
			}
		}

		if err := s.db.WithContext(ctx).Create(&segmentContacts).Error; err != nil {
			return fmt.Errorf("failed to create segment contacts: %w", err)
		}
	}

	segmentLog.Success("Refreshed segment %s: %d contacts", segmentID, len(contacts))
	return nil
}

// GetMatchingContacts returns contacts that match the segment rules
func (s *SegmentService) GetMatchingContacts(ctx context.Context, segment *models.Segment) ([]models.Contact, error) {
	var rules models.SegmentRules
	if err := json.Unmarshal(segment.Rules, &rules); err != nil {
		return nil, fmt.Errorf("failed to parse rules: %w", err)
	}

	// Build query
	query := s.db.WithContext(ctx).Model(&models.Contact{}).Where("team_id = ? AND is_deleted = ?", segment.TeamID, false)

	// Apply rules
	for i, rule := range rules.Rules {
		condition, args := s.buildRuleCondition(rule, i)
		if rules.MatchType == models.MatchTypeAll {
			query = query.Where(condition, args...)
		} else {
			query = query.Or(condition, args...)
		}
	}

	var contacts []models.Contact
	if err := query.Find(&contacts).Error; err != nil {
		return nil, err
	}

	return contacts, nil
}

// buildRuleCondition builds a SQL condition for a single rule
func (s *SegmentService) buildRuleCondition(rule models.SegmentRule, index int) (string, []interface{}) {
	field := rule.Field
	op := rule.Operator
	value := rule.Value

	// Handle special field mappings
	if strings.HasPrefix(field, "contact_") {
		field = strings.TrimPrefix(field, "contact_")
	}

	switch op {
	case models.OpEquals:
		return fmt.Sprintf("%s = ?", field), []interface{}{value}

	case models.OpNotEquals:
		return fmt.Sprintf("%s != ?", field), []interface{}{value}

	case models.OpContains:
		return fmt.Sprintf("%s ILIKE ?", field), []interface{}{"%" + fmt.Sprint(value) + "%"}

	case models.OpNotContains:
		return fmt.Sprintf("%s NOT ILIKE ?", field), []interface{}{"%" + fmt.Sprint(value) + "%"}

	case models.OpStartsWith:
		return fmt.Sprintf("%s ILIKE ?", field), []interface{}{fmt.Sprint(value) + "%"}

	case models.OpEndsWith:
		return fmt.Sprintf("%s ILIKE ?", field), []interface{}{"%" + fmt.Sprint(value)}

	case models.OpIsSet:
		return fmt.Sprintf("%s IS NOT NULL AND %s != ''", field, field), []interface{}{}

	case models.OpIsNotSet:
		return fmt.Sprintf("(%s IS NULL OR %s = '')", field, field), []interface{}{}

	case models.OpGreaterThan:
		numValue, _ := strconv.ParseFloat(fmt.Sprint(value), 64)
		return fmt.Sprintf("CAST(%s AS NUMERIC) > ?", field), []interface{}{numValue}

	case models.OpLessThan:
		numValue, _ := strconv.ParseFloat(fmt.Sprint(value), 64)
		return fmt.Sprintf("CAST(%s AS NUMERIC) < ?", field), []interface{}{numValue}

	case models.OpGreaterOrEqual:
		numValue, _ := strconv.ParseFloat(fmt.Sprint(value), 64)
		return fmt.Sprintf("CAST(%s AS NUMERIC) >= ?", field), []interface{}{numValue}

	case models.OpLessOrEqual:
		numValue, _ := strconv.ParseFloat(fmt.Sprint(value), 64)
		return fmt.Sprintf("CAST(%s AS NUMERIC) <= ?", field), []interface{}{numValue}

	case models.OpIncludes:
		// For JSONB array fields
		return fmt.Sprintf("%s @> ?::jsonb", field), []interface{}{fmt.Sprintf(`["%s"]`, value)}

	case models.OpNotIncludes:
		return fmt.Sprintf("NOT (%s @> ?::jsonb)", field), []interface{}{fmt.Sprintf(`["%s"]`, value)}

	case models.OpIsEmpty:
		return fmt.Sprintf("(%s IS NULL OR %s = '[]'::jsonb OR %s = '{}'::jsonb)", field, field, field), []interface{}{}

	case models.OpIsNotEmpty:
		return fmt.Sprintf("(%s IS NOT NULL AND %s != '[]'::jsonb AND %s != '{}'::jsonb)", field, field, field), []interface{}{}

	case models.OpBefore:
		return fmt.Sprintf("%s < ?", field), []interface{}{value}

	case models.OpAfter:
		return fmt.Sprintf("%s > ?", field), []interface{}{value}

	case models.OpWithinDays:
		days, _ := strconv.Atoi(fmt.Sprint(value))
		return fmt.Sprintf("%s > NOW() - INTERVAL '%d days'", field, days), []interface{}{}

	case models.OpOlderThan:
		days, _ := strconv.Atoi(fmt.Sprint(value))
		return fmt.Sprintf("%s < NOW() - INTERVAL '%d days'", field, days), []interface{}{}

	case models.OpIsTrue:
		return fmt.Sprintf("%s = true", field), []interface{}{}

	case models.OpIsFalse:
		return fmt.Sprintf("%s = false", field), []interface{}{}

	default:
		segmentLog.Warn("Unknown operator: %s", op)
		return "1 = 1", []interface{}{} // Always true fallback
	}
}

// GetSegmentContacts returns contacts in a segment (paginated)
func (s *SegmentService) GetSegmentContacts(ctx context.Context, segmentID string, page, limit int) ([]models.Contact, int64, error) {
	var contacts []models.Contact
	var total int64

	// Count total
	if err := s.db.WithContext(ctx).
		Model(&models.Contact{}).
		Joins("JOIN segment_contacts ON segment_contacts.contact_id = contacts.id").
		Where("segment_contacts.segment_id = ?", segmentID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated contacts
	offset := (page - 1) * limit
	if err := s.db.WithContext(ctx).
		Model(&models.Contact{}).
		Joins("JOIN segment_contacts ON segment_contacts.contact_id = contacts.id").
		Where("segment_contacts.segment_id = ?", segmentID).
		Offset(offset).
		Limit(limit).
		Find(&contacts).Error; err != nil {
		return nil, 0, err
	}

	return contacts, total, nil
}

// PreviewSegment returns a preview of contacts that would match the rules (without saving)
func (s *SegmentService) PreviewSegment(ctx context.Context, teamID string, rules models.SegmentRules, limit int) ([]models.Contact, int64, error) {
	// Create temporary segment for preview
	tempSegment := &models.Segment{
		TeamID:    teamID,
		Rules:     nil,
		MatchType: rules.MatchType,
		IsDynamic: true,
	}

	rulesJSON, err := json.Marshal(rules)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal rules: %w", err)
	}
	tempSegment.Rules = rulesJSON

	// Get matching contacts
	contacts, err := s.GetMatchingContacts(ctx, tempSegment)
	if err != nil {
		return nil, 0, err
	}

	total := int64(len(contacts))

	// Limit results for preview
	if limit > 0 && len(contacts) > limit {
		contacts = contacts[:limit]
	}

	return contacts, total, nil
}

// IsContactInSegment checks if a contact is in a segment
func (s *SegmentService) IsContactInSegment(ctx context.Context, segmentID, contactID string) (bool, error) {
	var count int64
	if err := s.db.WithContext(ctx).
		Model(&models.SegmentContact{}).
		Where("segment_id = ? AND contact_id = ?", segmentID, contactID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// AddContactToSegment adds a contact to a static segment
func (s *SegmentService) AddContactToSegment(ctx context.Context, segmentID, contactID string) error {
	// Check if segment is static
	segment, err := s.Get(ctx, segmentID)
	if err != nil {
		return err
	}

	if segment.IsDynamic {
		return fmt.Errorf("cannot manually add contacts to dynamic segment")
	}

	// Check if already exists
	exists, err := s.IsContactInSegment(ctx, segmentID, contactID)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already in segment
	}

	// Add to segment
	segmentContact := &models.SegmentContact{
		SegmentID: segmentID,
		ContactID: contactID,
		AddedAt:   time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(segmentContact).Error; err != nil {
		return err
	}

	// Update contact count
	if err := s.db.WithContext(ctx).Model(&models.Segment{}).
		Where("id = ?", segmentID).
		UpdateColumn("contact_count", gorm.Expr("contact_count + 1")).Error; err != nil {
		return err
	}

	return nil
}

// RemoveContactFromSegment removes a contact from a static segment
func (s *SegmentService) RemoveContactFromSegment(ctx context.Context, segmentID, contactID string) error {
	// Check if segment is static
	segment, err := s.Get(ctx, segmentID)
	if err != nil {
		return err
	}

	if segment.IsDynamic {
		return fmt.Errorf("cannot manually remove contacts from dynamic segment")
	}

	// Remove from segment
	if err := s.db.WithContext(ctx).
		Where("segment_id = ? AND contact_id = ?", segmentID, contactID).
		Delete(&models.SegmentContact{}).Error; err != nil {
		return err
	}

	// Update contact count
	if err := s.db.WithContext(ctx).Model(&models.Segment{}).
		Where("id = ?", segmentID).
		UpdateColumn("contact_count", gorm.Expr("GREATEST(contact_count - 1, 0)")).Error; err != nil {
		return err
	}

	return nil
}
