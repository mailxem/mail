package processors

import (
	"fmt"
	"kori/internal/automation"
	"kori/internal/models"

	"gorm.io/gorm"
)

// StartProcessor handles START nodes
type StartProcessor struct {
	db *gorm.DB
}

// NewStartProcessor creates a new START node processor
func NewStartProcessor(db *gorm.DB) *StartProcessor {
	return &StartProcessor{db: db}
}

// Type returns the node type this processor handles
func (p *StartProcessor) Type() models.NodeType {
	return models.NodeTypeStart
}

// Validate checks if the START node data is valid
func (p *StartProcessor) Validate(node *models.AutomationNode) error {
	if node.Type != models.NodeTypeStart {
		return fmt.Errorf("invalid node type for StartProcessor")
	}
	return nil
}

// Process executes the START node logic
func (p *StartProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	// START node simply passes execution to the next node
	// Load edges to find next node
	var edges []models.AutomationNodeEdge
	if err := p.db.Where("automation_id = ? AND source_id = ?", ctx.AutomationID, node.ID).Find(&edges).Error; err != nil {
		return nil, fmt.Errorf("failed to load edges: %w", err)
	}

	if len(edges) == 0 {
		return &automation.ProcessResult{
			Complete: true,
			Message:  "START node has no outgoing edges, workflow complete",
		}, nil
	}

	// Get next node IDs
	nextNodeIDs := make([]string, len(edges))
	for i, edge := range edges {
		nextNodeIDs[i] = edge.TargetID
	}

	ctx.RefreshContactVariables()

	return &automation.ProcessResult{
		NextNodeIDs: nextNodeIDs,
		Message:     "Started automation workflow",
	}, nil
}
