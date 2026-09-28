package sending

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func unapprovedSending(t *testing.T) (*Service, string, Domain, *fakeProvider) {
	t.Helper()
	s, team, d, p := setup(t)
	require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Updates(map[string]any{"approved": false, "daily_limit": 100, "monthly_limit": 3000, "monthly_budget_micros": 3000000}).Error)
	require.NoError(t, s.DB.Model(&Domain{}).Where("id = ?", d.ID).Updates(map[string]any{"ready": false, "provisioned": false, "ownership": false, "dkim_tokens": ""}).Error)
	p.provisions = 0
	return s, team, d, p
}

func TestAutomaticApprovalRequiresCompleteDNS(t *testing.T) {
	s, team, d, p := unapprovedSending(t)
	ctx := context.Background()
	s.DNS = fakeDNS{}
	d, err := s.RefreshDomain(ctx, team, d.ID)
	require.NoError(t, err)
	require.False(t, d.Ownership)
	require.Zero(t, p.provisions, "unverified domains must not create AWS resources")
	s.DNS = fakeDNS{"_xem." + d.Name: {d.Token}, "_dmarc." + d.Name: {"v=DMARC1; p=none"}}
	p.identity.Verified, p.identity.DKIM, p.identity.MAILFROM = false, "PENDING", "PENDING"
	d, err = s.RefreshDomain(ctx, team, d.ID)
	require.NoError(t, err)
	require.True(t, d.Ownership)
	require.True(t, d.Provisioned)
	require.False(t, d.Ready)
	require.Equal(t, 1, p.provisions)
	require.Len(t, s.view(d).Records, 6, "ownership, three DKIM, bounce MX and SPF must be returned before approval")
	a, err := s.Account(ctx, team)
	require.NoError(t, err)
	require.False(t, a.Approved)
	_, err = s.Submit(ctx, input(team, d))
	require.ErrorIs(t, err, ErrDenied)

	p.identity.Verified, p.identity.DKIM = true, "SUCCESS"
	d, err = s.RefreshDomain(ctx, team, d.ID)
	require.NoError(t, err)
	require.False(t, d.Ready, "MAIL FROM must pass")
	p.identity.MAILFROM = "SUCCESS"
	s.DNS = fakeDNS{"_xem." + d.Name: {d.Token}}
	d, err = s.RefreshDomain(ctx, team, d.ID)
	require.NoError(t, err)
	require.False(t, d.Ready, "DMARC must pass")
	a, err = s.Account(ctx, team)
	require.NoError(t, err)
	require.False(t, a.Approved)
	s.DNS = fakeDNS{"_xem." + d.Name: {d.Token}, "_dmarc." + d.Name: {"v=DMARC1; p=none"}}
	d, err = s.RefreshDomain(ctx, team, d.ID)
	require.NoError(t, err)
	require.True(t, d.Ready)
	a, err = s.Account(ctx, team)
	require.NoError(t, err)
	require.True(t, a.Approved)
	require.EqualValues(t, 100, a.DailyLimit)
	require.EqualValues(t, 3000, a.MonthlyLimit)
	require.EqualValues(t, 3000000, a.MonthlyBudgetMicros)
	_, err = s.Submit(ctx, input(team, d))
	require.NoError(t, err)
	_, err = s.RefreshDomain(ctx, team, d.ID)
	require.NoError(t, err)
	var count int64
	require.NoError(t, s.DB.Model(&Audit{}).Where("team_id = ? AND action = ?", team, "auto_approve").Count(&count).Error)
	require.EqualValues(t, 1, count, "rechecking must not issue approval again")
	require.Equal(t, 1, p.provisions)
	// The assigned starter limit must be enforced by submission, not just shown.
	require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Update("daily_used", 100).Error)
	_, err = s.Submit(ctx, input(team, d))
	require.ErrorIs(t, err, ErrLimit)
}

func TestAutoApprovalRollsBackWhenAuditCannotBeSaved(t *testing.T) {
	s, team, d, _ := unapprovedSending(t)
	require.NoError(t, s.DB.Callback().Create().Before("gorm:create").Register("fail_approval_audit", func(tx *gorm.DB) {
		if tx.Statement.Table == "managed_audits" {
			tx.AddError(errors.New("audit unavailable"))
		}
	}))
	t.Cleanup(func() { s.DB.Callback().Create().Remove("fail_approval_audit") })
	_, err := s.RefreshDomain(context.Background(), team, d.ID)
	require.Error(t, err)
	var stored Domain
	require.NoError(t, s.DB.First(&stored, "id = ?", d.ID).Error)
	require.False(t, stored.Ready)
	a, err := s.Account(context.Background(), team)
	require.NoError(t, err)
	require.False(t, a.Approved)
}

