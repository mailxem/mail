package models

import (
	"time"

	"gorm.io/gorm"
)

// AutomationVersion is a published, immutable copy of an automation's graph.
// Executions record the version they started on and keep running on it.
type AutomationVersion struct {
	Base
	AutomationID string               `gorm:"type:uuid;not null;uniqueIndex:idx_automation_version_number" json:"automationId"`
	Number       int                  `gorm:"not null;uniqueIndex:idx_automation_version_number" json:"number"`
	TriggerEvent string               `json:"triggerEvent"`
	Nodes        []AutomationNode     `gorm:"serializer:json;type:jsonb" json:"nodes,omitempty"`
	Edges        []AutomationNodeEdge `gorm:"serializer:json;type:jsonb" json:"edges,omitempty"`
	PublishedAt  time.Time            `gorm:"not null" json:"publishedAt"`
	PublishedBy  *string              `gorm:"type:uuid" json:"publishedBy"`
}

// LoadAutomationGraph fills a.Nodes and a.Edges from the given published
// version, or from the draft tables when version is 0.
func LoadAutomationGraph(db *gorm.DB, a *Automation, version int) error {
	if version == 0 {
		a.Nodes, a.Edges = nil, nil
		if err := db.Where("automation_id = ? AND is_deleted = ?", a.ID, false).Find(&a.Nodes).Error; err != nil {
			return err
		}
		return db.Where("automation_id = ?", a.ID).Find(&a.Edges).Error
	}
	var v AutomationVersion
	if err := db.Where("automation_id = ? AND number = ?", a.ID, version).First(&v).Error; err != nil {
		return err
	}
	a.Nodes, a.Edges = v.Nodes, v.Edges
	return nil
}

// BackfillAutomationVersions publishes every existing automation's current
// graph as version 1 and pins its executions to it, so nothing changes for
// contacts already inside. It runs once, when the versions table is created.
func BackfillAutomationVersions(tx *gorm.DB) error {
	var automations []Automation
	if err := tx.Where("is_deleted = ? AND live_version = ?", false, 0).Find(&automations).Error; err != nil {
		return err
	}
	for i := range automations {
		a := &automations[i]
		if err := LoadAutomationGraph(tx, a, 0); err != nil {
			return err
		}
		if err := tx.Create(&AutomationVersion{AutomationID: a.ID, Number: 1, TriggerEvent: a.TriggerEvent, Nodes: a.Nodes, Edges: a.Edges, PublishedAt: a.UpdatedAt}).Error; err != nil {
			return err
		}
		if err := tx.Model(&Automation{}).Where("id = ?", a.ID).UpdateColumn("live_version", 1).Error; err != nil {
			return err
		}
		if err := tx.Model(&AutomationExecution{}).Where("automation_id = ?", a.ID).UpdateColumn("version", 1).Error; err != nil {
			return err
		}
	}
	return nil
}
