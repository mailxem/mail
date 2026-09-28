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

// SegmentFilterProcessor handles SEGMENT_FILTER nodes
type SegmentFilterProcessor struct {
	db             *gorm.DB
	segmentService *services.SegmentService
}

// SegmentFilterNodeData represents the data structure for SEGMENT_FILTER nodes
type SegmentFilterNodeData struct {
	SegmentID string            `json:"segmentId"` // Segment to check
	Branches  map[string]string `json:"branches"`  // "true" and "false" branch node IDs
}

// NewSegmentFilterProcessor creates a new SEGMENT_FILTER node processor
func NewSegmentFilterProcessor(db *gorm.DB) *SegmentFilterProcessor {
	return &SegmentFilterProcessor{
		db:             db,
		segmentService: services.NewSegmentService(db),
	}
}

// Type returns the node type this processor handles
func (p *SegmentFilterProcessor) Type() models.NodeType {
	return "SEGMENT_FILTER"
}

// Validate checks if the SEGMENT_FILTER node data is valid
func (p *SegmentFilterProcessor) Validate(node *models.AutomationNode) error {
	if node.Type != p.Type() {
		return fmt.Errorf("invalid node type for SegmentFilterProcessor")
	}

	var data SegmentFilterNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return fmt.Errorf("invalid segment filter node data: %w", err)
	}

	if data.SegmentID == "" {
		return fmt.Errorf("segmentId is required")
	}

	if data.Branches == nil || len(data.Branches) == 0 {
		return fmt.Errorf("branches are required")
	}

	if data.Branches["true"] == "" || data.Branches["false"] == "" {
		return fmt.Errorf("both 'true' and 'false' branches are required")
	}

	return nil
}

// Process executes the SEGMENT_FILTER node logic
func (p *SegmentFilterProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	var data SegmentFilterNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse segment filter node data: %w", err)
	}

	// Check if contact is in segment
	isInSegment, err := p.segmentService.IsContactInSegment(context.Background(), data.SegmentID, ctx.ContactID)
	if err != nil {
		return nil, fmt.Errorf("failed to check segment membership: %w", err)
	}

	// Route based on segment membership
	var nextNodeID string
	if isInSegment {
		nextNodeID = data.Branches["true"]
	} else {
		nextNodeID = data.Branches["false"]
	}

	return &automation.ProcessResult{
		NextNodeIDs: []string{nextNodeID},
		Message:     fmt.Sprintf("Contact is %s segment %s", map[bool]string{true: "in", false: "not in"}[isInSegment], data.SegmentID),
		Data: map[string]interface{}{
			"segmentId":   data.SegmentID,
			"isInSegment": isInSegment,
		},
	}, nil
}
