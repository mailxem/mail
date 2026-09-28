package models

import (
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

func TestCloudflareMarketingRejectedAtPersistenceBoundary(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TABLE smtp_configs (id TEXT, team_id TEXT, provider TEXT)").Error)
	require.NoError(t, db.Exec("CREATE TABLE email_categories (id TEXT, team_id TEXT, name TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO smtp_configs VALUES ('cf', 'team', 'CLOUDFLARE'), ('custom', 'team', 'CUSTOM')").Error)
	require.NoError(t, db.Exec("INSERT INTO email_categories VALUES ('transaction', 'team', 'Transactional'), ('marketing', 'team', 'Marketing')").Error)
	require.Error(t, rejectCloudflareMarketing(db, "team", "cf", true, "transaction"))
	require.Error(t, rejectCloudflareMarketing(db, "team", "cf", false, "marketing"))
	require.Error(t, rejectCloudflareMarketing(db, "team", "cf", false, "unknown"))
	require.NoError(t, rejectCloudflareMarketing(db, "team", "cf", false, "transaction"))
	require.NoError(t, rejectCloudflareMarketing(db, "team", "custom", true, "marketing"))
}
