package automation

import (
	"kori/internal/models"
	"time"
)

// NodeProcessor defines the interface for processing automation nodes
type NodeProcessor interface {
	// Process executes the node logic and returns the result
	Process(ctx *ExecutionContext, node *models.AutomationNode) (*ProcessResult, error)

	// Validate checks if the node data is valid
	Validate(node *models.AutomationNode) error

	// Type returns the node type this processor handles
	Type() models.NodeType
}

// ProcessResult contains the outcome of processing a node
type ProcessResult struct {
	// NextNodeIDs contains the IDs of nodes to execute next
	// Multiple IDs for CONDITION nodes, single ID for linear flow
	NextNodeIDs []string

	// Wait specifies a delay before processing next node
	Wait *time.Duration

	// Complete indicates the workflow has reached a terminal state
	Complete bool

	// UpdateVars contains variable updates from this node
	UpdateVars map[string]interface{}

	// Message is a human-readable description of what happened
	Message string

	// Data contains additional metadata about the execution
	Data map[string]interface{}
}

// ProcessorRegistry manages all node processors
type ProcessorRegistry struct {
	processors map[models.NodeType]NodeProcessor
}

// NewProcessorRegistry creates a new processor registry
func NewProcessorRegistry() *ProcessorRegistry {
	return &ProcessorRegistry{
		processors: make(map[models.NodeType]NodeProcessor),
	}
}

// Register adds a processor to the registry
func (r *ProcessorRegistry) Register(processor NodeProcessor) {
	r.processors[processor.Type()] = processor
}

// Get retrieves a processor by node type
func (r *ProcessorRegistry) Get(nodeType models.NodeType) (NodeProcessor, bool) {
	processor, exists := r.processors[nodeType]
	return processor, exists
}

// Has checks if a processor exists for the given node type
func (r *ProcessorRegistry) Has(nodeType models.NodeType) bool {
	_, exists := r.processors[nodeType]
	return exists
}
