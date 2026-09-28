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

// AddToSegmentProcessor handles ADD_TO_SEGMENT nodes
type AddToSegmentProcessor struct {
	db             *gorm.DB
	segmentService *services.SegmentService
}

// AddToSegmentNodeData represents the data structure for ADD_TO_SEGMENT nodes
type AddToSegmentNodeData struct {
	SegmentID string `json:"segmentId"` // Segment to add contact to
}

// NewAddToSegmentProcessor creates a new ADD_TO_SEGMENT node processor
func NewAddToSegmentProcessor(db *gorm.DB) *AddToSegmentProcessor {
	return &AddToSegmentProcessor{
		db:             db,
		segmentService: services.NewSegmentService(db),
	}
}

// Type returns the node type this processor handles
func (p *AddToSegmentProcessor) Type() models.NodeType {
	return models.NodeTypeAddToList // Reuse existing type or create new one
}

// Validate checks if the ADD_TO_SEGMENT node data is valid
func (p *AddToSegmentProcessor) Validate(node *models.AutomationNode) error {
	var data AddToSegmentNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return fmt.Errorf("invalid add to segment node data: %w", err)
	}

	if data.SegmentID == "" {
		return fmt.Errorf("segmentId is required")
	}

	return nil
}

// Process executes the ADD_TO_SEGMENT node logic
func (p *AddToSegmentProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	var data AddToSegmentNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse add to segment node data: %w", err)
	}

	// Add contact to segment
	if err := p.segmentService.AddContactToSegment(context.Background(), data.SegmentID, ctx.ContactID); err != nil {
		return nil, fmt.Errorf("failed to add contact to segment: %w", err)
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
		Message:     fmt.Sprintf("Added contact to segment %s", data.SegmentID),
		Data: map[string]interface{}{
			"segmentId": data.SegmentID,
		},
	}, nil
}
