package models

import (
	"fmt"
	"kori/internal/config"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"

	console "kori/internal/utils/logger"

	"gorm.io/gorm"
)

var log = console.New("SEEDER")

// Default resources and their actions
var defaultResources = []Resource{
	{Name: "marketing", Action: "read"}, {Name: "marketing", Action: "create"},
	// Campaign resources
	{Name: "campaigns", Action: "create"},
	{Name: "campaigns", Action: "read"},
	{Name: "campaigns", Action: "update"},
	{Name: "campaigns", Action: "delete"},

	// Template resources
	{Name: "templates", Action: "create"},
	{Name: "templates", Action: "read"},
	{Name: "templates", Action: "update"},
	{Name: "templates", Action: "delete"},

	// Mailing list resources
	{Name: "lists", Action: "create"},
	{Name: "lists", Action: "read"},
	{Name: "lists", Action: "update"},
	{Name: "lists", Action: "delete"},

	// Contact resources
	{Name: "contacts", Action: "create"},
	{Name: "contacts", Action: "read"},
	{Name: "contacts", Action: "update"},
	{Name: "contacts", Action: "delete"},

	// Team resources
	{Name: "teams", Action: "create"},
	{Name: "teams", Action: "read"},
	{Name: "teams", Action: "update"},
	{Name: "teams", Action: "delete"},

	// User resources
	{Name: "users", Action: "create"},
	{Name: "users", Action: "read"},
	{Name: "users", Action: "update"},
	{Name: "users", Action: "delete"},

	// API key resources
	{Name: "api_keys", Action: "create"},
	{Name: "api_keys", Action: "read"},
	{Name: "api_keys", Action: "update"},
	{Name: "api_keys", Action: "delete"},

	// API key usage resources
	{Name: "api_key_usage", Action: "read"},

	// Automation resources
	{Name: "automations", Action: "create"},
	{Name: "automations", Action: "read"},
	{Name: "automations", Action: "update"},
	{Name: "automations", Action: "delete"},

	// SMTP config resources
	{Name: "smtp_configs", Action: "create"},
	{Name: "smtp_configs", Action: "read"},
	{Name: "smtp_configs", Action: "update"},
	{Name: "smtp_configs", Action: "delete"},

	// IMAP config resources
	{Name: "imap_configs", Action: "create"},
	{Name: "imap_configs", Action: "read"},
	{Name: "imap_configs", Action: "update"},
	{Name: "imap_configs", Action: "delete"},

	// Domain resources
	{Name: "domains", Action: "create"},
	{Name: "domains", Action: "read"},
	{Name: "domains", Action: "update"},
	{Name: "domains", Action: "delete"},

	// Permission resources
	{Name: "permissions", Action: "create"},
	{Name: "permissions", Action: "read"},
	{Name: "permissions", Action: "update"},
	{Name: "permissions", Action: "delete"},

	// Role resources
	{Name: "roles", Action: "create"},
	{Name: "roles", Action: "read"},
	{Name: "roles", Action: "update"},
	{Name: "roles", Action: "delete"},

	// Webhook resources
	{Name: "webhooks", Action: "create"},
	{Name: "webhooks", Action: "read"},
	{Name: "webhooks", Action: "update"},
	{Name: "webhooks", Action: "delete"},

	// Model resources
	{Name: "models", Action: "create"},
	{Name: "models", Action: "read"},
	{Name: "models", Action: "update"},
	{Name: "models", Action: "delete"},

	// Email resources
	{Name: "emails", Action: "create"},
	{Name: "emails", Action: "read"},

	// Analytics resources
	{Name: "analytics", Action: "read"},

	// Team invite resources
	{Name: "team_invites", Action: "create"},
	{Name: "team_invites", Action: "read"},
	{Name: "team_invites", Action: "update"},
	{Name: "team_invites", Action: "delete"},

	// Contact import resources
	{Name: "contact_imports", Action: "create"},
	{Name: "contact_imports", Action: "read"},
	{Name: "contact_imports", Action: "update"},
	{Name: "contact_imports", Action: "delete"},

	// File resources
	{Name: "files", Action: "create"},
	{Name: "files", Action: "read"},
	{Name: "files", Action: "update"},
	{Name: "files", Action: "delete"},

	// Team settings resources
	{Name: "team_settings", Action: "create"},
	{Name: "team_settings", Action: "read"},
	{Name: "team_settings", Action: "update"},
	{Name: "team_settings", Action: "delete"},

	// Branding settings resources
	{Name: "branding_settings", Action: "create"},
	{Name: "branding_settings", Action: "read"},
	{Name: "branding_settings", Action: "update"},
	{Name: "branding_settings", Action: "delete"},
}

