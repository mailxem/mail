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

// ScoreThresholdProcessor handles SCORE_THRESHOLD nodes
type ScoreThresholdProcessor struct {
	db             *gorm.DB
	scoringService *services.LeadScoringService
}

// ScoreThresholdNodeData represents the data structure for SCORE_THRESHOLD nodes
type ScoreThresholdNodeData struct {
	Threshold int               `json:"threshold"` // Score threshold to check
	Operator  string            `json:"operator"`  // ">=", ">", "<=", "<", "=="
	Branches  map[string]string `json:"branches"`  // "true" and "false" branch node IDs
}

// NewScoreThresholdProcessor creates a new SCORE_THRESHOLD node processor
func NewScoreThresholdProcessor(db *gorm.DB) *ScoreThresholdProcessor {
	return &ScoreThresholdProcessor{
		db:             db,
		scoringService: services.NewLeadScoringService(db),
	}
}

// Type returns the node type this processor handles
func (p *ScoreThresholdProcessor) Type() models.NodeType {
	return "SCORE_THRESHOLD"
}

// Validate checks if the SCORE_THRESHOLD node data is valid
func (p *ScoreThresholdProcessor) Validate(node *models.AutomationNode) error {
	if node.Type != p.Type() {
		return fmt.Errorf("invalid node type for ScoreThresholdProcessor")
	}

	var data ScoreThresholdNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return fmt.Errorf("invalid score threshold node data: %w", err)
	}

	// Validate operator
	validOperators := map[string]bool{
		">=": true,
		">":  true,
		"<=": true,
		"<":  true,
		"==": true,
	}
	if !validOperators[data.Operator] {
		return fmt.Errorf("invalid operator: %s. Must be one of: >=, >, <=, <, ==", data.Operator)
	}

	if data.Branches == nil || len(data.Branches) == 0 {
		return fmt.Errorf("branches are required")
	}

	if data.Branches["true"] == "" || data.Branches["false"] == "" {
		return fmt.Errorf("both 'true' and 'false' branches are required")
	}

	return nil
}

// Process executes the SCORE_THRESHOLD node logic
func (p *ScoreThresholdProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	var data ScoreThresholdNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse score threshold node data: %w", err)
	}

	// Get contact's current score
	score, err := p.scoringService.GetContactScore(context.Background(), ctx.ContactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact score: %w", err)
	}

	// Evaluate threshold
	var passes bool
	switch data.Operator {
	case ">=":
		passes = score.Score >= data.Threshold
	case ">":
		passes = score.Score > data.Threshold
	case "<=":
		passes = score.Score <= data.Threshold
	case "<":
		passes = score.Score < data.Threshold
	case "==":
		passes = score.Score == data.Threshold
	default:
		return nil, fmt.Errorf("invalid operator: %s", data.Operator)
	}

	// Route based on result
	var nextNodeID string
	if passes {
		nextNodeID = data.Branches["true"]
	} else {
		nextNodeID = data.Branches["false"]
	}

	return &automation.ProcessResult{
		NextNodeIDs: []string{nextNodeID},
		Message: fmt.Sprintf("Score check: %d %s %d = %v. Current grade: %s",
			score.Score, data.Operator, data.Threshold, passes, score.Grade),
		Data: map[string]interface{}{
			"score":     score.Score,
			"grade":     score.Grade,
			"threshold": data.Threshold,
			"operator":  data.Operator,
			"passes":    passes,
		},
	}, nil
}
