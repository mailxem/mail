package processors

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"kori/internal/automation"
	"kori/internal/models"
	"kori/internal/workflowconfig"
)

type FlowControlProcessor struct {
	db   *gorm.DB
	kind models.NodeType
}

func NewPercentageSplitProcessor(db *gorm.DB) *FlowControlProcessor {
	return &FlowControlProcessor{db: db, kind: "PERCENTAGE_SPLIT"}
}
func NewSetVariableProcessor(db *gorm.DB) *FlowControlProcessor {
	return &FlowControlProcessor{db: db, kind: "SET_VARIABLE"}
}
func (p *FlowControlProcessor) Type() models.NodeType { return p.kind }
func (p *FlowControlProcessor) Validate(n *models.AutomationNode) error {
	if n.Type != p.kind {
		return fmt.Errorf("unexpected node type")
	}
	return workflowconfig.Validate(string(n.Type), n.Data)
}
func (p *FlowControlProcessor) Process(ctx *automation.ExecutionContext, n *models.AutomationNode) (*automation.ProcessResult, error) {
	if err := p.Validate(n); err != nil {
		return nil, err
	}
	var edges []models.AutomationNodeEdge
	if err := p.db.Where("automation_id = ? AND source_id = ?", ctx.AutomationID, n.ID).Find(&edges).Error; err != nil {
		return nil, err
	}
	if p.kind == "SET_VARIABLE" {
		var c struct {
			Variable string `json:"variable"`
			Value    string `json:"value"`
		}
		if err := json.Unmarshal(n.Data, &c); err != nil {
			return nil, err
		}
		if len(edges) != 1 {
			return nil, fmt.Errorf("variable step needs one continuation")
		}
		return &automation.ProcessResult{NextNodeIDs: []string{edges[0].TargetID}, UpdateVars: map[string]interface{}{c.Variable: c.Value}, Message: "Set " + c.Variable}, nil
	}
	var c struct {
		Percentage int `json:"percentage"`
	}
	if err := json.Unmarshal(n.Data, &c); err != nil {
		return nil, err
	}
	// Stable assignment across process restarts and task retries, scoped to this step.
	hash := sha256.Sum256([]byte(ctx.AutomationID + ":" + n.ID + ":" + ctx.ContactID))
	bucket := binary.BigEndian.Uint64(hash[:8]) % 100
	branch := "B"
	if bucket < uint64(c.Percentage) {
		branch = "A"
	}
	if len(edges) != 2 {
		return nil, fmt.Errorf("split needs two paths")
	}
	for _, edge := range edges {
		if edge.Label == branch {
			return &automation.ProcessResult{NextNodeIDs: []string{edge.TargetID}, Message: "Assigned to path " + branch, Data: map[string]interface{}{"branch": branch, "percentage": c.Percentage}}, nil
		}
	}
	return nil, fmt.Errorf("missing split path %s", branch)
}