// Role-based permission mappings
var rolePermissions = map[UserRole][]string{
	UserRoleAdmin: {
		// Admin has all permissions
		"marketing:*", "campaigns:*", "templates:*", "lists:*", "contacts:*",
		"teams:*", "users:*", "api_keys:*", "automations:*",
		"smtp_configs:*", "domains:*",
		"permissions:*",
		"roles:*",
		"webhooks:*",
		"models:*",
		"emails:*",
		"analytics:*",
		"api_key_usage:*",
		"team_invites:*",
		"contact_imports:*",
		"files:*",
		"team_settings:*",
		"branding_settings:*",
		"imap_configs:*",
	},
	UserRoleMember: {
		// Member has limited permissions
		"marketing:read", "marketing:create", "campaigns:read", "campaigns:create",
		"templates:read",
		"lists:read",
		"contacts:read", "contacts:create",
		"teams:read",
		"users:read",
		"automations:read",
		"automations:create",
		"automations:update",
		"domains:read",
		"webhooks:read",
		"models:read",
		"emails:read",
		"emails:create",
		"analytics:read",
		"api_key_usage:read",
		"team_invites:read",
		"contact_imports:read",
		"files:read",
		"team_settings:read",
		"branding_settings:read",
		"imap_configs:read",
	},
	UserRoleSuperAdmin: {
		// SuperAdmin has all permissions
		"*:*",
	},
}