func TestAutoApprovalPreservesLimitsUsageAndOperatorDecisions(t *testing.T) {
	for _, mode := range []string{"lower limits", "operator approved", "paused", "suspended"} {
		t.Run(mode, func(t *testing.T) {
			s, team, d, p := unapprovedSending(t)
			fields := map[string]any{"daily_limit": 7, "monthly_limit": 20, "monthly_budget_micros": 12000, "day": s.Now().Format("2006-01-02"), "month": s.Now().Format("2006-01"), "daily_used": 3, "monthly_used": 5, "budget_used_micros": 5000}
			if mode == "operator approved" {
				fields["approved"], fields["daily_limit"] = true, 500
			}
			if mode == "paused" {
				fields["paused"] = true
			}
			if mode == "suspended" {
				fields["suspended"] = true
			}
			require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Updates(fields).Error)
			_, err := s.RefreshDomain(context.Background(), team, d.ID)
			if mode == "suspended" {
				require.ErrorIs(t, err, ErrDenied)
				require.Zero(t, p.provisions)
			} else {
				require.NoError(t, err)
			}
			a, err := s.Account(context.Background(), team)
			require.NoError(t, err)
			require.Equal(t, mode == "lower limits" || mode == "operator approved", a.Approved)
			require.EqualValues(t, fields["daily_limit"], a.DailyLimit)
			require.EqualValues(t, 20, a.MonthlyLimit)
			require.EqualValues(t, 12000, a.MonthlyBudgetMicros)
			require.EqualValues(t, 3, a.DailyUsed)
			require.EqualValues(t, 5, a.MonthlyUsed)
			require.EqualValues(t, 5000, a.BudgetUsedMicros)
			require.Equal(t, mode == "paused", a.Paused)
			require.Equal(t, mode == "suspended", a.Suspended)
		})
	}
}

type onboardingProvider struct {
	*fakeProvider
	identity func() (Identity, error)
}

func TestConcurrentDNSChecksApproveOnlyOnce(t *testing.T) {
	s, team, d, _ := unapprovedSending(t)
	require.NoError(t, s.DB.Model(&Domain{}).Where("id = ?", d.ID).Update("provisioned", true).Error)
	var wg sync.WaitGroup
	errs := make(chan error, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := s.RefreshDomain(context.Background(), team, d.ID); errs <- err }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			require.ErrorIs(t, err, ErrDenied)
		}
	}
	a, err := s.Account(context.Background(), team)
	require.NoError(t, err)
	require.True(t, a.Approved)
	var count int64
	require.NoError(t, s.DB.Model(&Audit{}).Where("team_id = ? AND action = ?", team, "auto_approve").Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func (p onboardingProvider) Identity(context.Context, string) (Identity, error) { return p.identity() }

func TestAutoApprovalRejectsStaleAndFailedVerification(t *testing.T) {
	for _, mode := range []string{"provider failure", "newer check", "disconnect", "suspend during check"} {
		t.Run(mode, func(t *testing.T) {
			s, team, d, p := unapprovedSending(t)
			s.Provider = onboardingProvider{p, func() (Identity, error) {
				switch mode {
				case "provider failure":
					return Identity{}, errors.New("provider unavailable")
				case "newer check":
					require.NoError(t, s.DB.Model(&Domain{}).Where("id = ?", d.ID).Update("check_id", uuid.NewString()).Error)
				case "disconnect":
					require.NoError(t, s.DB.Model(&Domain{}).Where("id = ?", d.ID).Update("token", "rotated").Error)
				case "suspend during check":
					require.NoError(t, s.DB.Model(&Account{}).Where("team_id = ?", team).Update("suspended", true).Error)
				}
				return p.identity, nil
			}}
			result, err := s.RefreshDomain(context.Background(), team, d.ID)
			if mode == "suspend during check" {
				require.NoError(t, err)
				require.False(t, result.Ready)
			} else {
				require.Error(t, err)
			}
			var stored Domain
			require.NoError(t, s.DB.First(&stored, "id = ?", d.ID).Error)
			require.False(t, stored.Ready)
			a, err := s.Account(context.Background(), team)
			require.NoError(t, err)
			require.False(t, a.Approved)
		})
	}
}
