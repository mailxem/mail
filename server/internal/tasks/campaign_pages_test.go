package tasks

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/models"
)

// pageDatabases runs fn on SQLite and, when POSTHOOT_TEST_DATABASE_URL is
// set, on PostgreSQL in a throwaway schema.
func pageDatabases(t *testing.T, fn func(t *testing.T, db *gorm.DB)) {
	open := map[string]func(t *testing.T) *gorm.DB{"sqlite": func(t *testing.T) *gorm.DB {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := db.DB()
		raw.SetMaxOpenConns(1) // one in-memory database
		return db
	}}
	if dsn := os.Getenv("POSTHOOT_TEST_DATABASE_URL"); dsn != "" {
		open["postgres"] = func(t *testing.T) *gorm.DB {
			admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
			if err != nil {
				t.Fatal(err)
			}
			schema := "pages_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
			if err := admin.Exec("CREATE SCHEMA " + schema).Error; err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { admin.Exec("DROP SCHEMA " + schema + " CASCADE"); raw, _ := admin.DB(); raw.Close() })
			db, err := gorm.Open(postgres.Open(dsn+" search_path="+schema), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, Logger: logger.Default.LogMode(logger.Silent)})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { raw, _ := db.DB(); raw.Close() })
			return db
		}
	}
	for name, o := range open {
		t.Run(name, func(t *testing.T) {
			db := o(t)
			if err := db.AutoMigrate(&models.Contact{}, &models.Email{}, &models.SMTPConfig{}); err != nil {
				t.Fatal(err)
			}
			// The index internal/db adds in production.
			if err := db.Exec("CREATE INDEX emails_campaign_contact ON emails (campaign_id, contact_id) WHERE campaign_id IS NOT NULL").Error; err != nil {
				t.Fatal(err)
			}
			fn(t, db)
		})
	}
}

type pageFixture struct {
	db                         *gorm.DB
	team, list, campaign, smtp string
	inserts                    int
}

func newPageFixture(t *testing.T, db *gorm.DB, active int) *pageFixture {
	f := &pageFixture{db: db, team: uuid.NewString(), list: uuid.NewString(), campaign: uuid.NewString(), smtp: uuid.NewString()}
	var contacts []models.Contact
	add := func(list string, status models.SubscriberStatus) {
		contacts = append(contacts, models.Contact{Base: models.Base{ID: uuid.NewString()}, Email: uuid.NewString() + "@example.com", TeamID: f.team, ListID: list, Status: status})
	}
	for range active {
		add(f.list, models.SubscriberStatusActive)
	}
	for range 7 {
		add(f.list, models.SubscriberStatusUnsubscribed)
		add(uuid.NewString(), models.SubscriberStatusActive) // another list
	}
	if err := db.Session(&gorm.Session{SkipHooks: true}).CreateInBatches(contacts, 500).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Callback().Create().After("gorm:create").Register("count_email_inserts", func(tx *gorm.DB) {
		if tx.Statement.Table == "emails" {
			f.inserts++
		}
	}); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *pageFixture) build(c models.Contact) (*models.Email, error) {
	return &models.Email{From: "a@example.com", To: c.Email, Subject: "s", Body: "b", Status: models.EmailStatusPending,
		TeamID: f.team, ContactID: c.ID, SMTPConfigID: f.smtp, CategoryID: f.team, CampaignID: f.campaign}, nil
}

// run sends with pageSize and limit, recording page sizes; failOnPage makes
// that page's send fail after its emails were saved, like a crashed worker.
func (f *pageFixture) run(t *testing.T, pageSize, limit, failOnPage int) (pages []int, sent int, err error) {
	// The contacts a campaign may still email, as marketing.Recipients selects them.
	eligible := f.db.Model(&models.Contact{}).Where("contacts.list_id = ? AND contacts.status = ?", f.list, models.SubscriberStatusActive).
		Where("NOT EXISTS (SELECT 1 FROM emails WHERE emails.campaign_id = ? AND emails.contact_id = contacts.id)", f.campaign).Session(&gorm.Session{})
	sent, err = sendCampaignPages(f.db, eligible, false, pageSize, limit, f.build, func(emails []*models.Email) error {
		pages = append(pages, len(emails))
		if len(pages) == failOnPage {
			return errors.New("worker stopped")
		}
		return nil
	})
	return
}

