package processors

import (
	"encoding/json"
	"fmt"
	"kori/internal/automation"
	"kori/internal/models"
	"kori/internal/workflowconfig"
	"time"

	"gorm.io/gorm"
)

// WaitProcessor handles WAIT nodes
type WaitProcessor struct {
	db *gorm.DB
}

// WaitNodeData represents the data structure for WAIT nodes
type WaitNodeData = workflowconfig.Wait

// NewWaitProcessor creates a new WAIT node processor
func NewWaitProcessor(db *gorm.DB) *WaitProcessor {
	return &WaitProcessor{db: db}
}

// Type returns the node type this processor handles
func (p *WaitProcessor) Type() models.NodeType {
	return models.NodeTypeWait
}

// Validate checks if the WAIT node data is valid
func (p *WaitProcessor) Validate(node *models.AutomationNode) error {
	if node.Type != models.NodeTypeWait {
		return fmt.Errorf("invalid node type for WaitProcessor")
	}

	return workflowconfig.Validate("WAIT", node.Data)
}

// Process executes the WAIT node logic
func (p *WaitProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	if err := p.Validate(node); err != nil {
		return nil, err
	}
	var data WaitNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse wait node data: %w", err)
	}

	// Parse duration
	duration, err := workflowconfig.WaitDuration(data, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to parse duration: %w", err)
	}

	// Get next nodes
	var edges []models.AutomationNodeEdge
	if err := p.db.Where("automation_id = ? AND source_id = ?", ctx.AutomationID, node.ID).Find(&edges).Error; err != nil {
		return nil, fmt.Errorf("failed to load edges: %w", err)
	}

	if len(edges) != 1 {
		return nil, fmt.Errorf("wait step needs one continuation")
	}

	nextNodeIDs := make([]string, len(edges))
	for i, edge := range edges {
		nextNodeIDs[i] = edge.TargetID
	}

	// Return result with wait duration
	return &automation.ProcessResult{
		NextNodeIDs: nextNodeIDs,
		Wait:        &duration,
		Message:     fmt.Sprintf("Waiting for %s before continuing", duration),
		Data: map[string]interface{}{
			"duration": data.Duration,
			"resumeAt": time.Now().Add(duration).Format(time.RFC3339),
		},
	}, nil
}
