package models

import (
	"kori/internal/events"

	"gorm.io/gorm"
)

func (a *APIKey) AfterCreate(tx *gorm.DB) error {
	if a.AssistantUserID != "" {
		return nil
	}

	// 🔍 Get resources for read and create actions
	var resources []Resource
	if err := tx.Where("name = ? AND action IN (?)", "emails", []string{"read", "create"}).Find(&resources).Error; err != nil {
		return err
	}

	log.Info("Found resources %v", resources)

	// 🔑 Get permissions for found resources
	var permissions []ResourcePermission
	var resourceIDs []string
	for _, r := range resources {
		resourceIDs = append(resourceIDs, r.ID)
	}

	log.Info("Found resources %v", resourceIDs)

	if err := tx.Where("resource_id IN ? AND scope IN ?", resourceIDs, []string{"emails:read", "emails:create"}).Find(&permissions).Error; err != nil {
		return err
	}

	log.Info("Found permissions %v", permissions)

	for _, p := range permissions {
		// create api key permission
		apiKeyPermission := &APIKeyPermission{
			KeyID:                a.ID,
			ResourcePermissionID: p.ID,
		}
		if err := tx.Create(apiKeyPermission).Error; err != nil {
			return err
		}
	}

	return nil
}

func (t *TeamInvite) AfterCreate(tx *gorm.DB) error {
	log.Info("Team invite created %v", t)
	events.Emit("invite.created", t)
	return nil
}

func (c *ContactImport) AfterCreate(tx *gorm.DB) error {
	log.Info("Contact import created %v", c)
	events.Emit("contact_import.created", c)
	return nil
}

func (c *Campaign) AfterCreate(tx *gorm.DB) error {
	if c.NewsletterID != nil {
		return nil
	}
	log.Info("Campaign created %s", c.ID)
	events.Emit("campaign.created", c)
	return nil
}

func (c *Contact) AfterCreate(tx *gorm.DB) error {
	log.Info("Contact created %s", c.ID)
	events.Emit("contact.created", c)
	if err := SyncSubscribersCountByID(tx, c.ListID); err != nil {
		return err
	}
	return nil
}

func (c *Contact) AfterDelete(tx *gorm.DB) error {
	log.Info("Contact deleted %s", c.ID)
	events.Emit("contact.deleted", c)
	if err := SyncSubscribersCountByID(tx, c.ListID); err != nil {
		return err
	}
	return nil
}

func (c *Contact) AfterUpdate(tx *gorm.DB) error {
	log.Info("Contact updated %s", c.ID)
	events.Emit("contact.updated", c)
	if err := SyncSubscribersCountByID(tx, c.ListID); err != nil {
		return err
	}
	return nil
}

func SyncSubscribersCount(tx *gorm.DB, mailingList *MailingList) error {
	if mailingList == nil {
		return nil
	}
	return SyncSubscribersCountByID(tx, mailingList.ID)
}

func SyncSubscribersCountByID(tx *gorm.DB, listID string) error {
	count := int64(0)
	err := tx.Model(&Contact{}).Where("list_id = ?", listID).Where("status = ?", SubscriberStatusActive).Count(&count)
	if err.Error != nil {
		return err.Error
	}
	return tx.Model(&MailingList{}).Where("id = ?", listID).Update("subscribers_count", count).Error
}
