package sending

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSharedAddressIsReadyWithoutDNSAndStaysRemoved(t *testing.T) {
	s, team, _, p := setup(t)
	ctx := context.Background()
	s.Config.SharedDomain = "send.xem.test"
	s.DNS = fakeDNS{"_dmarc.send.xem.test": {"v=DMARC1; p=none"}}

	_, e := s.AddSharedAddress(ctx, team, "bounce")
	require.ErrorIs(t, e, ErrInvalidLabel)
	_, e = s.AddSharedAddress(ctx, team, "a")
	require.ErrorIs(t, e, ErrInvalidLabel)

	d, e := s.AddSharedAddress(ctx, team, "Sunday-Edit")
	require.NoError(t, e)
	require.Equal(t, "sunday-edit.send.xem.test", d.Name)
	require.True(t, d.Shared)
	require.True(t, d.Ready, "the operator's verified identity makes it ready")
	require.Empty(t, s.view(d).Records)

	_, e = s.AddSharedAddress(ctx, team, "another")
	require.ErrorIs(t, e, ErrHasShared, "one shared address per team")

	other := uuid.NewString()
	_, e = s.Account(ctx, other)
	require.NoError(t, e)
	require.NoError(t, s.DB.Model(&Domain{}).Where("id = ?", d.ID).Update("token", "removed").Error)
	_, e = s.AddSharedAddress(ctx, other, "sunday-edit")
	require.ErrorIs(t, e, ErrTaken, "another team cannot take a released label")

	// A removed address stays removed when the periodic check runs.
	d, e = s.RefreshDomain(ctx, team, d.ID)
	require.NoError(t, e)
	require.False(t, d.Ready)

	// The owner can take it back.
	d, e = s.AddSharedAddress(ctx, team, "sunday-edit")
	require.NoError(t, e)
	require.True(t, d.Ready)
	require.Positive(t, p.provisions)
}
