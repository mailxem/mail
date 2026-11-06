// Package hygiene keeps addresses that would bounce, complain or never be read
// out of audiences before Xem ever emails them.
package hygiene

import (
	"context"
	"errors"
	"net"
	"net/mail"
	"strings"
	"time"
)

// Reasons an address is left out. Imports report counts keyed by these.
const (
	Invalid    = "invalid email"
	NoMail     = "domain can't receive mail"
	Disposable = "disposable address"
	Role       = "role address"
	NeverEmail = "on the shared never-email list"
)

// Typo is the reason for a misspelled big-provider domain.
func Typo(domain string) string { return "did you mean " + domain }

// Resolver is the part of net.Resolver the domain check needs.
type Resolver interface {
	LookupMX(ctx context.Context, name string) ([]*net.MX, error)
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

// Checker caches domain lookups, so use one per import or signup.
type Checker struct {
	Resolver Resolver // nil uses net.DefaultResolver
	Import   bool     // role addresses are only skipped on imports, not when a person types one
	receives map[string]bool
}

// Reason says why email should not be added, or "" to keep it. It does not
// consult the shared never-email list; see Blocked.
func (c *Checker) Reason(ctx context.Context, email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	a, err := mail.ParseAddress(email)
	at := strings.LastIndexByte(email, '@')
	if err != nil || a.Address != email || at < 1 || !strings.Contains(email[at:], ".") {
		return Invalid
	}
	local, domain := email[:at], email[at+1:]
	if fix := typos[domain]; fix != "" {
		return Typo(fix)
	}
	if disposable[domain] {
		return Disposable
	}
	if c.Import && roles[strings.SplitN(local, "+", 2)[0]] {
		return Role
	}
	if !c.canReceive(ctx, domain) {
		return NoMail
	}
	return ""
}

// canReceive is false only when DNS definitely says the domain has no mail
// host: a null MX, or neither MX nor A/AAAA records. Timeouts and other DNS
// errors keep the address.
func (c *Checker) canReceive(ctx context.Context, domain string) bool {
	if reserved(domain) {
		return true
	}
	if ok, cached := c.receives[domain]; cached {
		return ok
	}
	r := c.Resolver
	if r == nil {
		r = net.DefaultResolver
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	ok := true
	mx, err := r.LookupMX(ctx, domain)
	if err == nil && len(mx) == 1 && mx[0].Host == "." {
		ok = false // RFC 7505 null MX: the domain accepts no mail
	} else if (err == nil && len(mx) == 0) || notFound(err) {
		_, err = r.LookupIPAddr(ctx, domain)
		ok = !notFound(err)
	}
	if c.receives == nil {
		c.receives = map[string]bool{}
	}
	c.receives[domain] = ok
	return ok
}

func notFound(err error) bool {
	var d *net.DNSError
	return errors.As(err, &d) && d.IsNotFound
}

// reserved names (RFC 2606) are used in docs, tests and local setups; never look them up.
func reserved(domain string) bool {
	for _, d := range []string{"example.com", "example.net", "example.org"} {
		if domain == d || strings.HasSuffix(domain, "."+d) {
			return true
		}
	}
	for _, tld := range []string{".test", ".example", ".localhost"} {
		if strings.HasSuffix(domain, tld) {
			return true
		}
	}
	return false
}

var typos = map[string]string{
	"gmial.com": "gmail.com", "gmal.com": "gmail.com", "gmai.com": "gmail.com", "gamil.com": "gmail.com",
	"gnail.com": "gmail.com", "gmaill.com": "gmail.com", "gmali.com": "gmail.com", "gmail.co": "gmail.com",
	"gmail.con": "gmail.com", "gmail.cm": "gmail.com", "gmail.om": "gmail.com", "gmail.cmo": "gmail.com",
	"googlemail.co": "googlemail.com",
	"hotmial.com":   "hotmail.com", "hotmal.com": "hotmail.com", "hotmai.com": "hotmail.com", "hotmil.com": "hotmail.com",
	"hotmail.co": "hotmail.com", "hotmail.con": "hotmail.com", "hotmail.cm": "hotmail.com",
	"yaho.com": "yahoo.com", "yahooo.com": "yahoo.com", "yhaoo.com": "yahoo.com", "yahoo.co": "yahoo.com",
	"yahoo.con": "yahoo.com", "yahoo.cm": "yahoo.com",
	"outlok.com": "outlook.com", "outllok.com": "outlook.com", "outlook.co": "outlook.com", "outlook.con": "outlook.com",
	"iclod.com": "icloud.com", "icoud.com": "icloud.com", "icloud.co": "icloud.com", "icloud.con": "icloud.com",
	"aol.co": "aol.com", "aol.con": "aol.com",
	"protonmial.com": "protonmail.com", "protonmail.co": "protonmail.com",
}

// ponytail: small fixed list of the busiest throwaway services; swap for a maintained list if they slip through.
var disposable = setOf("mailinator.com", "guerrillamail.com", "guerrillamail.net", "guerrillamail.org", "sharklasers.com",
	"grr.la", "10minutemail.com", "10minutemail.net", "temp-mail.org", "temp-mail.io", "tempmail.com", "tempmailo.com",
	"tempr.email", "yopmail.com", "yopmail.net", "trashmail.com", "getnada.com", "dispostable.com", "maildrop.cc",
	"throwawaymail.com", "fakeinbox.com", "mailnesia.com", "mintemail.com", "moakt.com", "emailondeck.com",
	"discard.email", "mohmal.com", "mytemp.email", "burnermail.io", "mailcatch.com", "inboxkitten.com",
	"spamgourmet.com", "getairmail.com", "mailpoof.com", "tempinbox.com")

var roles = setOf("info", "admin", "administrator", "support", "noreply", "no-reply", "donotreply", "do-not-reply",
	"postmaster", "hostmaster", "webmaster", "abuse", "sales", "marketing", "billing", "help", "contact",
	"office", "root", "security", "mailer-daemon")

func setOf(items ...string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, s := range items {
		m[s] = true
	}
	return m
}
