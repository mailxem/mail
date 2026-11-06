package sending

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"kori/internal/hygiene"
)

func TestHardBounceKeepsAddressFromEveryWorkspace(t *testing.T) {
	t.Setenv("SUPPRESSION_HASH_KEY", "test-key")
	s, team, d, _ := setup(t)
	require.NoError(t, s.DB.AutoMigrate(&hygiene.NeverEmailEntry{}))
	ctx := context.Background()
	other := uuid.NewString()

	m := reputationMessage(t, s, team, d, []string{"gone@example.net", "soft@example.net"})
	require.NoError(t, s.ApplyFeedback(ctx, "bounce-event", reputationFeedback(t, m, "Bounce", "Permanent", "gone@example.net")))

	blocked, err := Suppressed(s.DB, other, "gone@example.net", false)
	require.NoError(t, err)
	require.True(t, blocked, "a hard bounce in one workspace blocks the address in another")
	blocked, err = Suppressed(s.DB, other, "soft@example.net", false)
	require.NoError(t, err)
	require.False(t, blocked)

	var hashes []string
	require.NoError(t, s.DB.Model(&hygiene.NeverEmailEntry{}).Pluck("hash", &hashes).Error)
	require.Len(t, hashes, 1)
	require.NotContains(t, hashes[0], "@")

	t.Setenv("SUPPRESSION_HASH_KEY", "")
	blocked, err = Suppressed(s.DB, other, "gone@example.net", false)
	require.NoError(t, err)
	require.False(t, blocked, "without a key the shared list is off")
}
