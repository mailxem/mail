package services

import (
	"context"
	"fmt"
	"kori/internal/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// draft loads an automation with its draft graph and trigger.
func (s *AutomationService) draft(ctx context.Context, id string) (*models.Automation, error) {
	var a models.Automation
	if err := s.db.WithContext(ctx).Where("id = ? AND is_deleted = ?", id, false).First(&a).Error; err != nil {
		return nil, err
	}
	if a.DraftTriggerEvent != "" {
		a.TriggerEvent = a.DraftTriggerEvent
	}
	return &a, models.LoadAutomationGraph(s.db.WithContext(ctx), &a, 0)
}

// Publish makes the draft the new live version. New contacts start on it;
// contacts already inside stay on their version unless moveRunning is set,
// in which case they move by step ID and those on a removed step exit.
func (s *AutomationService) Publish(ctx context.Context, id, userID string, moveRunning bool) (*models.AutomationVersion, error) {
	a, err := s.draft(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.ValidateGraph(a); err != nil {
		return nil, err
	}
	if err := s.ValidateConfiguration(a); err != nil {
		return nil, err
	}
	v := &models.AutomationVersion{AutomationID: id, TriggerEvent: a.TriggerEvent, Nodes: a.Nodes, Edges: a.Edges, PublishedAt: time.Now()}
	if userID != "" {
		v.PublishedBy = &userID
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked models.Automation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND is_deleted = ?", id, false).First(&locked).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.AutomationVersion{}).Where("automation_id = ?", id).Select("COALESCE(MAX(number), 0) + 1").Scan(&v.Number).Error; err != nil {
			return err
		}
		if err := tx.Create(v).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Automation{}).Where("id = ?", id).Updates(map[string]interface{}{"live_version": v.Number, "has_draft": false, "trigger_event": v.TriggerEvent, "draft_trigger_event": "", "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		if !moveRunning {
			return nil
		}
		ids := make([]string, len(v.Nodes))
		for i, n := range v.Nodes {
			ids[i] = n.ID
		}
		inside := []string{string(models.ExecutionStatusRunning), string(models.ExecutionStatusWaiting), string(models.ExecutionStatusPaused)}
		if err := tx.Model(&models.AutomationExecution{}).Where("automation_id = ? AND version <> ? AND status IN ? AND (current_node_id IS NULL OR current_node_id IN ?)", id, v.Number, inside, ids).
			UpdateColumn("version", v.Number).Error; err != nil {
			return err
		}
		return tx.Model(&models.AutomationExecution{}).Where("automation_id = ? AND version <> ? AND status IN ?", id, v.Number, inside).
			UpdateColumns(map[string]interface{}{"status": models.ExecutionStatusCompleted, "completed_at": time.Now(), "error": fmt.Sprintf("Exited: its step was removed in version %d", v.Number)}).Error
	})
	if err != nil {
		return nil, err
	}
	automationLog.Success("Published version %d of automation %s", v.Number, id)
	return v, nil
}

// RestoreDraft replaces the draft with a published version. Discarding the
// draft is restoring the live version.
func (s *AutomationService) RestoreDraft(ctx context.Context, id string, number int) error {
	a, err := s.draft(ctx, id)
	if err != nil {
		return err
	}
	if number == 0 {
		number = a.LiveVersion
	}
	var v models.AutomationVersion
	if err := s.db.WithContext(ctx).Where("automation_id = ? AND number = ?", id, number).First(&v).Error; err != nil {
		return err
	}
	a.TriggerEvent, a.Nodes, a.Edges = v.TriggerEvent, v.Nodes, v.Edges
	if err := s.SaveGraph(ctx, a); err != nil {
		return err
	}
	if number != a.LiveVersion {
		return nil
	}
	return s.db.WithContext(ctx).Model(&models.Automation{}).Where("id = ?", id).Updates(map[string]interface{}{"has_draft": false, "draft_trigger_event": ""}).Error
}

// VersionSummary is a published version without its graph.
type VersionSummary struct {
	Number       int       `json:"number"`
	TriggerEvent string    `json:"triggerEvent"`
	PublishedAt  time.Time `json:"publishedAt"`
	PublishedBy  string    `json:"publishedBy"`
}

func (s *AutomationService) Versions(ctx context.Context, id string) ([]VersionSummary, error) {
	versions := []VersionSummary{}
	return versions, s.db.WithContext(ctx).Table("automation_versions v").
		Select("v.number, v.trigger_event, v.published_at, COALESCE(NULLIF(TRIM(u.first_name || ' ' || u.last_name), ''), u.email, '') published_by").
		Joins("LEFT JOIN users u ON u.id = v.published_by").
		Where("v.automation_id = ?", id).Order("v.number DESC").Scan(&versions).Error
}
