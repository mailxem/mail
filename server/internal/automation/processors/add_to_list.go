package processors

import (
	"encoding/json"
	"fmt"
	"kori/internal/automation"
	"kori/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AddToListProcessor handles ADD_TO_LIST nodes
type AddToListProcessor struct {
	db *gorm.DB
}

// AddToListNodeData represents the data structure for ADD_TO_LIST nodes
type AddToListNodeData struct {
	ListID string `json:"listId"` // Mailing list ID to add contact to
}

// NewAddToListProcessor creates a new ADD_TO_LIST node processor
func NewAddToListProcessor(db *gorm.DB) *AddToListProcessor {
	return &AddToListProcessor{db: db}
}

// Type returns the node type this processor handles
func (p *AddToListProcessor) Type() models.NodeType {
	return models.NodeTypeAddToList
}

// Validate checks if the ADD_TO_LIST node data is valid
func (p *AddToListProcessor) Validate(node *models.AutomationNode) error {
	if node.Type != models.NodeTypeAddToList {
		return fmt.Errorf("invalid node type for AddToListProcessor")
	}

	var data AddToListNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return fmt.Errorf("invalid add_to_list node data: %w", err)
	}

	if data.ListID == "" {
		return fmt.Errorf("listId is required")
	}

	return nil
}

// Process executes the ADD_TO_LIST node logic
func (p *AddToListProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	var data AddToListNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse add_to_list node data: %w", err)
	}

	if err := p.Validate(node); err != nil {
		return nil, err
	}
	var list models.MailingList
	if err := p.db.Transaction(func(tx *gorm.DB) error {
		var contact models.Contact
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND team_id = ? AND is_deleted = ?", ctx.ContactID, ctx.TeamID, false).First(&contact).Error; err != nil {
			return err
		}
		// Lock affected lists in ID order so opposite moves cannot deadlock.
		var lists []models.MailingList
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id IN ? AND team_id = ? AND is_deleted = ?", []string{contact.ListID, data.ListID}, ctx.TeamID, false).Order("id").Find(&lists).Error; err != nil {
			return err
		}
		for _, candidate := range lists {
			if candidate.ID == data.ListID {
				list = candidate
			}
		}
		if list.ID == "" {
			return fmt.Errorf("mailing list not found or access denied")
		}
		if contact.ListID == list.ID {
			ctx.Contact = &contact
			return nil
		}
		previous := contact.ListID
		if err := tx.Model(&contact).Update("list_id", list.ID).Error; err != nil {
			return err
		}
		if err := models.SyncSubscribersCountByID(tx, previous); err != nil {
			return err
		}
		if err := models.SyncSubscribersCountByID(tx, list.ID); err != nil {
			return err
		}

		ctx.Contact = &contact
		return nil
	}); err != nil {
		return nil, err
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
		Message:     fmt.Sprintf("Moved contact to list '%s'", list.Name),
		UpdateVars: map[string]interface{}{
			"current_list_id":   data.ListID,
			"current_list_name": list.Name,
		},
		Data: map[string]interface{}{
			"listId":   list.ID,
			"listName": list.Name,
		},
	}, nil
}
