package hygiene

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NeverEmailEntry is an address that hard-bounced for any workspace. Only a
// keyed hash is stored, never the address.
type NeverEmailEntry struct {
	Hash      string    `gorm:"primaryKey;size:64"`
	ExpiresAt time.Time `gorm:"not null;index"`
}

func (NeverEmailEntry) TableName() string { return "global_suppressions" }

// NeverEmailFor is how long a hard bounce keeps an address off every workspace.
const NeverEmailFor = 180 * 24 * time.Hour

// hash is "" when SUPPRESSION_HASH_KEY is unset, which turns the list off.
func hash(email string) string {
	key := os.Getenv("SUPPRESSION_HASH_KEY")
	if key == "" {
		return ""
	}
	m := hmac.New(sha256.New, []byte(key))
	m.Write([]byte(strings.ToLower(strings.TrimSpace(email))))
	return hex.EncodeToString(m.Sum(nil))
}

// Block adds email to the shared list, or pushes its expiry out again.
func Block(tx *gorm.DB, email string, now time.Time) error {
	h := hash(email)
	if h == "" {
		return nil
	}
	return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "hash"}}, DoUpdates: clause.AssignmentColumns([]string{"expires_at"})}).
		Create(&NeverEmailEntry{Hash: h, ExpiresAt: now.Add(NeverEmailFor)}).Error
}

// Blocked returns the given addresses that are on the shared list, keyed as passed.
func Blocked(tx *gorm.DB, emails ...string) (map[string]bool, error) {
	out := map[string]bool{}
	if os.Getenv("SUPPRESSION_HASH_KEY") == "" {
		return out, nil
	}
	byHash := map[string][]string{}
	hashes := make([]string, 0, len(emails))
	for _, e := range emails {
		h := hash(e)
		if byHash[h] == nil {
			hashes = append(hashes, h)
		}
		byHash[h] = append(byHash[h], e)
	}
	for start := 0; start < len(hashes); start += 500 {
		var found []string
		if err := tx.Model(&NeverEmailEntry{}).Where("hash IN ? AND expires_at > ?", hashes[start:min(start+500, len(hashes))], time.Now()).Pluck("hash", &found).Error; err != nil {
			return nil, err
		}
		for _, h := range found {
			for _, e := range byHash[h] {
				out[e] = true
			}
		}
	}
	return out, nil
}
