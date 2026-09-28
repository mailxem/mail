package processors

import (
	"fmt"
	"kori/internal/automation"
	"kori/internal/models"

	"gorm.io/gorm"
)

// ExitProcessor handles EXIT nodes
type ExitProcessor struct {
	db *gorm.DB
}

// NewExitProcessor creates a new EXIT node processor
func NewExitProcessor(db *gorm.DB) *ExitProcessor {
	return &ExitProcessor{db: db}
}

// Type returns the node type this processor handles
func (p *ExitProcessor) Type() models.NodeType {
	return models.NodeTypeExit
}

// Validate checks if the EXIT node data is valid
func (p *ExitProcessor) Validate(node *models.AutomationNode) error {
	if node.Type != models.NodeTypeExit {
		return fmt.Errorf("invalid node type for ExitProcessor")
	}
	return nil
}

// Process executes the EXIT node logic
func (p *ExitProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	// EXIT node marks the workflow as complete
	return &automation.ProcessResult{
		Complete: true,
		Message:  "Workflow completed successfully",
	}, nil
}
