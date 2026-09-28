package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"kori/internal/automation"
	"kori/internal/models"
	"kori/internal/services"

	"gorm.io/gorm"
)

// ScoreChangeProcessor handles SCORE_CHANGE nodes
type ScoreChangeProcessor struct {
	db             *gorm.DB
	scoringService *services.LeadScoringService
}

// ScoreChangeNodeData represents the data structure for SCORE_CHANGE nodes
type ScoreChangeNodeData struct {
	Points       int    `json:"points"`       // Points to add (can be negative)
	ActivityType string `json:"activityType"` // Activity type for audit log
	Description  string `json:"description"`  // Human-readable description
}

// NewScoreChangeProcessor creates a new SCORE_CHANGE node processor
func NewScoreChangeProcessor(db *gorm.DB) *ScoreChangeProcessor {
	return &ScoreChangeProcessor{
		db:             db,
		scoringService: services.NewLeadScoringService(db),
	}
}

// Type returns the node type this processor handles
func (p *ScoreChangeProcessor) Type() models.NodeType {
	return "SCORE_CHANGE"
}

// Validate checks if the SCORE_CHANGE node data is valid
func (p *ScoreChangeProcessor) Validate(node *models.AutomationNode) error {
	if node.Type != p.Type() {
		return fmt.Errorf("invalid node type for ScoreChangeProcessor")
	}

	var data ScoreChangeNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return fmt.Errorf("invalid score change node data: %w", err)
	}

	if data.Points == 0 {
		return fmt.Errorf("points cannot be zero")
	}

	if data.ActivityType == "" {
		return fmt.Errorf("activityType is required")
	}

	return nil
}

// Process executes the SCORE_CHANGE node logic
func (p *ScoreChangeProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	var data ScoreChangeNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse score change node data: %w", err)
	}

	// Prepare metadata
	metadata := map[string]interface{}{
		"automation_id": ctx.AutomationID,
	}

	// Add points to contact's score
	if err := p.scoringService.AddPoints(
		context.Background(),
		ctx.ContactID,
		data.ActivityType,
		data.Points,
		data.Description,
		metadata,
	); err != nil {
		return nil, fmt.Errorf("failed to update score: %w", err)
	}

	// Get updated score
	score, err := p.scoringService.GetContactScore(context.Background(), ctx.ContactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated score: %w", err)
	}

	// Get next nodes
	var edges []models.AutomationNodeEdge
	if err := p.db.Where("automation_id = ? AND source_id = ?", ctx.AutomationID, node.ID).Find(&edges).Error; err != nil {
		return nil, fmt.Errorf("failed to load edges: %w", err)
	}

	nextNodeIDs := make([]string, len(edges))
	for i, edge := range edges {
		nextNodeIDs[i] = edge.TargetID
	}

	return &automation.ProcessResult{
		NextNodeIDs: nextNodeIDs,
		Message:     fmt.Sprintf("Updated score by %d points. New score: %d (Grade: %s)", data.Points, score.Score, score.Grade),
		Data: map[string]interface{}{
			"points":    data.Points,
			"newScore":  score.Score,
			"newGrade":  score.Grade,
			"activityType": data.ActivityType,
		},
	}, nil
}
