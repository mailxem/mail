package processors

import (
	"encoding/json"
	"fmt"
	"kori/internal/automation"
	"kori/internal/models"
	"kori/internal/workflowconfig"

	"gorm.io/gorm"
)

// TagProcessor handles TAG nodes
type TagProcessor struct {
	db *gorm.DB
}

// TagNodeData represents the data structure for TAG nodes
type TagNodeData struct {
	Action string   `json:"action"` // "add" or "remove"
	Tags   []string `json:"tags"`   // Tag names to add/remove
}

// NewTagProcessor creates a new TAG node processor
func NewTagProcessor(db *gorm.DB) *TagProcessor {
	return &TagProcessor{db: db}
}

// Type returns the node type this processor handles
func (p *TagProcessor) Type() models.NodeType {
	return models.NodeTypeTag
}

// Validate checks if the TAG node data is valid
func (p *TagProcessor) Validate(node *models.AutomationNode) error {
	if node.Type != models.NodeTypeTag {
		return fmt.Errorf("invalid node type for TagProcessor")
	}

	return workflowconfig.Validate("TAG", node.Data)
}

// Process executes the TAG node logic
func (p *TagProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	if err := p.Validate(node); err != nil {
		return nil, err
	}
	if ctx.Contact == nil || ctx.Contact.TeamID != ctx.TeamID {
		return nil, fmt.Errorf("contact is outside workspace")
	}
	var data TagNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse tag node data: %w", err)
	}

	var processedTags []string

	if data.Action == "add" {
		// Add tags to contact
		for _, tagName := range data.Tags {
			// Find or create tag
			var tag models.Tag
			err := p.db.Where("team_id = ? AND name = ? AND is_deleted = false", ctx.Contact.TeamID, tagName).First(&tag).Error
			if err == gorm.ErrRecordNotFound {
				// Create new tag
				tag = models.Tag{Name: tagName, TeamID: ctx.Contact.TeamID}
				if err := p.db.Create(&tag).Error; err != nil {
					return nil, fmt.Errorf("failed to create tag '%s': %w", tagName, err)
				}
			} else if err != nil {
				return nil, fmt.Errorf("failed to query tag '%s': %w", tagName, err)
			}

			// Associate tag with contact (using many-to-many relationship)
			if err := p.db.Model(ctx.Contact).Association("Tags").Append(&tag); err != nil {
				return nil, fmt.Errorf("failed to add tag '%s' to contact: %w", tagName, err)
			}

			processedTags = append(processedTags, tagName)
		}
	} else { // remove
		// Remove tags from contact
		for _, tagName := range data.Tags {
			var tag models.Tag
			err := p.db.Where("team_id = ? AND name = ? AND is_deleted = false", ctx.Contact.TeamID, tagName).First(&tag).Error
			if err == gorm.ErrRecordNotFound {
				// Tag doesn't exist, skip
				continue
			} else if err != nil {
				return nil, fmt.Errorf("failed to query tag '%s': %w", tagName, err)
			}

			// Remove association
			if err := p.db.Model(ctx.Contact).Association("Tags").Delete(&tag); err != nil {
				return nil, fmt.Errorf("failed to remove tag '%s' from contact: %w", tagName, err)
			}

			processedTags = append(processedTags, tagName)
		}
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

	action := "Added"
	if data.Action == "remove" {
		action = "Removed"
	}

	return &automation.ProcessResult{
		NextNodeIDs: nextNodeIDs,
		Message:     fmt.Sprintf("%s %d tag(s) %s contact", action, len(processedTags), data.Action),
		Data: map[string]interface{}{
			"action":        data.Action,
			"processedTags": processedTags,
			"requestedTags": data.Tags,
		},
	}, nil
}
