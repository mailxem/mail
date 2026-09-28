package models

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BackfillTagWorkspaces preserves legacy associations without sharing tags across tenants.
// Unassigned originals are retained for audit; only contact associations move to scoped copies.
// Run inside the schema migration transaction before serving requests.
func BackfillTagWorkspaces(tx *gorm.DB) error {
	var legacy []Tag
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("team_id IS NULL").Order("id").Find(&legacy).Error; err != nil {
		return err
	}
	for _, old := range legacy {
		var teams []string
		if err := tx.Table("contacts").Distinct("contacts.team_id").Joins("JOIN contact_tags ON contacts.id = contact_tags.contact_id").Where("contact_tags.tag_id = ?", old.ID).Pluck("contacts.team_id", &teams).Error; err != nil {
			return err
		}
		for _, team := range teams {
			if uuid.Validate(team) != nil {
				return fmt.Errorf("legacy tag %s has an invalid contact workspace", old.ID)
			}
			var target Tag
			err := tx.Where("team_id = ? AND LOWER(name) = LOWER(?) AND value = ? AND is_deleted = ?", team, old.Name, old.Value, old.IsDeleted).Order("id").First(&target).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				target = Tag{Base: Base{IsDeleted: old.IsDeleted}, TeamID: team, Name: old.Name, Value: old.Value}
				if err = tx.Create(&target).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
			if err = tx.Exec("INSERT INTO contact_tags (contact_id, tag_id) SELECT contacts.id, ? FROM contacts JOIN contact_tags ON contacts.id = contact_tags.contact_id WHERE contact_tags.tag_id = ? AND contacts.team_id = ? ON CONFLICT DO NOTHING", target.ID, old.ID, team).Error; err != nil {
				return err
			}
			if err = tx.Exec("DELETE FROM contact_tags WHERE tag_id = ? AND contact_id IN (SELECT id FROM contacts WHERE team_id = ?)", old.ID, team).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
