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

// ABSplitProcessor handles AB_SPLIT nodes for A/B testing
type ABSplitProcessor struct {
	db          *gorm.DB
	testService *services.ABTestService
}

// ABSplitNodeData represents the data structure for AB_SPLIT nodes
type ABSplitNodeData struct {
	TestID string            `json:"testId"`  // A/B test ID
	Routes map[string]string `json:"routes"`  // Map variant ID to node ID
}

// NewABSplitProcessor creates a new AB_SPLIT node processor
func NewABSplitProcessor(db *gorm.DB) *ABSplitProcessor {
	return &ABSplitProcessor{
		db:          db,
		testService: services.NewABTestService(db),
	}
}

// Type returns the node type this processor handles
func (p *ABSplitProcessor) Type() models.NodeType {
	return "AB_SPLIT"
}

// Validate checks if the AB_SPLIT node data is valid
func (p *ABSplitProcessor) Validate(node *models.AutomationNode) error {
	if node.Type != p.Type() {
		return fmt.Errorf("invalid node type for ABSplitProcessor")
	}

	var data ABSplitNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return fmt.Errorf("invalid AB split node data: %w", err)
	}

	if data.TestID == "" {
		return fmt.Errorf("testId is required")
	}

	if data.Routes == nil || len(data.Routes) < 2 {
		return fmt.Errorf("at least 2 routes are required for A/B testing")
	}

	return nil
}

// Process executes the AB_SPLIT node logic
func (p *ABSplitProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	var data ABSplitNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse AB split node data: %w", err)
	}

	// Assign contact to a variant (or get existing assignment)
	variant, err := p.testService.AssignVariant(context.Background(), data.TestID, ctx.ContactID)
	if err != nil {
		return nil, fmt.Errorf("failed to assign variant: %w", err)
	}

	// Get the route for this variant
	nextNodeID, ok := data.Routes[variant.ID]
	if !ok {
		return nil, fmt.Errorf("no route configured for variant %s", variant.ID)
	}

	return &automation.ProcessResult{
		NextNodeIDs: []string{nextNodeID},
		Message:     fmt.Sprintf("Assigned to variant %s (%d%% split)", variant.VariantName, variant.SplitPercentage),
		Data: map[string]interface{}{
			"testId":          data.TestID,
			"variantId":       variant.ID,
			"variantName":     variant.VariantName,
			"splitPercentage": variant.SplitPercentage,
		},
	}, nil
}
