package db

import (
	"kori/internal/config"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestDatabaseDSNPreservesEmptyAndQuotedCredentials(t *testing.T) {
	for _, password := range []string{"", "simple", "space in password", "quote'and\\slash", "dbname=other port=9999"} {
		cfg := config.DatabaseConfig{Host: "127.0.0.1", Port: 5432, User: "local user", Password: password, Name: "xem_mailcheck_database", SSLMode: "disable"}
		parsed, err := pgx.ParseConfig(databaseDSN(cfg))
		require.NoError(t, err)
		require.Equal(t, cfg.Name, parsed.Database)
		require.Equal(t, cfg.User, parsed.User)
		require.Equal(t, cfg.Password, parsed.Password)
		require.Equal(t, cfg.Host, parsed.Host)
		require.EqualValues(t, cfg.Port, parsed.Port)
	}
}

func TestPoolSettingsDefaultsAndOverrides(t *testing.T) {
	t.Setenv("DB_MAX_OPEN_CONNS", "")
	t.Setenv("DB_MAX_IDLE_CONNS", "")
	t.Setenv("DB_CONN_MAX_LIFETIME", "")
	open, idle, lifetime, err := poolSettings()
	require.NoError(t, err)
	require.Equal(t, []any{50, 50, 30 * time.Minute}, []any{open, idle, lifetime})
	t.Setenv("DB_MAX_OPEN_CONNS", "20")
	t.Setenv("DB_MAX_IDLE_CONNS", "8")
	t.Setenv("DB_CONN_MAX_LIFETIME", "5m")
	open, idle, lifetime, err = poolSettings()
	require.NoError(t, err)
	require.Equal(t, []any{20, 8, 5 * time.Minute}, []any{open, idle, lifetime})
	t.Setenv("DB_MAX_OPEN_CONNS", "0")
	_, _, _, err = poolSettings()
	require.Error(t, err)
}