// emailed returns how many emails each contact has for the campaign.
func (f *pageFixture) emailed(t *testing.T) (contacts int, perContact map[string]int) {
	var rows []struct {
		ContactID string
		N         int
	}
	if err := f.db.Model(&models.Email{}).Select("contact_id, count(*) AS n").Where("campaign_id = ?", f.campaign).Group("contact_id").Scan(&rows).Error; err != nil {
		t.Fatal(err)
	}
	perContact = map[string]int{}
	for _, r := range rows {
		perContact[r.ContactID] = r.N
	}
	return len(rows), perContact
}

func (f *pageFixture) checkEveryoneOnce(t *testing.T, active int) {
	n, per := f.emailed(t)
	if n != active {
		t.Fatalf("emailed %d contacts, want %d", n, active)
	}
	for id, c := range per {
		if c != 1 {
			t.Fatalf("contact %s got %d emails", id, c)
		}
	}
	var wrong int64
	f.db.Model(&models.Contact{}).Where("id IN (?) AND (list_id <> ? OR status <> ?)", f.db.Model(&models.Email{}).Select("contact_id").Where("campaign_id = ?", f.campaign), f.list, models.SubscriberStatusActive).Count(&wrong)
	if wrong != 0 {
		t.Fatalf("%d emails went to contacts outside the list or not active", wrong)
	}
}

func TestCampaignPagesEmailEveryContactOnceAcrossRetry(t *testing.T) {
	pageDatabases(t, func(t *testing.T, db *gorm.DB) {
		f := newPageFixture(t, db, 2500)
		pages, sent, err := f.run(t, 1200, -1, 2)
		if err == nil || sent != 2400 || len(pages) != 2 {
			t.Fatalf("first run: pages %v, sent %d, err %v; want a failure after 2 pages", pages, sent, err)
		}
		pages, sent, err = f.run(t, 1200, -1, 0) // the retried task
		if err != nil || sent != 100 || len(pages) != 1 || pages[0] != 100 {
			t.Fatalf("retry: pages %v, sent %d, err %v; want only the last 100", pages, sent, err)
		}
		f.checkEveryoneOnce(t, 2500)
		if f.inserts != 7 { // 3 per full page of 1200, 1 for the last 100
			t.Fatalf("%d INSERT statements, want 7", f.inserts)
		}
		if pages, sent, _ := f.run(t, 1200, -1, 0); sent != 0 || len(pages) != 0 {
			t.Fatalf("finished campaign sent again: %v", pages)
		}
	})
}

func TestCampaignPagesRespectStageLimits(t *testing.T) {
	pageDatabases(t, func(t *testing.T, db *gorm.DB) {
		f := newPageFixture(t, db, 2500)
		for _, stage := range []struct {
			limit int
			pages []int
		}{
			{125, []int{125}},               // under one page
			{1000, []int{400, 400, 200}},    // across pages, last one cut to the limit
			{-1, []int{400, 400, 400, 175}}, // final stage: everyone left
		} {
			pages, sent, err := f.run(t, 400, stage.limit, 0)
			want := 0
			for _, p := range stage.pages {
				want += p
			}
			if err != nil || sent != want || len(pages) != len(stage.pages) {
				t.Fatalf("limit %d: pages %v, sent %d, err %v; want %v", stage.limit, pages, sent, err, stage.pages)
			}
			for i := range pages {
				if pages[i] != stage.pages[i] {
					t.Fatalf("limit %d: pages %v, want %v", stage.limit, pages, stage.pages)
				}
			}
		}
		f.checkEveryoneOnce(t, 2500)
	})
}
