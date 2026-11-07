package sending

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// DNS hosts we offer a button for. Connect is a Domain Connect link that adds
// every record in one approval; it is empty until the registrar has onboarded
// Xem's template and DOMAIN_CONNECT_PRIVATE_KEY is set. Manage opens the
// registrar's DNS page for the zone, which always works.
type DNSHost struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Zone     string `json:"zone"`
	Detected bool   `json:"detected"`
	Connect  string `json:"connect,omitempty"`
	Manage   string `json:"manage"`
}

const dcProvider = "xem.email"

var dcAPIHosts = map[string]bool{"api.godaddy.com": true, "api.cloudflare.com/client/v4/dns/domainconnect": true}

var dnsHosts = []struct{ id, name, ns, manage string }{
	{"cloudflare", "Cloudflare", ".ns.cloudflare.com", "https://dash.cloudflare.com/?to=/:account/%s/dns/records"},
	{"godaddy", "GoDaddy", ".domaincontrol.com", "https://dcc.godaddy.com/manage/%s/dns"},
	{"namecheap", "Namecheap", ".registrar-servers.com", "https://ap.www.namecheap.com/Domains/DomainControlPanel/%s/advancedns"},
}

type nsResolver interface {
	LookupNS(context.Context, string) ([]*net.NS, error)
}

// zoneOf walks up from name to the first label with NS records: the zone the
// records are added to, and which registrar serves it.
func (s *Service) zoneOf(ctx context.Context, name string) (zone, host string) {
	r, ok := s.DNS.(nsResolver)
	if !ok {
		return registrable(name), ""
	}
	labels := strings.Split(name, ".")
	for i := 0; i < len(labels)-1; i++ {
		z := strings.Join(labels[i:], ".")
		ns, e := r.LookupNS(ctx, z)
		if e != nil || len(ns) == 0 {
			continue
		}
		for _, h := range dnsHosts {
			if strings.HasSuffix(strings.TrimSuffix(strings.ToLower(ns[0].Host), "."), h.ns) {
				return z, h.id
			}
		}
		return z, ""
	}
	return registrable(name), ""
}

// ponytail: last two labels; wrong for co.uk-style suffixes, which only matters
// when the NS lookup above fails.
func registrable(name string) string {
	labels := strings.Split(name, ".")
	if len(labels) <= 2 {
		return name
	}
	return strings.Join(labels[len(labels)-2:], ".")
}

func (s *Service) dnsHosts(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 8*time.Second)
	defer cancel()
	var d Domain
	if e := s.DB.WithContext(ctx).Where("id = ? AND team_id = ?", c.Param("id"), workspace(c)).First(&d).Error; e != nil {
		return echo.NewHTTPError(404, "Domain not found")
	}
	zone, detected := s.zoneOf(ctx, d.Name)
	connect := s.domainConnect(ctx, d, zone)
	out := make([]DNSHost, 0, len(dnsHosts))
	for _, h := range dnsHosts {
		v := DNSHost{ID: h.id, Name: h.name, Zone: zone, Detected: h.id == detected, Manage: strings.Replace(h.manage, "%s", url.PathEscape(zone), 1)}
		if v.Detected {
			v.Connect = connect
		}
		out = append(out, v)
	}
	return c.JSON(200, out)
}

// domainConnect returns a signed synchronous-flow apply link, or "" when the
// zone's DNS host doesn't offer Xem's template (yet).
func (s *Service) domainConnect(ctx context.Context, d Domain, zone string) string {
	key, e := domainConnectKey()
	if key == nil || e != nil {
		return ""
	}
	// The customer controls this TXT record, so only follow it to the
	// registrars we know; anything else would let a domain point our server
	// at an arbitrary host.
	txt, e := s.DNS.LookupTXT(ctx, "_domainconnect."+zone)
	if e != nil || len(txt) == 0 || !dcAPIHosts[txt[0]] {
		return ""
	}
	var settings struct{ URLSyncUX, URLAPI string }
	if getJSON(ctx, "https://"+txt[0]+"/v2/"+url.PathEscape(zone)+"/settings", &settings) != nil || !strings.HasPrefix(settings.URLSyncUX, "https://") || !strings.HasPrefix(settings.URLAPI, "https://") {
		return ""
	}
	service := "ownership"
	q := url.Values{"domain": {zone}, "token": {d.Token}}
	tokens := strings.Split(d.DKIMTokens, ",")
	if d.Provisioned && len(tokens) == 3 {
		service = "sending"
		q.Set("dkim1", tokens[0])
		q.Set("dkim2", tokens[1])
		q.Set("dkim3", tokens[2])
		q.Set("region", s.Config.Region)
	}
	path := "/v2/domainTemplates/providers/" + dcProvider + "/services/" + service
	if getJSON(ctx, settings.URLAPI+path, nil) != nil {
		return ""
	}
	if sub := strings.TrimSuffix(d.Name, "."+zone); sub != d.Name {
		q.Set("host", sub)
	}
	q.Set("redirect_uri", "https://app.xem.email/onboarding?dns=applied")
	query := q.Encode()
	sum := sha256.Sum256([]byte(query))
	sig, e := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum[:])
	if e != nil {
		return ""
	}
	return settings.URLSyncUX + path + "/apply?" + query + "&key=" + url.QueryEscape(keyHost()) + "&sig=" + url.QueryEscape(base64.StdEncoding.EncodeToString(sig))
}

// keyHost is the TXT label under xem.email holding the public key (syncPubKeyDomain).
func keyHost() string {
	if h := os.Getenv("DOMAIN_CONNECT_KEY_HOST"); h != "" {
		return h
	}
	return "_dck1"
}

func domainConnectKey() (*rsa.PrivateKey, error) {
	raw := os.Getenv("DOMAIN_CONNECT_PRIVATE_KEY")
	if raw == "" {
		return nil, nil
	}
	block, _ := pem.Decode([]byte(strings.ReplaceAll(raw, `\n`, "\n")))
	if block == nil {
		return nil, errors.New("DOMAIN_CONNECT_PRIVATE_KEY is not PEM")
	}
	if k, e := x509.ParsePKCS1PrivateKey(block.Bytes); e == nil {
		return k, nil
	}
	k, e := x509.ParsePKCS8PrivateKey(block.Bytes)
	if r, ok := k.(*rsa.PrivateKey); ok {
		return r, nil
	}
	return nil, errors.Join(e, errors.New("DOMAIN_CONNECT_PRIVATE_KEY is not RSA"))
}

var dcClient = &http.Client{Timeout: 4 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

func getJSON(ctx context.Context, u string, v any) error {
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return e
	}
	res, e := dcClient.Do(req)
	if e != nil {
		return e
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return errors.New(res.Status)
	}
	if v == nil {
		return nil
	}
	return json.NewDecoder(http.MaxBytesReader(nil, res.Body, 64<<10)).Decode(v)
}
