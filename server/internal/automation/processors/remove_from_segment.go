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

// RemoveFromSegmentProcessor handles REMOVE_FROM_SEGMENT nodes
type RemoveFromSegmentProcessor struct {
	db             *gorm.DB
	segmentService *services.SegmentService
}

// RemoveFromSegmentNodeData represents the data structure for REMOVE_FROM_SEGMENT nodes
type RemoveFromSegmentNodeData struct {
	SegmentID string `json:"segmentId"` // Segment to remove contact from
}

// NewRemoveFromSegmentProcessor creates a new REMOVE_FROM_SEGMENT node processor
func NewRemoveFromSegmentProcessor(db *gorm.DB) *RemoveFromSegmentProcessor {
	return &RemoveFromSegmentProcessor{
		db:             db,
		segmentService: services.NewSegmentService(db),
	}
}

// Type returns the node type this processor handles
func (p *RemoveFromSegmentProcessor) Type() models.NodeType {
	return "REMOVE_FROM_SEGMENT"
}

// Validate checks if the REMOVE_FROM_SEGMENT node data is valid
func (p *RemoveFromSegmentProcessor) Validate(node *models.AutomationNode) error {
	var data RemoveFromSegmentNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return fmt.Errorf("invalid remove from segment node data: %w", err)
	}

	if data.SegmentID == "" {
		return fmt.Errorf("segmentId is required")
	}

	return nil
}

// Process executes the REMOVE_FROM_SEGMENT node logic
func (p *RemoveFromSegmentProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	var data RemoveFromSegmentNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse remove from segment node data: %w", err)
	}

	// Remove contact from segment
	if err := p.segmentService.RemoveContactFromSegment(context.Background(), data.SegmentID, ctx.ContactID); err != nil {
		return nil, fmt.Errorf("failed to remove contact from segment: %w", err)
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
		Message:     fmt.Sprintf("Removed contact from segment %s", data.SegmentID),
		Data: map[string]interface{}{
			"segmentId": data.SegmentID,
		},
	}, nil
}
