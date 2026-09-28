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

// GoalTrackProcessor handles GOAL_TRACK nodes for recording goal conversions
type GoalTrackProcessor struct {
	db          *gorm.DB
	goalService *services.GoalTrackingService
}

// GoalTrackNodeData represents the data structure for GOAL_TRACK nodes
type GoalTrackNodeData struct {
	GoalID string  `json:"goalId"` // Goal ID to track
	Value  float64 `json:"value"`  // Optional value (for revenue goals)
}

// NewGoalTrackProcessor creates a new GOAL_TRACK node processor
func NewGoalTrackProcessor(db *gorm.DB) *GoalTrackProcessor {
	return &GoalTrackProcessor{
		db:          db,
		goalService: services.NewGoalTrackingService(db),
	}
}

// Type returns the node type this processor handles
func (p *GoalTrackProcessor) Type() models.NodeType {
	return "GOAL_TRACK"
}

// Validate checks if the GOAL_TRACK node data is valid
func (p *GoalTrackProcessor) Validate(node *models.AutomationNode) error {
	if node.Type != p.Type() {
		return fmt.Errorf("invalid node type for GoalTrackProcessor")
	}

	var data GoalTrackNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return fmt.Errorf("invalid goal track node data: %w", err)
	}

	if data.GoalID == "" {
		return fmt.Errorf("goalId is required")
	}

	return nil
}

// Process executes the GOAL_TRACK node logic
func (p *GoalTrackProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	var data GoalTrackNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse goal track node data: %w", err)
	}

	// Get goal details
	goal, err := p.goalService.GetGoal(context.Background(), data.GoalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}

	// Record conversion
	conversion := &models.GoalConversion{
		GoalID:          data.GoalID,
		AutomationID:    ctx.AutomationID,
		ContactID:       ctx.ContactID,
		ConversionValue: data.Value,
	}

	// Add execution ID if available from context
	if executionID, ok := ctx.Variables["execution_id"].(string); ok {
		conversion.ExecutionID = &executionID
	}

	if err := p.goalService.RecordConversion(context.Background(), conversion); err != nil {
		return nil, fmt.Errorf("failed to record conversion: %w", err)
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
		Message:     fmt.Sprintf("Recorded conversion for goal '%s' with value %.2f", goal.Name, data.Value),
		Data: map[string]interface{}{
			"goalId":    data.GoalID,
			"goalName":  goal.Name,
			"goalType":  goal.GoalType,
			"value":     data.Value,
		},
	}, nil
}