// SeedPermissions creates default resources and permissions
func SeedPermissions(db *gorm.DB) error {
	// Create resources
	for _, resource := range defaultResources {
		if err := db.FirstOrCreate(&resource, Resource{
			Name:   resource.Name,
			Action: resource.Action,
		}).Error; err != nil {
			return fmt.Errorf("failed to create resource %s:%s: %v", resource.Name, resource.Action, err)
		}
	}

	// Create resource permissions for each role
	for role, permissions := range rolePermissions {
		log.Info("Creating permissions for role: %s", role)

		for _, permScope := range permissions {
			// Handle wildcard permissions
			if strings.HasSuffix(permScope, ":*") {
				resourceName := strings.TrimSuffix(permScope, ":*") // Remove :*
				var resources []Resource
				if err := db.Where("name = ?", resourceName).Find(&resources).Error; err != nil {
					return fmt.Errorf("failed to find resources for %s: %v", resourceName, err)
				}

				// Create permissions for all actions of this resource
				for _, resource := range resources {
					if err := createResourcePermission(db, resource); err != nil {
						return err
					}
				}
			} else {
				// Handle specific permissions
				parts := strings.Split(permScope, ":")
				if len(parts) != 2 {
					return fmt.Errorf("invalid permission scope format: %s", permScope)
				}

				resourceName, action := parts[0], parts[1]
				var resource Resource
				if err := db.Where("name = ? AND action = ?", resourceName, action).First(&resource).Error; err != nil {
					return fmt.Errorf("failed to find resource %s:%s: %v", resourceName, action, err)
				}

				if err := createResourcePermission(db, resource); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func createResourcePermission(db *gorm.DB, resource Resource) error {
	scope := fmt.Sprintf("%s:%s", resource.Name, resource.Action)

	permission := ResourcePermission{
		ResourceID: resource.ID,
		Scope:      scope,
	}

	if err := db.FirstOrCreate(&permission, ResourcePermission{
		ResourceID: resource.ID,
		Scope:      scope,
	}).Error; err != nil {
		return fmt.Errorf("failed to create permission %s: %v", scope, err)
	}

	return nil
}

// AssignDefaultPermissions assigns default permissions to a user based on their role
func AssignDefaultPermissions(db *gorm.DB, user *User) error {
	var permissions []ResourcePermission

	if user.Role == UserRoleAdmin {
		// For admin, get all resource permissions
		if err := db.Find(&permissions).Error; err != nil {
			return fmt.Errorf("failed to fetch permissions: %v", err)
		}
	} else {
		// For other roles, get specific permissions based on rolePermissions mapping
		rolePerm := rolePermissions[user.Role]
		for _, permScope := range rolePerm {
			if strings.HasSuffix(permScope, ":*") {
				// Handle wildcard permissions
				resourceName := strings.TrimSuffix(permScope, ":*")
				var resources []Resource
				if err := db.Where("name = ?", resourceName).Find(&resources).Error; err != nil {
					return fmt.Errorf("failed to find resources for %s: %v", resourceName, err)
				}

				for _, resource := range resources {
					var perm ResourcePermission
					if err := db.Where("resource_id = ?", resource.ID).First(&perm).Error; err != nil {
						return fmt.Errorf("failed to find permission for resource %s: %v", resource.Name, err)
					}
					permissions = append(permissions, perm)
				}
			} else {
				// Handle specific permissions
				parts := strings.Split(permScope, ":")
				if len(parts) != 2 {
					return fmt.Errorf("invalid permission scope format: %s", permScope)
				}

				resourceName, action := parts[0], parts[1]
				var resource Resource
				if err := db.Where("name = ? AND action = ?", resourceName, action).First(&resource).Error; err != nil {
					return fmt.Errorf("failed to find resource %s:%s: %v", resourceName, action, err)
				}

				var perm ResourcePermission
				if err := db.Where("resource_id = ?", resource.ID).First(&perm).Error; err != nil {
					return fmt.Errorf("failed to find permission for resource %s: %v", resource.Name, err)
				}
				permissions = append(permissions, perm)
			}
		}
	}

	// Create UserPermission entries in bulk
	var userPerms []UserPermission
	for _, perm := range permissions {
		userPerms = append(userPerms, UserPermission{
			UserID:               user.ID,
			ResourcePermissionID: perm.ID,
		})
	}

	if err := db.CreateInBatches(&userPerms, 100).Error; err != nil {
		return fmt.Errorf("failed to create user permissions in bulk: %v", err)
	}

	return nil
}

func CreateSuperAdminFromEnv(db *gorm.DB, cfg *config.Config) error {
	role := UserRoleSuperAdmin

	// check if super admin already exists
	var count int64
	db.Model(&User{}).Where("role = ?", role).Count(&count)
	log.Info("Super admin count: %d", count)
	if count > 0 {
		return nil
	}

	email, ok := os.LookupEnv("SUPERADMIN_EMAIL")

	if !ok {
		return fmt.Errorf("SUPERADMIN_EMAIL not set")
	}

	password, ok := os.LookupEnv("SUPERADMIN_PASSWORD")

	if !ok {
		return fmt.Errorf("SUPERADMIN_PASSWORD not set")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	name, ok := os.LookupEnv("SUPERADMIN_NAME")

	if !ok {
		return fmt.Errorf("SUPERADMIN_NAME not set")
	}

	teamName, ok := os.LookupEnv("SUPERADMIN_TEAM_NAME")

	if !ok {
		return fmt.Errorf("SUPERADMIN_TEAM_NAME not set")
	}

	team := Team{
		Name: teamName,
	}

	if err := db.Create(&team).Error; err != nil {
		return fmt.Errorf("failed to create team: %v", err)
	}
	user := User{
		FirstName: name,
		LastName:  "",
		Email:     email,
		Role:      role,
		Password:  string(hashedPassword),
		TeamID:    team.ID,
	}

	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("failed to create superadmin user: %v", err)
	}

	smtpConfig := SMTPConfig{
		TeamID:      team.ID,
		Host:        cfg.SMTP.Host,
		Port:        cfg.SMTP.Port,
		Username:    cfg.SMTP.User,
		Password:    cfg.SMTP.Password,
		SupportsTLS: true,
		FromEmail:   cfg.SMTP.FromEmail,
	}

	if err := db.Create(&smtpConfig).Error; err != nil {
		return fmt.Errorf("failed to create smtp config: %v", err)
	}

	return nil
}
