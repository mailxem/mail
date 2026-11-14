package sending

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

type cnameDNS struct {
	fakeDNS
	cnames map[string]string
}

func (r cnameDNS) LookupCNAME(_ context.Context, name string) (string, error) {
	if c, ok := r.cnames[name]; ok {
		return c, nil
	}
	return "", &net.DNSError{Name: name, IsNotFound: true}
}

func TestRefreshDomainFlagsCloudflareProxiedDKIM(t *testing.T) {
	s, team, d, _ := setup(t)
	txt := fakeDNS{"_xem.example.com": {d.Token}, "_dmarc.example.com": {"v=DMARC1; p=none"}}
	// A proxied record flattens the CNAME: the name resolves to itself.
	s.DNS = cnameDNS{txt, map[string]string{
		"one._domainkey.example.com": "one.dkim.amazonses.com.",
		"two._domainkey.example.com": "two._domainkey.example.com.",
	}}
	d, e := s.RefreshDomain(context.Background(), team, d.ID)
	require.NoError(t, e)
	require.Equal(t, "CLOUDFLARE_PROXIED", d.DKIMStatus)
	require.False(t, d.Ready)

	s.DNS = cnameDNS{txt, map[string]string{"two._domainkey.example.com": "two.dkim.amazonses.com."}}
	d, e = s.RefreshDomain(context.Background(), team, d.ID)
	require.NoError(t, e)
	require.Equal(t, "SUCCESS", d.DKIMStatus)
	require.True(t, d.Ready)
}
