package models

import "time"

type Resource struct {
	Base
	Name   string `gorm:"not null" json:"name"`   // e.g. "campaigns", "templates", "lists"
	Action string `gorm:"not null" json:"action"` // "create", "read", "update", "delete"
}

type ResourcePermission struct {
	Base
	ResourceID string    `gorm:"type:uuid;not null" json:"resourceId"`
	Resource   *Resource `json:"resource,omitempty"`
	// Scope defines the permission level, e.g. "read:campaigns", "write:templates"
	Scope string `gorm:"not null" json:"scope"`
}

type APIKeyPermission struct {
	Base
	KeyID                string              `gorm:"type:uuid;not null" json:"keyId"`
	Key                  *APIKey             `json:"key,omitempty"`
	ResourcePermissionID string              `gorm:"type:uuid;not null" json:"resourcePermissionId"`
	ResourcePermission   *ResourcePermission `json:"resourcePermission,omitempty"`
	CreatedAt            time.Time           `json:"createdAt"`
}

type UserPermission struct {
	Base
	UserID               string              `gorm:"type:uuid;not null" json:"userId"`
	User                 *User               `json:"user,omitempty"`
	ResourcePermissionID string              `gorm:"type:uuid;not null" json:"resourcePermissionId"`
	ResourcePermission   *ResourcePermission `json:"resourcePermission,omitempty"`
	CreatedAt            time.Time           `json:"createdAt"`
}
