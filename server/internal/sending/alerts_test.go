package sending

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kori/internal/models"
)

func enableAlerts(t *testing.T, s *Service) {
	t.Helper()
	s.NotifyAlerts = true
	require.NoError(t, s.DB.AutoMigrate(&models.ServiceNotice{}, &models.ServiceNoticeEvent{}))
}
func TestSendingFailureAndUnknownAlertHooks(t *testing.T) {
	s, team, d, p := setup(t)
	enableAlerts(t, s)
	p.err = errors.New("private-token-details-must-not-be-emailed")
	m, err := s.Submit(context.Background(), input(team, d))
	require.NoError(t, err)
	require.NoError(t, s.ProcessOne(context.Background()))
	var n models.ServiceNotice
	require.NoError(t, s.DB.Where("kind = ?", "sending_unknown").First(&n).Error)
	require.Equal(t, team, n.TeamID)
	require.EqualValues(t, 1, n.EventCount)
	// Maintenance records worker crashes, with the same event key as dispatch.
	require.NoError(t, s.DB.Model(&Message{}).Where("id = ?", m.ID).Updates(map[string]any{"status": "SENDING", "updated_at": s.Now().Add(-10 * time.Minute)}).Error)
	require.NoError(t, s.MaintainMessages(context.Background()))
	require.NoError(t, s.DB.First(&n, "key = ?", n.Key).Error)
	require.EqualValues(t, 1, n.EventCount)
}
func TestFeedbackQueuesDeduplicatedAlertsAndSuspension(t *testing.T) {
	for _, kind := range []string{"Bounce", "Complaint", "Reject", "DeliveryDelay"} {
		t.Run(kind, func(t *testing.T) {
			s, team, d, _ := setup(t)
			enableAlerts(t, s)
			m := reputationMessage(t, s, team, d, []string{"reader@example.net"})
			raw := reputationFeedback(t, m, kind, "Permanent", "reader@example.net")
			require.NoError(t, s.ApplyFeedback(context.Background(), "event-one", raw))
			require.NoError(t, s.ApplyFeedback(context.Background(), "event-one", raw))
			require.NoError(t, s.ApplyFeedback(context.Background(), "event-two", raw))
			var n models.ServiceNotice
			noticeKind := map[string]string{"Bounce": "sending_bounced", "Complaint": "sending_complaint", "Reject": "sending_failed", "DeliveryDelay": "sending_delayed"}[kind]
			require.NoError(t, s.DB.Where("kind = ?", noticeKind).First(&n).Error)
			require.EqualValues(t, 1, n.EventCount)
			if kind == "Complaint" {
				var suspended models.ServiceNotice
				require.NoError(t, s.DB.Where("kind = ?", "sending_suspended").First(&suspended).Error)
			}
		})
	}
}
func TestTestMessagesKeepTheirMilestoneWithoutDuplicateAlert(t *testing.T) {
	s, team, d, p := setup(t)
	enableAlerts(t, s)
	p.err = errors.New("unknown")
	in := input(team, d)
	in.IsTest = true
	_, err := s.Submit(context.Background(), in)
	require.NoError(t, err)
	require.NoError(t, s.ProcessOne(context.Background()))
	var n int64
	require.NoError(t, s.DB.Model(&models.ServiceNotice{}).Count(&n).Error)
	require.Zero(t, n)
	var milestone MilestoneEmail
	require.NoError(t, s.DB.Where("kind = ?", "test_attention").First(&milestone).Error)
}
func TestAlertFailureRollsBackManagedStatus(t *testing.T) {
	s, team, d, p := setup(t)
	enableAlerts(t, s)
	p.err = errors.New("unknown")
	m, err := s.Submit(context.Background(), input(team, d))
	require.NoError(t, err)
	require.NoError(t, s.DB.Migrator().DropTable(&models.ServiceNoticeEvent{}))
	require.Error(t, s.ProcessOne(context.Background()))
	require.NoError(t, s.DB.Session(&gorm.Session{}).First(&m, "id = ?", m.ID).Error)
	require.Equal(t, "SENDING", m.Status)
}
