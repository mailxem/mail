package notifications

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"kori/internal/models"
)

func fixture(t *testing.T) (*Worker, models.User) {
	t.Helper()
	var dialect gorm.Dialector = sqlite.Open(":memory:")
	if dsn := os.Getenv("POSTHOOT_TEST_DATABASE_URL"); dsn != "" {
		admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		require.NoError(t, err)
		schema := "notice_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		require.NoError(t, admin.Exec("CREATE SCHEMA "+schema).Error)
		t.Cleanup(func() { admin.Exec("DROP SCHEMA " + schema + " CASCADE"); raw, _ := admin.DB(); raw.Close() })
		dialect = postgres.Open(dsn + " search_path=" + schema)
	}
	db, err := gorm.Open(dialect, &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true, Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.Team{}, &models.PasswordReset{}, &models.ServiceNotice{}, &models.ServiceNoticeEvent{}))
	raw, err := db.DB()
	require.NoError(t, err)
	if os.Getenv("POSTHOOT_TEST_DATABASE_URL") == "" {
		raw.SetMaxOpenConns(1)
	}
	t.Cleanup(func() { raw.Close() })
	user := models.User{Base: models.Base{ID: uuid.NewString()}, TeamID: uuid.NewString(), Role: models.UserRoleAdmin, Email: "owner@example.net", FirstName: "Ada"}
	team := models.Team{Base: models.Base{ID: user.TeamID}, Name: "Studio <Xem>", OwnerUserID: user.ID}
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&team).Error)
	require.NoError(t, db.Session(&gorm.Session{SkipHooks: true}).Create(&user).Error)
	w, err := New(db, "https://mail.example.net", nil)
	require.NoError(t, err)
	now := time.Date(2026, 9, 30, 10, 0, 0, 0, time.UTC)
	w.Now = func() time.Time { return now }
	return w, user
}
func queue(t *testing.T, w *Worker, fn func(*gorm.DB) error) {
	t.Helper()
	require.NoError(t, w.DB.Transaction(fn))
}
func notice(t *testing.T, w *Worker, kind string) models.ServiceNotice {
	t.Helper()
	var n models.ServiceNotice
	require.NoError(t, w.DB.Where("kind = ?", kind).First(&n).Error)
	return n
}

