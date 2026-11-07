package models

import (
	"fmt"
	"kori/internal/events"
	"kori/internal/workflowconfig"
	"time"

	"gorm.io/gorm"
)

func (a *APIKey) AfterCreate(tx *gorm.DB) error {
	if a.AssistantUserID != "" {
		return nil
	}
	if a.Admin {
		for _, name := range AdminKeyResources {
			var resource Resource
			if err := tx.Where("name = ? AND action = ? AND is_deleted = false", name, "admin").FirstOrCreate(&resource, Resource{Name: name, Action: "admin"}).Error; err != nil {
				return err
			}
			var permission ResourcePermission
			if err := tx.Where("resource_id = ? AND is_deleted = false", resource.ID).FirstOrCreate(&permission, ResourcePermission{ResourceID: resource.ID, Scope: "admin:" + name}).Error; err != nil {
				return err
			}
			if err := tx.Create(&APIKeyPermission{KeyID: a.ID, ResourcePermissionID: permission.ID}).Error; err != nil {
				return err
			}
		}
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

// BeforeSave keeps contact time zones to IANA names so sending rules can rely on them.
func (c *Contact) BeforeSave(tx *gorm.DB) error {
	if c.Timezone != "" && !workflowconfig.ValidTimezone(c.Timezone) {
		return fmt.Errorf("unknown time zone %q", c.Timezone)
	}
	return nil
}

func (c *Contact) AfterCreate(tx *gorm.DB) error {
	log.Info("Contact created %s", c.ID)
	events.Emit("contact.created", c)
	EmitTagsAdded(c, nil, c.Tags)
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

// BeforeUpdate remembers the stored list so AfterUpdate can tell a move from an edit.
func (c *Contact) BeforeUpdate(tx *gorm.DB) error {
	c.previousListID = ""
	if c.ID == "" {
		return nil
	}
	var stored []string
	if err := tx.Session(&gorm.Session{NewDB: true}).Model(&Contact{}).Where("id = ?", c.ID).Limit(1).Pluck("list_id", &stored).Error; err != nil {
		return err
	}
	if len(stored) == 1 {
		c.previousListID = stored[0]
	}
	return nil
}

func (c *Contact) AfterUpdate(tx *gorm.DB) error {
	log.Info("Contact updated %s", c.ID)
	events.Emit("contact.updated", c)
	if c.previousListID != "" && c.ListID != "" && c.ListID != c.previousListID {
		moved := *c
		events.Emit("list.subscriber_added", &moved)
		if err := SyncSubscribersCountByID(tx, c.previousListID); err != nil {
			return err
		}
	}
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

// TagAdded is the payload of "contact.tag_added": emitted once when a tag is newly attached.
type TagAdded struct {
	Contact *Contact
	Tag     Tag
	At      time.Time
}

// SegmentEntry is the payload of "segment.entered". Since is when the window of
// the entry began (the previous refresh of a dynamic segment), so a refresh that
// is retried or races another one names the same entries.
type SegmentEntry struct {
	TeamID     string
	SegmentID  string
	ContactIDs []string
	Since      time.Time
}

// EmitTagsAdded announces tags that were not on the contact before.
func EmitTagsAdded(contact *Contact, before map[string]bool, after []Tag) {
	snapshot := *contact // handlers run later, on other goroutines
	snapshot.Tags = nil
	contact = &snapshot
	now := time.Now()
	for _, tag := range after {
		if tag.ID != "" && !before[tag.ID] {
			events.Emit("contact.tag_added", &TagAdded{Contact: contact, Tag: tag, At: now})
		}
	}
}
