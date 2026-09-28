package models

import (
	"time"

	"gorm.io/datatypes"
)

type User struct {
	Base
	Bio              string           `json:"bio"`
	Email            string           `gorm:"uniqueIndex;not null" json:"email"`
	Password         string           `gorm:"not null" json:"-"`
	FirstName        string           `json:"firstName"`
	LastName         string           `json:"lastName"`
	Role             UserRole         `gorm:"not null;default:'member'" json:"role"`
	TeamID           string           `gorm:"type:uuid;not null" json:"teamId"`
	Team             *Team            `json:"team,omitempty"`
	Permissions      []UserPermission `gorm:"foreignKey:UserID" json:"permissions,omitempty"`
	Invites          []TeamInvite     `gorm:"foreignKey:InviterID" json:"invites,omitempty"`
	Files            []File           `gorm:"foreignKey:UserID" json:"files,omitempty"`
	ProfilePicture   File             `gorm:"foreignKey:ProfilePictureID" json:"profilePicture,omitempty"`
	ProfilePictureID string           `gorm:"type:uuid;default:NULL" json:"profilePictureId,omitempty"`
	Provider         string           `gorm:"default:'local'" json:"provider"`          // 'local', 'google', etc.
	ProviderID       string           `gorm:"index" json:"providerId,omitempty"`        // ID from the OAuth provider
	ProviderData     datatypes.JSON   `gorm:"type:jsonb" json:"providerData,omitempty"` // Additional data from provider
}

type PasswordReset struct {
	Base
	User      *User     `json:"user,omitempty"`
	UserID    string    `gorm:"type:uuid;not null" json:"userId"`
	Code      string    `gorm:"not null" json:"code"`
	Used      bool      `gorm:"default:false" json:"used"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type AuthTransaction struct {
	Base
	UserID    string    `gorm:"type:uuid;not null" json:"userId"`
	User      *User     `json:"user,omitempty"`
	TeamID    string    `gorm:"type:uuid;not null" json:"teamId"`
	Team      *Team     `json:"team,omitempty"`
	Token     string    `gorm:"not null" json:"token"`
	Refresh   string    `gorm:"not null" json:"refresh"`
	IPAddress string    `json:"ipAddress"`
	UserAgent string    `json:"userAgent"`
	ExpiresAt time.Time `json:"expiresAt"`
}
