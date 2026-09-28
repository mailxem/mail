package sending

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func reputationMessage(t *testing.T, s *Service, team string, d Domain, recipients []string) Message {
	t.Helper()
	in := input(team, d)
	in.Recipients = recipients
	m, err := s.Submit(context.Background(), in)
	require.NoError(t, err)
	require.NoError(t, s.ProcessOne(context.Background()))
	return m
}

func reputationFeedback(t *testing.T, m Message, kind, bounceType string, recipients ...string) []byte {
	t.Helper()
	addresses := []map[string]string{}
	for _, r := range recipients {
		addresses = append(addresses, map[string]string{"emailAddress": r})
	}
	raw, err := json.Marshal(map[string]any{
		"eventType": kind,
		"mail":      map[string]any{"messageId": "ses-message", "source": m.From, "tags": map[string][]string{"xem_message": {m.ID}, "ses:configuration-set": {tenant(m.TeamID)}}},
		"bounce":    map[string]any{"bounceType": bounceType, "bouncedRecipients": addresses},
		"complaint": map[string]any{"complainedRecipients": addresses},
	})
	require.NoError(t, err)
	return raw
}

func TestComplaintSuspendsWorkspaceAndCannotBeClearedByDNS(t *testing.T) {
	s, team, d, p := setup(t)
	ctx := context.Background()
	m := reputationMessage(t, s, team, d, []string{"reader@example.net"})
	_, err := s.Submit(ctx, input(team, d)) // queued before the complaint
	require.NoError(t, err)
	raw := reputationFeedback(t, m, "Complaint", "", "reader@example.net")
	require.NoError(t, s.ApplyFeedback(ctx, "complaint-event", raw))
	require.NoError(t, s.ApplyFeedback(ctx, "complaint-event", raw))
	a, err := s.Account(ctx, team)
	require.NoError(t, err)
	require.True(t, a.Suspended)
	require.EqualValues(t, 2, a.DailyUsed, "suspension must not reset usage")
	_, err = s.Submit(ctx, input(team, d))
	require.ErrorIs(t, err, ErrDenied)
	_, _, err = s.CreateCredential(ctx, team, d.ID, "after suspension")
	require.ErrorIs(t, err, ErrDenied)
	require.NoError(t, s.ProcessOne(ctx))
	require.Equal(t, 1, p.sends, "queued mail must not dispatch after suspension")
	_, err = s.RefreshDomain(ctx, team, d.ID)
	require.ErrorIs(t, err, ErrDenied)
	a, err = s.Account(ctx, team)
	require.NoError(t, err)
	require.True(t, a.Suspended)
	var count int64
	require.NoError(t, s.DB.Model(&Audit{}).Where("team_id = ? AND action = ?", team, "auto_suspend_complaint").Count(&count).Error)
	require.EqualValues(t, 1, count)
	other, err := s.Account(ctx, uuid.NewString())
	require.NoError(t, err)
	require.False(t, other.Suspended)
}

