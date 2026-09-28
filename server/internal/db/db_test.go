package db

import (
	"os"
	"strconv"
	"testing"

	"gorm.io/gorm"
	"kori/internal/config"
)

// Run only against a disposable local PostgreSQL database: Connect migrates it.
func TestMigrationsAndSchemaChanges(t *testing.T) {
	portText := os.Getenv("XEM_MIGRATION_TEST_PORT")
	if portText == "" {
		t.Skip("XEM_MIGRATION_TEST_PORT required (disposable PostgreSQL)")
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Database: config.DatabaseConfig{
		Host: "127.0.0.1", Port: port, User: "postgres", Name: "postgres", SSLMode: "disable",
	}}
	if err := Connect(cfg); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close(); DB = nil })
	if err := runMigrations(); err != nil {
		t.Fatalf("repeat migration: %v", err)
	}
	err = DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("CREATE TEMP TABLE migration_plan_probe (id integer)").Error; err != nil {
			return err
		}
		for i := 0; i < 2; i++ {
			rows, err := tx.Raw("SELECT * FROM migration_plan_probe LIMIT 1").Rows()
			if err != nil {
				return err
			}
			if err := rows.Close(); err != nil {
				return err
			}
			if i == 0 {
				if err := tx.Exec("ALTER TABLE migration_plan_probe ADD COLUMN metadata jsonb DEFAULT '{}'").Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("query after schema change: %v", err)
	}
}
