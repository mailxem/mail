package middleware

import (
	"context"
	"fmt"
	"kori/internal/models"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Permission scopes
const (
	ScopeAdmin = "admin"
	ScopeRead  = "read"
	ScopeWrite = "create"
)

// ValidateAPIKeyPermissions validates if an API key has the required permissions
func ValidateAPIKeyPermissions(ctx context.Context, db *gorm.DB, apiKeyID string, requiredPermissions []string) error {
	var permissions []models.APIKeyPermission

	log.Info("Validating API key permissions for %s", apiKeyID)
	log.Info("Required permissions %v", requiredPermissions)

	// Get API key permissions with resource permission details
	err := db.WithContext(ctx).Where("key_id = ? AND is_deleted = false", apiKeyID).
		Preload("ResourcePermission", "is_deleted = false").
		Preload("ResourcePermission.Resource", "is_deleted = false").
		Find(&permissions).Error
	if err != nil {
		return fmt.Errorf("failed to get permissions: %w", err)
	}

	// Check each required permission
	for _, required := range requiredPermissions {
		hasPermission := false
		requiredParts := strings.Split(required, ":")
		if len(requiredParts) != 2 {
			continue // Invalid permission format
		}

		requiredResource := requiredParts[0]
		requiredScope := requiredParts[1]
		// Legacy routes call the create scope "write".
		if requiredScope == "write" {
			requiredScope = ScopeWrite
		}

		for _, perm := range permissions {
			if perm.ResourcePermission == nil || perm.ResourcePermission.Resource == nil {
				continue
			}

			resource := perm.ResourcePermission.Resource

			// Check if the permission matches
			if resource.Name == requiredResource {
				switch resource.Action {
				case ScopeAdmin:
					hasPermission = true
				case ScopeWrite:
					hasPermission = requiredScope == ScopeWrite || requiredScope == ScopeRead
				case ScopeRead:
					hasPermission = requiredScope == ScopeRead
				}
				if hasPermission {
					break
				}
			}
		}

		if !hasPermission {
			return echo.NewHTTPError(http.StatusForbidden, fmt.Sprintf("missing required permission: %s", required))
		}
	}

	return nil
}

// ValidateMethodPermission validates if a given scope allows a specific HTTP method
func ValidateMethodPermission(method string, scope string) bool {
	switch scope {
	case ScopeAdmin:
		return true
	case ScopeWrite:
		return method == http.MethodPost || method == http.MethodPut ||
			method == http.MethodDelete || method == http.MethodPatch
	case ScopeRead:
		return method == http.MethodGet
	default:
		return false
	}
}

// GetRequiredPermissionForMethod returns the required permission scope for a given HTTP method
func GetRequiredPermissionForMethod(method string) string {
	switch method {
	case http.MethodGet:
		return ScopeRead
	case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		return ScopeWrite
	default:
		return ""
	}
}

// RequirePermissions middleware checks if the user/API key has the required permissions
func RequirePermissions(db *gorm.DB, requiredPermissions ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if user has admin access first
			if hasAdmin, ok := c.Get("hasAdminAccess").(bool); ok && hasAdmin {
				return next(c)
			}

			isAPIKey := c.Get("isAPIKey").(bool)
			method := c.Request().Method

			if isAPIKey {
				apiKeyID := c.Get("apiKeyID").(string)
				if err := ValidateAPIKeyPermissions(c.Request().Context(), db, apiKeyID, requiredPermissions); err != nil {
					return err
				}
			} else {
				// For JWT auth, check role-based permissions
				role := c.Get("role").(string)
				scopes := c.Get("scopes").([]string)

				// Admin role has all permissions
				if role == "admin" {
					return next(c)
				}

				// Check if user has any of the required permissions
				hasPermission := false
				for _, scope := range scopes {
					if ValidateMethodPermission(method, scope) {
						hasPermission = true
						break
					}
				}

				if !hasPermission {
					return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
				}
			}

			return next(c)
		}
	}
}