func TestHardBounceThresholdCountsDistinctRecentRecipients(t *testing.T) {
	s, team, d, _ := setup(t)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		require.NoError(t, s.DB.Create(&Suppression{TeamID: team, Address: uuid.NewString() + "@old.test", Reason: "Bounce", CreatedAt: s.Now().Add(-25 * time.Hour)}).Error)
		require.NoError(t, s.DB.Create(&Suppression{TeamID: uuid.NewString(), Address: uuid.NewString() + "@other.test", Reason: "Bounce", CreatedAt: s.Now()}).Error)
	}
	recipients := []string{"one@example.net", "two@example.net", "three@example.net", "four@example.net", "five@example.net"}
	m := reputationMessage(t, s, team, d, recipients)
	require.NoError(t, s.ApplyFeedback(ctx, "soft", reputationFeedback(t, m, "Bounce", "Transient", recipients...)))
	a, err := s.Account(ctx, team)
	require.NoError(t, err)
	require.False(t, a.Suspended, "soft bounces must not count")
	for i, recipient := range recipients {
		raw := reputationFeedback(t, m, "Bounce", "Permanent", recipient, recipient)
		require.NoError(t, s.ApplyFeedback(ctx, recipient, raw))
		require.NoError(t, s.ApplyFeedback(ctx, recipient, raw))
		require.NoError(t, s.ApplyFeedback(ctx, recipient+"-duplicate", raw))
		a, err = s.Account(ctx, team)
		require.NoError(t, err)
		require.Equal(t, i == 4, a.Suspended)
	}
	var count int64
	require.NoError(t, s.DB.Model(&Audit{}).Where("team_id = ? AND action = ?", team, "auto_suspend_bounces").Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func TestUnmatchedComplaintCannotSuspendOrSuppress(t *testing.T) {
	s, team, d, _ := setup(t)
	m := reputationMessage(t, s, team, d, []string{"reader@example.net"})
	valid := reputationFeedback(t, m, "Complaint", "", "reader@example.net")
	for _, raw := range [][]byte{
		reputationFeedback(t, m, "Complaint", ""),
		reputationFeedback(t, m, "Complaint", "", "reader@example.net", "foreign@example.net"),
		[]byte(strings.Replace(string(valid), tenant(team), tenant(uuid.NewString()), 1)),
		[]byte(strings.Replace(string(valid), "ses-message", "foreign-message", 1)),
	} {
		require.Error(t, s.ApplyFeedback(context.Background(), uuid.NewString(), raw))
	}
	a, err := s.Account(context.Background(), team)
	require.NoError(t, err)
	require.False(t, a.Suspended)
	var count int64
	require.NoError(t, s.DB.Model(&Suppression{}).Count(&count).Error)
	require.Zero(t, count, "invalid events must roll back suppression writes")
	require.NoError(t, s.DB.Model(&Audit{}).Count(&count).Error)
	require.Zero(t, count)
}

func TestConcurrentBouncesCannotMissSuspensionThreshold(t *testing.T) {
	s, team, d, _ := setup(t)
	recipients := []string{"one@example.net", "two@example.net", "three@example.net", "four@example.net", "five@example.net"}
	m := reputationMessage(t, s, team, d, recipients)
	var wg sync.WaitGroup
	errs := make(chan error, len(recipients))
	for _, recipient := range recipients {
		raw := reputationFeedback(t, m, "Bounce", "Permanent", recipient)
		wg.Add(1)
		go func(id string, raw []byte) { defer wg.Done(); errs <- s.ApplyFeedback(context.Background(), id, raw) }(recipient, raw)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	a, err := s.Account(context.Background(), team)
	require.NoError(t, err)
	require.True(t, a.Suspended)
	var count int64
	require.NoError(t, s.DB.Model(&Audit{}).Where("team_id = ? AND action = ?", team, "auto_suspend_bounces").Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func TestSuspensionRollsBackWithFeedbackWhenAuditFails(t *testing.T) {
	s, team, d, _ := setup(t)
	m := reputationMessage(t, s, team, d, []string{"reader@example.net"})
	require.NoError(t, s.DB.Callback().Create().Before("gorm:create").Register("fail_suspension_audit", func(tx *gorm.DB) {
		if tx.Statement.Table == "managed_audits" {
			tx.AddError(errors.New("audit unavailable"))
		}
	}))
	t.Cleanup(func() { s.DB.Callback().Create().Remove("fail_suspension_audit") })
	require.Error(t, s.ApplyFeedback(context.Background(), "complaint", reputationFeedback(t, m, "Complaint", "", "reader@example.net")))
	a, err := s.Account(context.Background(), team)
	require.NoError(t, err)
	require.False(t, a.Suspended)
	var count int64
	require.NoError(t, s.DB.Model(&Suppression{}).Count(&count).Error)
	require.Zero(t, count)
	require.NoError(t, s.DB.Model(&Event{}).Count(&count).Error)
	require.Zero(t, count)
	require.NoError(t, s.DB.First(&m, "id = ?", m.ID).Error)
	require.Equal(t, "SENT", m.Status)
}
