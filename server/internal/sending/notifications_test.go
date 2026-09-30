package sending

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kori/internal/models"
)

func notificationOwner(t *testing.T, s *Service, team string) models.User {
	t.Helper()
	require.NoError(t, s.DB.AutoMigrate(&models.Team{}, &models.User{}))
	owner := models.User{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Email: "owner@example.net", Role: models.UserRoleAdmin}
	row := models.Team{Base: models.Base{ID: team}, Name: "Workspace", OwnerUserID: owner.ID}
	require.NoError(t, s.DB.Session(&gorm.Session{SkipHooks: true}).Create(&row).Error)
	require.NoError(t, s.DB.Session(&gorm.Session{SkipHooks: true}).Create(&owner).Error)
	s.NotificationURL = "https://app.xem.email/settings/sending"
	return owner
}
func TestMilestonesAreAtomicAndDeduplicated(t *testing.T) {
	s, team, d, _ := setup(t)
	for i := 0; i < 3; i++ {
		_, err := s.RefreshDomain(context.Background(), team, d.ID)
		require.NoError(t, err)
	}
	var count int64
	require.NoError(t, s.DB.Model(&MilestoneEmail{}).Where("team_id = ? AND kind = ?", team, "domain_ready").Count(&count).Error)
	require.EqualValues(t, 1, count)
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		require.NoError(t, QueueMilestone(tx, team, d.ID, "sender_connected"))
		return errors.New("rollback")
	})
	require.Error(t, err)
	require.NoError(t, s.DB.Model(&MilestoneEmail{}).Where("kind = ?", "sender_connected").Count(&count).Error)
	require.Zero(t, count)
}
func TestMilestoneDispatchUsesOwnerAndQuarantinesUnknown(t *testing.T) {
	s, team, _, _ := setup(t)
	owner := notificationOwner(t, s, team)
	other := models.User{Base: models.Base{ID: uuid.NewString()}, TeamID: team, Email: "other@example.net", Role: models.UserRoleAdmin}
	require.NoError(t, s.DB.Session(&gorm.Session{SkipHooks: true}).Create(&other).Error)
	calls := 0
	s.Notify = func(_ context.Context, key, to, subject, html, text string) (bool, error) {
		calls++
		require.Equal(t, owner.Email, to)
		require.Contains(t, html, "Open managed sending")
		require.NotContains(t, html, "xem-managed=")
		return true, errors.New("lost acknowledgement")
	}
	require.NoError(t, s.ProcessMilestone(context.Background()))
	var n MilestoneEmail
	require.NoError(t, s.DB.Where("status = ?", "DELIVERY_UNKNOWN").First(&n).Error)
	require.Equal(t, 1, calls)
	require.NoError(t, s.DB.Model(&MilestoneEmail{}).Where("key <> ?", n.Key).Update("status", "SKIPPED").Error)
	require.ErrorIs(t, s.ProcessMilestone(context.Background()), gorm.ErrRecordNotFound)
	require.Equal(t, 1, calls)
}
func TestMilestoneRetriesAndRechecksMembership(t *testing.T) {
	s, team, _, _ := setup(t)
	owner := notificationOwner(t, s, team)
	calls := 0
	s.Notify = func(context.Context, string, string, string, string, string) (bool, error) {
		calls++
		return false, errors.New("connection refused before DATA")
	}
	require.NoError(t, s.ProcessMilestone(context.Background()))
	var n MilestoneEmail
	require.NoError(t, s.DB.Where("attempts = 1").First(&n).Error)
	require.Equal(t, "QUEUED", n.Status)
	require.True(t, n.NextAttemptAt.After(s.Now()))
	require.NoError(t, s.DB.Model(&MilestoneEmail{}).Where("key <> ?", n.Key).Update("status", "SKIPPED").Error)
	require.NoError(t, s.DB.Model(&owner).Update("is_deleted", true).Error)
	later := s.Now().Add(time.Hour)
	s.Now = func() time.Time { return later }
	require.NoError(t, s.ProcessMilestone(context.Background()))
	require.Equal(t, 1, calls)
	require.NoError(t, s.DB.First(&n, "key = ?", n.Key).Error)
	require.Equal(t, "SKIPPED", n.Status)
}
func TestStaleMilestoneClaimIsNotResent(t *testing.T) {
	s, team, _, _ := setup(t)
	notificationOwner(t, s, team)
	require.NoError(t, s.DB.Model(&MilestoneEmail{}).Where("team_id = ?", team).Updates(map[string]any{"status": "SENDING", "updated_at": s.Now().Add(-10 * time.Minute)}).Error)
	s.Notify = func(context.Context, string, string, string, string, string) (bool, error) {
		t.Fatal("must not resend")
		return false, nil
	}
	require.ErrorIs(t, s.ProcessMilestone(context.Background()), gorm.ErrRecordNotFound)
	var count int64
	require.NoError(t, s.DB.Model(&MilestoneEmail{}).Where("status = ?", "DELIVERY_UNKNOWN").Count(&count).Error)
	require.Greater(t, count, int64(0))
}

func TestMilestoneMigrationBaselinesExistingWorkspaces(t *testing.T) {
	s, team, d, _ := setup(t)
	require.NoError(t, s.DB.Migrator().DropTable(&MilestoneEmail{}))
	require.NoError(t, Migrate(s.DB))
	var n MilestoneEmail
	require.NoError(t, s.DB.Where("team_id = ? AND kind = ?", team, "domain_ready").First(&n).Error)
	require.Equal(t, "SKIPPED", n.Status)
	_, err := s.RefreshDomain(context.Background(), team, d.ID)
	require.NoError(t, err)
	require.NoError(t, s.DB.First(&n, "key = ?", n.Key).Error)
	require.Equal(t, "SKIPPED", n.Status)
	require.NoError(t, QueueMilestone(s.DB, team, d.ID, "test_queued"))
	require.NoError(t, Migrate(s.DB))
	var test MilestoneEmail
	require.NoError(t, s.DB.Where("team_id = ? AND kind = ?", team, "test_queued").First(&test).Error)
	require.Equal(t, "QUEUED", test.Status)
}
