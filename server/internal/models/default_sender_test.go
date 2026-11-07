package models

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDefaultSenderFallsBackToTheActiveOne(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TABLE smtp_configs (id TEXT, team_id TEXT, provider TEXT, is_default BOOLEAN, is_active BOOLEAN, is_deleted BOOLEAN, created_at DATETIME)").Error)
	// A managed sender is active but never marked default. (MANAGED rows skip
	// password decryption, which needs keys this test does not load.)
	require.NoError(t, db.Exec("INSERT INTO smtp_configs VALUES ('old', 'team', 'MANAGED', false, false, false, '2026-01-01'), ('managed', 'team', 'MANAGED', false, true, false, '2026-02-01')").Error)
	got, err := GetSMTPConfig("team", "", "", db)
	require.NoError(t, err)
	require.Equal(t, "managed", got.ID)

	require.NoError(t, db.Exec("INSERT INTO smtp_configs VALUES ('chosen', 'team', 'MANAGED', true, true, false, '2026-03-01')").Error)
	got, err = GetSMTPConfig("team", "", "", db)
	require.NoError(t, err)
	require.Equal(t, "chosen", got.ID)

	_, err = GetSMTPConfig("empty", "", "", db)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
