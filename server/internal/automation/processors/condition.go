package processors

import (
	"encoding/json"
	"fmt"
	"kori/internal/automation"
	"kori/internal/models"
	"kori/internal/workflowconfig"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// ConditionProcessor handles CONDITION nodes
type ConditionProcessor struct {
	db *gorm.DB
}

// ConditionNodeData represents the data structure for CONDITION nodes
type ConditionNodeData struct {
	Conditions []Condition       `json:"conditions"`
	Operator   string            `json:"operator"` // "AND" or "OR"
	Branches   map[string]string `json:"branches"` // "true" -> nodeID, "false" -> nodeID
}

// Condition represents a single condition to evaluate
type Condition struct {
	Variable string `json:"variable"` // Variable name from context
	Operator string `json:"operator"` // ==, !=, >, <, >=, <=, contains, exists, not_exists
	Value    string `json:"value"`    // Value to compare against
}

// NewConditionProcessor creates a new CONDITION node processor
func NewConditionProcessor(db *gorm.DB) *ConditionProcessor {
	return &ConditionProcessor{db: db}
}

// Type returns the node type this processor handles
func (p *ConditionProcessor) Type() models.NodeType {
	return models.NodeTypeCondition
}

// Validate checks if the CONDITION node data is valid
func (p *ConditionProcessor) Validate(node *models.AutomationNode) error {
	if node.Type != models.NodeTypeCondition {
		return fmt.Errorf("invalid node type for ConditionProcessor")
	}

	return workflowconfig.Validate("CONDITION", node.Data)
}

// Process executes the CONDITION node logic
func (p *ConditionProcessor) Process(ctx *automation.ExecutionContext, node *models.AutomationNode) (*automation.ProcessResult, error) {
	if err := p.Validate(node); err != nil {
		return nil, err
	}
	var data ConditionNodeData
	if err := json.Unmarshal(node.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse condition node data: %w", err)
	}

	// Default operator is AND
	operator := "AND"
	if data.Operator != "" {
		operator = data.Operator
	}

	// Evaluate all conditions
	results := make([]bool, len(data.Conditions))
	for i, cond := range data.Conditions {
		result, err := p.evaluateCondition(cond, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate condition %d: %w", i, err)
		}
		results[i] = result
	}

	// Combine results based on operator
	finalResult := results[0]
	if operator == "AND" {
		for _, r := range results {
			finalResult = finalResult && r
		}
	} else { // OR
		finalResult = false
		for _, r := range results {
			finalResult = finalResult || r
		}
	}

	// Determine next node based on result
	var nextNodeID string
	branchKey := "false"
	if finalResult {
		branchKey = "true"
	}

	var edges []models.AutomationNodeEdge
	if err := p.db.Where("automation_id = ? AND source_id = ?", ctx.AutomationID, node.ID).Find(&edges).Error; err != nil {
		return nil, err
	}
	for _, edge := range edges {
		if edge.Label == branchKey {
			nextNodeID = edge.TargetID
		}
	}
	if nextNodeID == "" {
		return nil, fmt.Errorf("missing %s branch", branchKey)
	}

	return &automation.ProcessResult{
		NextNodeIDs: []string{nextNodeID},
		Message:     fmt.Sprintf("Condition evaluated to %v, routing to branch '%s'", finalResult, branchKey),
		Data: map[string]interface{}{
			"result":      finalResult,
			"branch":      branchKey,
			"evaluations": results,
		},
	}, nil
}

// evaluateCondition evaluates a single condition
func (p *ConditionProcessor) evaluateCondition(cond Condition, ctx *automation.ExecutionContext) (bool, error) {
	// Get variable value from context
	varValue, exists := ctx.GetVariable(cond.Variable)

	// Handle existence checks
	if cond.Operator == "exists" {
		return exists, nil
	}
	if cond.Operator == "not_exists" {
		return !exists, nil
	}

	if cond.Operator == "empty" {
		return !exists || varValue == nil || strings.TrimSpace(fmt.Sprint(varValue)) == "", nil
	}
	if cond.Operator == "not_empty" {
		return exists && varValue != nil && strings.TrimSpace(fmt.Sprint(varValue)) != "", nil
	}

	// For other operators, variable must exist
	if !exists {
		return false, nil
	}

	// Convert to string for comparison
	varStr := fmt.Sprintf("%v", varValue)
	condValue := cond.Value

	switch cond.Operator {
	case "==":
		return varStr == condValue, nil

	case "!=":
		return varStr != condValue, nil

	case "contains":
		return strings.Contains(strings.ToLower(varStr), strings.ToLower(condValue)), nil

	case "not_contains":
		return !strings.Contains(strings.ToLower(varStr), strings.ToLower(condValue)), nil
	case "starts_with":
		return strings.HasPrefix(strings.ToLower(varStr), strings.ToLower(condValue)), nil
	case "ends_with":
		return strings.HasSuffix(strings.ToLower(varStr), strings.ToLower(condValue)), nil
	case ">", "<", ">=", "<=":
		// Try numeric comparison first
		varNum, varErr := strconv.ParseFloat(varStr, 64)
		condNum, condErr := strconv.ParseFloat(condValue, 64)

		if varErr == nil && condErr == nil {
			switch cond.Operator {
			case ">":
				return varNum > condNum, nil
			case "<":
				return varNum < condNum, nil
			case ">=":
				return varNum >= condNum, nil
			case "<=":
				return varNum <= condNum, nil
			}
		}

		// Fall back to string comparison
		switch cond.Operator {
		case ">":
			return varStr > condValue, nil
		case "<":
			return varStr < condValue, nil
		case ">=":
			return varStr >= condValue, nil
		case "<=":
			return varStr <= condValue, nil
		}
	}

	return false, fmt.Errorf("unknown operator: %s", cond.Operator)
}
