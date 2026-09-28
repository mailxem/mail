package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/models"
	"kori/internal/workflowconfig"
	"strings"
)

func (s *AutomationService) SaveGraph(ctx context.Context, a *models.Automation) error {
	if len(strings.TrimSpace(a.Name)) < 2 || len(a.Name) > 120 || len(a.Description) > 1000 {
		return fmt.Errorf("a name of 2–120 characters is required")
	}
	if a.TriggerEvent == "" {
		a.TriggerEvent = "manual"
	}
	switch a.TriggerEvent {
	case "manual", "contact.created", "email.opened", "email.clicked", "form.completed":
	default:
		return fmt.Errorf("unsupported trigger")
	}
	if len(a.Nodes) > 100 || len(a.Edges) > 200 {
		return fmt.Errorf("workflow is too large")
	}
	if err := s.ValidateGraph(a); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if a.ID != "" {
			var old models.Automation
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND team_id = ? AND is_deleted = ?", a.ID, a.TeamID, false).First(&old).Error; err != nil {
				return err
			}
			if old.IsActive {
				return fmt.Errorf("pause this workflow before editing")
			}
			var running int64
			if err := tx.Model(&models.AutomationExecution{}).Where("automation_id = ? AND status IN ?", a.ID, []string{"RUNNING", "WAITING", "PAUSED"}).Count(&running).Error; err != nil {
				return err
			}
			if running > 0 {
				return fmt.Errorf("wait for in-flight executions before changing the graph")
			}
			a.Base = old.Base
		} else {
			a.ID = uuid.NewString()
		}
		a.IsActive = false
		a.Team = nil
		for i := range a.Nodes {
			n := &a.Nodes[i]
			if uuid.Validate(n.ID) != nil {
				return fmt.Errorf("node IDs must be UUIDs")
			}
			var count int64
			if err := tx.Model(&models.AutomationNode{}).Where("id = ? AND automation_id <> ?", n.ID, a.ID).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return fmt.Errorf("invalid node ownership")
			}
			n.AutomationID = a.ID
			n.Automation = nil
			n.EdgesFrom = nil
			n.EdgesTo = nil
			n.IsDeleted = false
		}
		for i := range a.Edges {
			e := &a.Edges[i]
			e.Base = models.Base{ID: uuid.NewString()}
			e.AutomationID = a.ID
			e.Automation = nil
			e.Source = nil
			e.Target = nil
		}
		if err := tx.Omit(clause.Associations).Save(a).Error; err != nil {
			return err
		}
		if err := tx.Where("automation_id = ?", a.ID).Delete(&models.AutomationNodeEdge{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.AutomationNode{}).Where("automation_id = ?", a.ID).Update("is_deleted", true).Error; err != nil {
			return err
		}
		for i := range a.Nodes {
			if err := tx.Omit(clause.Associations).Save(&a.Nodes[i]).Error; err != nil {
				return err
			}
		}
		if len(a.Edges) > 0 {
			return tx.Omit(clause.Associations).Create(&a.Edges).Error
		}
		return nil
	})
}
func (s *AutomationService) ValidateConfiguration(a *models.Automation) error {
	if a.TriggerEvent == "form.completed" {
		for _, node := range a.Nodes {
			if node.Type != models.NodeTypeStart {
				continue
			}
			var config struct {
				FormID string `json:"formId"`
			}
			if json.Unmarshal(node.Data, &config) != nil || uuid.Validate(config.FormID) != nil {
				return fmt.Errorf("choose a form for this submission trigger")
			}
			var count int64
			if err := s.db.Model(&models.Form{}).Where("id = ? AND team_id = ? AND is_deleted = ?", config.FormID, a.TeamID, false).Count(&count).Error; err != nil || count != 1 {
				return fmt.Errorf("choose a form in this workspace")
			}
		}
	}
	outgoing := map[string][]models.AutomationNodeEdge{}
	for _, edge := range a.Edges {
		outgoing[edge.SourceID] = append(outgoing[edge.SourceID], edge)
	}
	for _, n := range a.Nodes {
		if n.Type == models.NodeTypeExit && len(outgoing[n.ID]) != 0 {
			return fmt.Errorf("finish steps cannot have outgoing connections")
		}
		if n.Type != models.NodeTypeExit && n.Type != models.NodeTypeCondition && n.Type != "PERCENTAGE_SPLIT" && len(outgoing[n.ID]) != 1 {
			return fmt.Errorf("each step must have exactly one continuation; use a condition to branch")
		}
		if (n.Type == models.NodeTypeCondition || n.Type == "PERCENTAGE_SPLIT") && len(outgoing[n.ID]) != 2 {
			return fmt.Errorf("conditions require exactly two paths")
		}

		var data map[string]interface{}
		if len(n.Data) == 0 && (n.Type == models.NodeTypeStart || n.Type == models.NodeTypeExit) {
			continue
		}
		if err := json.Unmarshal(n.Data, &data); err != nil {
			return fmt.Errorf("invalid step configuration")
		}
		switch n.Type {
		case models.NodeTypeStart, models.NodeTypeExit:
		case models.NodeTypeEmail:
			for key, model := range map[string]interface{}{"templateId": &models.Template{}, "smtpConfigId": &models.SMTPConfig{}} {
				id, _ := data[key].(string)
				var count int64
				if s.db.Model(model).Where("id = ? AND team_id = ? AND is_deleted = ?", id, a.TeamID, false).Count(&count).Error != nil || count != 1 {
					return fmt.Errorf("email step needs a workspace template and sender")
				}
			}
			var template models.Template
			if err := s.db.Preload("Category").First(&template, "id = ?", data["templateId"]).Error; err != nil {
				return err
			}
			if template.HTMLBody == "" && template.HtmlFileID == "" {
				return fmt.Errorf("email template needs content")
			}
			if template.Category != nil && template.Category.Name == "Marketing" {
				address, _ := data["postalAddress"].(string)
				if len(strings.TrimSpace(address)) < 8 || len(address) > 500 {
					return fmt.Errorf("marketing email needs a sender postal address")
				}
			}
			if subject, _ := data["subject"].(string); len(subject) > 200 || strings.ContainsAny(subject, "\r\n") {
				return fmt.Errorf("invalid email subject")
			}
		case models.NodeTypeWait, models.NodeTypeTag, models.NodeTypeUpdateSubscriber, "SET_VARIABLE":
			if err := workflowconfig.Validate(string(n.Type), n.Data); err != nil {
				return err
			}
		case models.NodeTypeAddToList:
			listID, _ := data["listId"].(string)
			var count int64
			if uuid.Validate(listID) != nil || s.db.Model(&models.MailingList{}).Where("id = ? AND team_id = ? AND is_deleted = ?", listID, a.TeamID, false).Count(&count).Error != nil || count != 1 {
				return fmt.Errorf("choose a list in this workspace")
			}
		case models.NodeTypeCondition, "PERCENTAGE_SPLIT":
			if err := workflowconfig.Validate(string(n.Type), n.Data); err != nil {
				return err
			}
			labels := map[string]bool{}
			for _, edge := range a.Edges {
				if edge.SourceID == n.ID {
					labels[edge.Label] = true
				}
			}
			left, right := "true", "false"
			if n.Type == "PERCENTAGE_SPLIT" {
				left, right = "A", "B"
			}
			if !labels[left] || !labels[right] {
				return fmt.Errorf("branch needs %s and %s paths", left, right)
			}
			if branches, ok := data["branches"].(map[string]interface{}); ok && len(branches) > 0 {
				return fmt.Errorf("use graph edges for condition paths")
			}
		default:
			return fmt.Errorf("unsupported step type: %s", n.Type)
		}
	}
	return nil
}
