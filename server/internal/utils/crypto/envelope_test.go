package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestEnvelopeSecretsAreBoundAndSupportLongOAuthTokens(t *testing.T) {
	oldPrivate, oldPublic := PrivateKey, PublicKey
	t.Cleanup(func() { PrivateKey, PublicKey = oldPrivate, oldPublic })
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	PrivateKey, PublicKey = key, &key.PublicKey
	secret := []byte(strings.Repeat("oauth-refresh-token", 1000))
	sealed, err := SealSecret(secret, "team:one:connection:one")
	require.NoError(t, err)
	value, err := OpenSecret(sealed, "team:one:connection:one")
	require.NoError(t, err)
	require.Equal(t, secret, value)
	_, err = OpenSecret(sealed, "team:two:connection:one")
	require.Error(t, err)
	_, err = OpenSecret(sealed, "team:one:connection:two")
	require.Error(t, err)
	_, err = OpenSecret(sealed[:len(sealed)-5]+"abcde", "team:one:connection:one")
	require.Error(t, err)
	other, err := SealSecret(secret, "team:one:connection:one")
	require.NoError(t, err)
	require.NotEqual(t, sealed, other)
}
