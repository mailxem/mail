package processors

import (
	"encoding/json"
	"fmt"
	"kori/internal/automation"
	"kori/internal/models"
	"kori/internal/workflowconfig"

	"gorm.io/gorm"
)

// UpdateSubscriberProcessor handles UPDATE_SUBSCRIBER nodes
type UpdateSubscriberProcessor struct {
	db *gorm.DB
}

// UpdateSubscriberNodeData represents the data structure for UPDATE_SUBSCRIBER nodes
type UpdateSubscriberNodeData struct {
	Fields map[string]interface{} `json:"fields"` // Fields to update on the contact
}

// NewUpdateSubscriberProcessor creates a new UPDATE_SUBSCRIBER node processor
func NewUpdateSubscriberProcessor(db *gorm.DB) *UpdateSubscriberProcessor {
	return &UpdateSubscriberProcessor{db: db}
}

// Type returns the node type this processor handles
func (p *UpdateSubscriberProcessor) Type() models.NodeType {
	return models.NodeTypeUpdateSubscriber
}

// Validate checks if the UPDATE_SUBSCRIBER node data is valid
func (p *UpdateSubscriberProcessor) Validate(node *models.AutomationNode) error {
	if node.Type != models.NodeTypeUpdateSubscriber {
		return fmt.Errorf("invalid node type for UpdateSubscriberProcessor")
	}

	return workflowconfig.Validate("UPDATE_SUBSCRIBER", node.Data)
}

// Process executes the UPDATE_SUBSCRIBER node logic
func (p *UpdateSubscriberProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	if err := p.Validate(node); err != nil {
		return nil, err
	}
	var data UpdateSubscriberNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse update_subscriber node data: %w", err)
	}

	updates := map[string]interface{}{}
	for key, value := range data.Fields {
		if key == "linkedin" {
			key = "linked_in"
		}
		updates[key] = value
	}

	// Apply updates
	if err := p.db.Model(&models.Contact{}).Where("id = ? AND team_id = ? AND is_deleted = ?", ctx.ContactID, ctx.TeamID, false).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	// Reload contact to get updated values
	if err := p.db.Where("id = ? AND team_id = ? AND is_deleted = ?", ctx.ContactID, ctx.TeamID, false).First(ctx.Contact).Error; err != nil {
		return nil, fmt.Errorf("failed to reload contact: %w", err)
	}

	ctx.RefreshContactVariables()

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
		Message:     fmt.Sprintf("Updated %d field(s) on contact", len(updates)),
		Data: map[string]interface{}{
			"updatedFields": data.Fields,
			"contactId":     ctx.ContactID,
		},
	}, nil
}