func TestSendingNoticesAggregateAndDeduplicateAcrossHours(t *testing.T) {
	w, u := fixture(t)
	start := w.Now()
	for _, id := range []string{"one", "one", "two"} {
		queue(t, w, func(tx *gorm.DB) error { return models.QueueSendingNotice(tx, u.TeamID, "smtp", id, "FAILED", start) })
	}
	n := notice(t, w, "sending_failed")
	require.EqualValues(t, 2, n.EventCount)
	// A delayed retry callback must not become a fresh hourly alert.
	queue(t, w, func(tx *gorm.DB) error {
		return models.QueueSendingNotice(tx, u.TeamID, "smtp", "one", "FAILED", start.Add(time.Hour))
	})
	var count int64
	require.NoError(t, w.DB.Model(&models.ServiceNotice{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
	calls := 0
	w.Send = func(_ context.Context, key, to, subject, html, text string) (bool, error) {
		calls++
		require.Equal(t, u.Email, to)
		require.Contains(t, html, "Events recorded: <strong>2")
		require.Contains(t, html, "Studio &lt;Xem&gt;")
		require.NotContains(t, html, "smtp-password")
		return false, nil
	}
	require.ErrorIs(t, w.ProcessOne(context.Background()), gorm.ErrRecordNotFound)
	w.Now = func() time.Time { return start.Add(time.Minute) }
	require.NoError(t, w.ProcessOne(context.Background()))
	require.Equal(t, 1, calls)
	require.Equal(t, "ACCEPTED", notice(t, w, "sending_failed").Status)
	queue(t, w, func(tx *gorm.DB) error {
		return models.QueueSendingNotice(tx, u.TeamID, "smtp", "three", "FAILED", start.Add(2*time.Minute))
	})
	require.ErrorIs(t, w.ProcessOne(context.Background()), gorm.ErrRecordNotFound)
	require.Equal(t, 1, calls)
}
func TestNoticesRollbackWithTheSourceEvent(t *testing.T) {
	w, u := fixture(t)
	err := w.DB.Transaction(func(tx *gorm.DB) error {
		require.NoError(t, models.QueueSendingNotice(tx, u.TeamID, "smtp", "one", "FAILED", w.Now()))
		return errors.New("rollback")
	})
	require.Error(t, err)
	var count int64
	require.NoError(t, w.DB.Model(&models.ServiceNoticeEvent{}).Count(&count).Error)
	require.Zero(t, count)
	require.NoError(t, w.DB.Model(&models.ServiceNotice{}).Count(&count).Error)
	require.Zero(t, count)
}
func TestAccountMailUsesAccountRecipientAndValidResetOnly(t *testing.T) {
	w, u := fixture(t)
	u.Role = models.UserRoleMember
	require.NoError(t, w.DB.Model(&u).Update("role", u.Role).Error)
	reset := models.PasswordReset{UserID: u.ID, Code: "secret-reset-token", ExpiresAt: w.Now().Add(15 * time.Minute)}
	require.NoError(t, w.DB.Create(&reset).Error)
	queue(t, w, func(tx *gorm.DB) error { return models.QueueAccountNotice(tx, u, "password_reset", reset.ID, w.Now()) })
	n := notice(t, w, "password_reset")
	require.NotContains(t, fmt.Sprintf("%+v", n), reset.Code)
	calls := 0
	w.Suppressed = func(context.Context, string, string) (bool, error) {
		t.Fatal("account security must not use marketing suppression")
		return true, nil
	}
	w.Send = func(_ context.Context, key, to, subject, html, text string) (bool, error) {
		calls++
		require.Equal(t, u.Email, to)
		require.Contains(t, html, "https://mail.example.net/auth/reset-password/secret-reset-token")
		require.Contains(t, html, "15 minutes")
		return false, nil
	}
	require.NoError(t, w.ProcessOne(context.Background()))
	require.Equal(t, 1, calls)
	// A second request consumed before dispatch must never be sent.
	other := models.PasswordReset{UserID: u.ID, Code: "used-token", ExpiresAt: w.Now().Add(15 * time.Minute), Used: true}
	require.NoError(t, w.DB.Create(&other).Error)
	queue(t, w, func(tx *gorm.DB) error { return models.QueueAccountNotice(tx, u, "password_reset", other.ID, w.Now()) })
	require.NoError(t, w.ProcessOne(context.Background()))
	require.Equal(t, 1, calls)
}
func TestRecipientChangesAndSuppressionSkipAlerts(t *testing.T) {
	for _, scenario := range []string{"owner removed", "owner demoted", "suppressed", "email changed", "expired reset"} {
		t.Run(scenario, func(t *testing.T) {
			w, u := fixture(t)
			start := w.Now()
			if scenario == "email changed" {
				queue(t, w, func(tx *gorm.DB) error { return models.QueueAccountNotice(tx, u, "welcome", "", start) })
				require.NoError(t, w.DB.Model(&u).Update("email", "different@example.net").Error)
			} else if scenario == "expired reset" {
				r := models.PasswordReset{UserID: u.ID, Code: "expired", ExpiresAt: start.Add(-time.Minute)}
				require.NoError(t, w.DB.Create(&r).Error)
				queue(t, w, func(tx *gorm.DB) error { return models.QueueAccountNotice(tx, u, "password_reset", r.ID, start) })
			} else {
				queue(t, w, func(tx *gorm.DB) error {
					return models.QueueSendingNotice(tx, u.TeamID, "smtp", "one", "FAILED", start)
				})
			}
			switch scenario {
			case "owner removed":
				require.NoError(t, w.DB.Model(&u).Update("is_deleted", true).Error)
			case "owner demoted":
				require.NoError(t, w.DB.Model(&u).Update("role", models.UserRoleMember).Error)
			case "suppressed":
				w.Suppressed = func(context.Context, string, string) (bool, error) { return true, nil }
			}
			w.Now = func() time.Time { return start.Add(time.Minute) }
			w.Send = func(context.Context, string, string, string, string, string) (bool, error) {
				t.Fatal("must skip")
				return false, nil
			}
			require.NoError(t, w.ProcessOne(context.Background()))
			var n models.ServiceNotice
			require.NoError(t, w.DB.First(&n).Error)
			require.Equal(t, "SKIPPED", n.Status)
		})
	}
}
func TestRetryUnknownAndCrashedClaims(t *testing.T) {
	for _, scenario := range []string{"definite failure", "uncertain", "crashed"} {
		t.Run(scenario, func(t *testing.T) {
			w, u := fixture(t)
			start := w.Now()
			queue(t, w, func(tx *gorm.DB) error { return models.QueueAccountNotice(tx, u, "welcome", "", start) })
			calls := 0
			w.Send = func(context.Context, string, string, string, string, string) (bool, error) {
				calls++
				return scenario == "uncertain", errors.New("test SMTP failure")
			}
			if scenario == "crashed" {
				require.NoError(t, w.DB.Model(&models.ServiceNotice{}).Where("kind = ?", "welcome").Updates(map[string]any{"status": "SENDING", "updated_at": start.Add(-10 * time.Minute)}).Error)
				require.ErrorIs(t, w.ProcessOne(context.Background()), gorm.ErrRecordNotFound)
			} else {
				require.NoError(t, w.ProcessOne(context.Background()))
			}
			n := notice(t, w, "welcome")
			if scenario == "definite failure" {
				require.Equal(t, "QUEUED", n.Status)
				require.True(t, n.NextAttemptAt.After(start))
				require.Equal(t, 1, calls)
			} else {
				require.Equal(t, "DELIVERY_UNKNOWN", n.Status)
				require.ErrorIs(t, w.ProcessOne(context.Background()), gorm.ErrRecordNotFound)
			}
		})
	}
}
func TestConcurrentClaimsSendOnlyOnce(t *testing.T) {
	w, u := fixture(t)
	queue(t, w, func(tx *gorm.DB) error { return models.QueueAccountNotice(tx, u, "welcome", "", w.Now()) })
	var mu sync.Mutex
	calls := 0
	w.Send = func(context.Context, string, string, string, string, string) (bool, error) {
		mu.Lock()
		defer mu.Unlock()
		calls++
		return false, nil
	}
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() { errs <- w.ProcessOne(context.Background()) }()
	}
	for i := 0; i < 2; i++ {
		err := <-errs
		require.True(t, err == nil || errors.Is(err, gorm.ErrRecordNotFound), "%v", err)
	}
	require.Equal(t, 1, calls)
}

func TestDefiniteFailuresStopAfterFiveAttempts(t *testing.T) {
	w, u := fixture(t)
	queue(t, w, func(tx *gorm.DB) error { return models.QueueAccountNotice(tx, u, "welcome", "", w.Now()) })
	calls := 0
	w.Send = func(context.Context, string, string, string, string, string) (bool, error) {
		calls++
		return false, errors.New("connect failed")
	}
	for attempt := 1; attempt <= 5; attempt++ {
		require.NoError(t, w.ProcessOne(context.Background()))
		n := notice(t, w, "welcome")
		require.Equal(t, attempt, n.Attempts)
		if attempt < 5 {
			require.Equal(t, "QUEUED", n.Status)
			w.Now = func() time.Time { return n.NextAttemptAt }
		} else {
			require.Equal(t, "FAILED", n.Status)
		}
	}
	require.ErrorIs(t, w.ProcessOne(context.Background()), gorm.ErrRecordNotFound)
	require.Equal(t, 5, calls)
}
