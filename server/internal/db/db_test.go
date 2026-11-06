package db

import (
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"kori/internal/config"
	"kori/internal/models"
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
	// Upgrading to list cleanup starts existing readers' count at the upgrade, once.
	team, list, contact := uuid.NewString(), uuid.NewString(), uuid.NewString()
	raw := DB.Session(&gorm.Session{SkipHooks: true})
	if err := errors.Join(
		raw.Create(&models.Team{Base: models.Base{ID: team}, Name: "Upgrade"}).Error,
		raw.Create(&models.MailingList{Base: models.Base{ID: list}, TeamID: team, Name: "Old"}).Error,
		raw.Create(&models.Contact{Base: models.Base{ID: contact}, TeamID: team, ListID: list, Email: contact + "@example.com"}).Error,
		DB.Exec("ALTER TABLE contacts DROP COLUMN cleanup_since").Error,
		runMigrations(),
	); err != nil {
		t.Fatal(err)
	}
	since := func() *time.Time {
		var c models.Contact
		if err := DB.First(&c, "id = ?", contact).Error; err != nil {
			t.Fatal(err)
		}
		return c.CleanupSince
	}
	if s := since(); s == nil || time.Since(*s) > time.Minute {
		t.Fatalf("existing contact not started at the upgrade: %v", s)
	}
	old := time.Now().Add(-90 * 24 * time.Hour).UTC()
	if err := errors.Join(DB.Model(&models.Contact{}).Where("id = ?", contact).UpdateColumn("cleanup_since", old).Error, runMigrations()); err != nil {
		t.Fatal(err)
	}
	if s := since(); s == nil || !s.Equal(old.Truncate(time.Microsecond)) {
		t.Fatalf("later migration moved cleanup_since: %v", s)
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
