package sending

import (
	"context"
	"errors"
	"net"
	"testing"
)

type nsDNS map[string]string

func (r nsDNS) LookupTXT(context.Context, string) ([]string, error) { return nil, errors.New("none") }
func (r nsDNS) LookupNS(_ context.Context, name string) ([]*net.NS, error) {
	if h, ok := r[name]; ok {
		return []*net.NS{{Host: h}}, nil
	}
	return nil, errors.New("no such host")
}

func TestZoneOfFindsTheRegistrarForASubdomain(t *testing.T) {
	s := &Service{DNS: nsDNS{"boringproducts.xyz": "marek.ns.cloudflare.com."}}
	if zone, host := s.zoneOf(context.Background(), "humans.boringproducts.xyz"); zone != "boringproducts.xyz" || host != "cloudflare" {
		t.Fatalf("got %q %q", zone, host)
	}
	s.DNS = nsDNS{"shop.example.com": "ns07.domaincontrol.com"}
	if zone, host := s.zoneOf(context.Background(), "mail.shop.example.com"); zone != "shop.example.com" || host != "godaddy" {
		t.Fatalf("got %q %q", zone, host)
	}
	s.DNS = nsDNS{"example.org": "ns1.example.net"}
	if zone, host := s.zoneOf(context.Background(), "example.org"); zone != "example.org" || host != "" {
		t.Fatalf("got %q %q", zone, host)
	}
}

func TestDomainConnectNeedsAKnownRegistrar(t *testing.T) {
	t.Setenv("DOMAIN_CONNECT_PRIVATE_KEY", "")
	s := &Service{DNS: nsDNS{}}
	if u := s.domainConnect(context.Background(), Domain{Name: "a.example.com"}, "example.com"); u != "" {
		t.Fatalf("expected no link without a key, got %q", u)
	}
}
