package assistant

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"gorm.io/gorm"
	"kori/internal/models"
)

const Prefix = "xem_bot_"

// Lookup keeps assistant credentials irreversible at rest. Legacy customer keys
// retain their existing representation.
func Lookup(key string) string {
	if strings.HasPrefix(key, "bot_sha256:") {
		return "invalid-assistant-key"
	}
	if !strings.HasPrefix(key, Prefix) {
		return key
	}
	sum := sha256.Sum256([]byte(key))
	return "bot_sha256:" + hex.EncodeToString(sum[:])
}

func ValidateActor(db *gorm.DB, key *models.APIKey) bool {
	if key.AssistantUserID == "" {
		return !strings.HasPrefix(key.Key, "bot_sha256:")
	}
	var user models.User
	if db.Where("id = ? AND team_id = ? AND is_deleted = false", key.AssistantUserID, key.TeamID).First(&user).Error != nil {
		return false
	}
	return !key.AssistantWrite || user.Role == models.UserRoleAdmin || user.Role == models.UserRoleSuperAdmin
}

// Issue grants only product resources. Keys cannot administer other credentials,
// users, billing, or operator approvals. Every MCP call still checks permissions.
func Issue(db *gorm.DB, user models.User) (string, time.Time, error) {
	secret, expiry, _, err := IssueKey(db, user, "Xem assistant (short-lived)", 5*time.Minute)
	return secret, expiry, err
}

// IssueKey is Issue with a name and lifetime, for apps a user connected over
// OAuth. It also returns the key's ID so the caller can revoke it.
func IssueKey(db *gorm.DB, user models.User, name string, ttl time.Duration) (string, time.Time, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, "", err
	}
	secret := Prefix + hex.EncodeToString(raw)
	expiry := time.Now().UTC().Add(ttl)
	write := user.Role == models.UserRoleAdmin || user.Role == models.UserRoleSuperAdmin
	key := models.APIKey{Name: name, Key: Lookup(secret), TeamID: user.TeamID, ExpiresAt: expiry, AssistantUserID: user.ID, AssistantWrite: write}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&key).Error; err != nil {
			return err
		}
		action := "read"
		if write {
			action = "admin"
		}
		for _, name := range models.AdminKeyResources {
			var resource models.Resource
			if err := tx.Where("name = ? AND action = ? AND is_deleted = false", name, action).FirstOrCreate(&resource, models.Resource{Name: name, Action: action}).Error; err != nil {
				return err
			}
			var permission models.ResourcePermission
			if err := tx.Where("resource_id = ? AND is_deleted = false", resource.ID).FirstOrCreate(&permission, models.ResourcePermission{ResourceID: resource.ID, Scope: action + ":" + name}).Error; err != nil {
				return err
			}
			if err := tx.Create(&models.APIKeyPermission{KeyID: key.ID, ResourcePermissionID: permission.ID}).Error; err != nil {
				return err
			}
		}
		// Keep audit usage, but remove expired credential grants so idle workspaces do
		// not accumulate one active key per chat turn.
		return tx.Model(&models.APIKey{}).Where("assistant_user_id = ? AND expires_at < ?", user.ID, time.Now().UTC()).Update("is_deleted", true).Error
	})
	return secret, expiry, key.ID, err
}
