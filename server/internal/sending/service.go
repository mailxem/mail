package sending

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Starter ceilings apply only to automatic approval. Operators can review and
// raise limits separately; a DNS recheck never increases an existing allowance.
const (
	autoApprovalDailyLimit   int64 = 100
	autoApprovalMonthlyLimit int64 = 3000
	autoApprovalBudgetMicros int64 = 3000000
)

var ErrDenied = errors.New("sending is unavailable: check approval, pause status, domain readiness, and credential")
var ErrLimit = errors.New("sending limit reached; review usage or contact your administrator")
var ErrSuppressed = errors.New("recipient is suppressed or unsubscribed")
var domainPattern = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`)

func validDomain(s string) bool { return len(s) <= 253 && domainPattern.MatchString(s) }
func address(s string) (string, error) {
	a, e := mail.ParseAddress(s)
	if e != nil || strings.ContainsAny(a.Address, "\r\n") || strings.ContainsAny(a.Address, "\x00") {
		return "", errors.New("invalid email address")
	}
	for _, c := range a.Address {
		if c > 127 {
			return "", errors.New("international email addresses are not supported")
		}
	}
	return strings.ToLower(a.Address), nil
}
func randomSecret() string {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		panic(e)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

type Resolver interface {
	LookupTXT(context.Context, string) ([]string, error)
}
type Service struct {
	NotifyAlerts    bool
	Notify          MilestoneSender
	NotificationURL string
	DB              *gorm.DB
	Config          Config
	Provider        Provider
	DNS             Resolver
	Now             func() time.Time
}

func New(db *gorm.DB, c Config, p Provider) *Service {
	return &Service{DB: db, Config: c, Provider: p, DNS: net.DefaultResolver, Now: func() time.Time { return time.Now().UTC() }}
}

// applyPlan gives an account still on the automatic allowance its plan's
// limits; limits an operator set stay as they are.
func (s *Service) applyPlan(a *Account, daily, monthly int64, known bool) bool {
	if !known || a.DailyLimit != autoApprovalDailyLimit || a.MonthlyLimit != autoApprovalMonthlyLimit || a.MonthlyBudgetMicros != autoApprovalBudgetMicros {
		return false
	}
	a.DailyLimit, a.MonthlyLimit, a.MonthlyBudgetMicros = daily, monthly, monthly*s.Config.CostMicros
	return true
}

func (s *Service) Account(ctx context.Context, team string) (Account, error) {
	a := Account{TeamID: team, DailyLimit: s.Config.DailyLimit, MonthlyLimit: s.Config.MonthlyLimit, MonthlyBudgetMicros: s.Config.BudgetMicros}

	if uuid.Validate(team) != nil {
		return a, ErrDenied
	}
	if e := s.DB.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&a).Error; e != nil {
		return a, e
	}
	e := s.DB.WithContext(ctx).First(&a, "team_id = ?", team).Error
	// Usage display resets without writing; reservations perform the reset under lock.
	if a.Day != s.Now().Format("2006-01-02") {
		a.DailyUsed = 0
	}
	if a.Month != s.Now().Format("2006-01") {
		a.MonthlyUsed = 0
		a.BudgetUsedMicros = 0
	}
	return a, e
}
func (s *Service) AddDomain(ctx context.Context, team, name string) (Domain, error) {
	name = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(name), "."))
	d := Domain{ID: uuid.NewString(), TeamID: team, Name: name, Token: "xem-managed=" + randomSecret(), IdentityStatus: "NOT_STARTED", DKIMStatus: "NOT_STARTED", MAILFROMStatus: "NOT_STARTED", DMARCStatus: "NOT_CHECKED"}
	if !s.Config.Enabled || !validDomain(name) || strings.HasPrefix(name, "bounce.") {
		return d, errors.New("enter a valid sending domain, not its bounce subdomain")
	}
	if _, e := s.Account(ctx, team); e != nil {
		return d, e
	}
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var a Account
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&a, "team_id = ?", team).Error; e != nil {
			return e
		}
		var count int64
		if e := tx.Model(&Domain{}).Where("team_id = ?", team).Count(&count).Error; e != nil {
			return e
		}
		if count >= 10 {
			return ErrLimit
		}
		if err := tx.Create(&d).Error; err != nil {
			return err
		}
		return QueueMilestone(tx, team, d.ID, "domain_added")
	})
	return d, err
}
func (s *Service) RefreshDomain(ctx context.Context, team, id string) (result Domain, err error) {
	stage := "load_domain"
	defer func() {
		if err != nil {
			err = &domainCheckError{stage: stage, cause: err}
			logDomainCheckError(team, id, s.Config.Region, err)
		}
	}()
	var d Domain
	if !s.Config.Enabled {
		return d, ErrDenied
	}
	if e := s.DB.WithContext(ctx).First(&d, "id = ? AND team_id = ?", id, team).Error; e != nil {
		return d, e
	}
	// A newer check invalidates any in-flight result before it can approve an account.
	stage = "invalidate_check"
	d.CheckID = uuid.NewString()
	r := s.DB.WithContext(ctx).Model(&Domain{}).Where("id = ? AND team_id = ? AND token = ?", d.ID, d.TeamID, d.Token).Updates(map[string]any{"ready": false, "ownership": false, "check_id": d.CheckID})
	if r.Error != nil {
		return d, r.Error
	}
	if r.RowsAffected != 1 {
		return d, ErrDenied
	}
	d.Ready = false
	d.Ownership = false
	stage = "ownership_dns"
	records, e := s.DNS.LookupTXT(ctx, "_xem."+d.Name)
	if e != nil && !dnsRecordMissing(e) {
		return d, e
	}
	if e == nil {
		for _, r := range records {
			if r == d.Token {
				d.Ownership = true
			}
		}
	}
	now := s.Now()
	d.CheckedAt = &now
	if !d.Ownership {
		d.IdentityStatus = "OWNERSHIP_REQUIRED"
		stage = "save_domain"
		return d, s.persistDomain(ctx, &d)
	}
	stage = "load_account"
	a, err := s.Account(ctx, team)
	if err != nil {
		return d, err
	}
	if a.Suspended {
		stage = "save_domain"
		if err := s.persistDomain(ctx, &d); err != nil {
			return d, err
		}
		return d, ErrDenied
	}
	if !d.Provisioned {
		stage = "provision"
		if e = s.Provider.Provision(ctx, team, d.Name); e != nil {
			return d, fmt.Errorf("provider setup failed; contact the operator: %w", e)
		}
		d.Provisioned = true
	}
	stage = "identity"
	identity, e := s.Provider.Identity(ctx, d.Name)
	if e != nil {
		return d, e
	}
	d.IdentityStatus = identity.Status
	d.DKIMStatus = identity.DKIM
	d.MAILFROMStatus = identity.MAILFROM
	d.DKIMTokens = strings.Join(identity.Tokens, ",")
	d.DMARCStatus = "MISSING_OR_INVALID"
	stage = "dmarc_dns"
	records, e = s.DNS.LookupTXT(ctx, "_dmarc."+d.Name)
	if e != nil && !dnsRecordMissing(e) {
		return d, e
	}
	if e == nil && validDMARC(records) {
		d.DMARCStatus = "VALID"
	}
	d.Ready = identity.Verified && d.DKIMStatus == "SUCCESS" && d.MAILFROMStatus == "SUCCESS" && d.DMARCStatus == "VALID"
	stage = "save_domain"
	e = s.persistDomain(ctx, &d)
	return d, e
}
func validDMARC(records []string) bool {
	count := 0
	valid := false
	for _, r := range records {
		if !strings.HasPrefix(strings.TrimSpace(r), "v=DMARC1;") {
			continue
		}
		count++
		tags := map[string]string{}
		ok := true
		for _, part := range strings.Split(r, ";") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			kv := strings.SplitN(part, "=", 2)
			if len(kv) != 2 {
				ok = false
				break
			}
			if _, found := tags[kv[0]]; found {
				ok = false
			}
			tags[kv[0]] = strings.TrimSpace(kv[1])
		}
		valid = ok && (tags["p"] == "none" || tags["p"] == "quarantine" || tags["p"] == "reject")
	}
	return count == 1 && valid
}
func (s *Service) CreateCredential(ctx context.Context, team, domain, name string) (Credential, string, error) {
	secret := randomSecret()
	hash, e := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if e != nil {
		return Credential{}, "", e
	}
	cred := Credential{ID: uuid.NewString(), TeamID: team, DomainID: domain, Name: strings.TrimSpace(name), Hash: string(hash), ExpiresAt: s.Now().Add(90 * 24 * time.Hour)}
	if !s.Config.Enabled || !s.Config.SMTPEnabled || len(cred.Name) < 1 || len(cred.Name) > 80 {
		return cred, "", ErrDenied
	}
	e = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var a Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&a, "team_id = ?", team).Error; err != nil {
			return err
		}
		if !a.Approved || a.Suspended || a.Paused {
			return ErrDenied
		}
		var d Domain
		if err := tx.First(&d, "id = ? AND team_id = ? AND ready = true", domain, team).Error; err != nil {
			return ErrDenied
		}
		var n int64
		if err := tx.Model(&Credential{}).Where("team_id = ? AND revoked_at IS NULL AND expires_at > ?", team, s.Now()).Count(&n).Error; err != nil {
			return err
		}
		if n >= 20 {
			return ErrLimit
		}
		return tx.Create(&cred).Error
	})
	if e != nil {
		return cred, "", e
	}
	return cred, secret, nil
}
func (s *Service) Authenticate(ctx context.Context, id, password string) (Credential, error) {
	var c Credential
	if !s.Config.Enabled || len(password) > 128 || uuid.Validate(id) != nil {
		return c, ErrDenied
	}
	if e := s.DB.WithContext(ctx).First(&c, "id = ? AND revoked_at IS NULL AND expires_at > ?", id, s.Now()).Error; e != nil {
		return c, ErrDenied
	}
	if bcrypt.CompareHashAndPassword([]byte(c.Hash), []byte(password)) != nil {
		return Credential{}, ErrDenied
	}
	var a Account
	if e := s.DB.WithContext(ctx).First(&a, "team_id = ?", c.TeamID).Error; e != nil || !a.Approved || a.Suspended || a.Paused {
		return Credential{}, ErrDenied
	}
	return c, nil
}

// NormalizeRaw preserves MIME content, removes provider-routing and Bcc headers,
// rejects duplicate identity headers, and enforces envelope/header agreement.
func NormalizeRaw(raw []byte, envelope string) ([]byte, string, string, error) {
	m, e := mail.ReadMessage(bytes.NewReader(raw))
	if e != nil {
		return nil, "", "", errors.New("invalid MIME message")
	}
	if len(m.Header["From"]) != 1 {
		return nil, "", "", errors.New("exactly one From header is required")
	}
	from, e := address(m.Header.Get("From"))
	if e != nil {
		return nil, "", "", e
	}
	env, e := address(envelope)
	if e != nil || env != from {
		return nil, "", "", errors.New("envelope sender must match From")
	}
	if strings.ContainsAny(m.Header.Get("Subject"), "\r\n") {
		return nil, "", "", errors.New("invalid subject")
	}
	var out bytes.Buffer
	keys := make([]string, 0, len(m.Header))
	for k := range m.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		l := strings.ToLower(k)
		if strings.HasPrefix(l, "x-ses-") || l == "bcc" || l == "return-path" || l == "sender" || l == "dkim-signature" || l == "authentication-results" || strings.HasPrefix(l, "resent-") {
			continue
		}
		for _, v := range m.Header[k] {
			if strings.ContainsAny(k+v, "\r\n\x00") {
				return nil, "", "", errors.New("invalid MIME header")
			}
			fmt.Fprintf(&out, "%s: %s\r\n", k, v)
		}
	}
	out.WriteString("\r\n")
	if _, e = io.Copy(&out, m.Body); e != nil {
		return nil, "", "", e
	}
	return out.Bytes(), from, m.Header.Get("Subject"), nil
}
func validateMarketing(raw []byte) error {
	m, e := mail.ReadMessage(bytes.NewReader(raw))
	if e != nil {
		return e
	}
	u, e := url.Parse(strings.Trim(m.Header.Get("List-Unsubscribe"), "<>"))
	if e != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || m.Header.Get("List-Unsubscribe-Post") != "List-Unsubscribe=One-Click" {
		return errors.New("marketing mail requires HTTPS one-click unsubscribe headers")
	}
	return nil
}

// Suppressed checks recipient addresses, including submissions without contact IDs.
func Suppressed(tx *gorm.DB, team, recipient string, marketing bool) (bool, error) {
	blocked, e := suppressedSet(tx, team, []string{recipient}, marketing)
	return e != nil || blocked[recipient], e
}

type Submission struct {
	TeamID, DomainID, CredentialID, EmailID, Key, From string
	Recipients                                         []string
	Raw                                                []byte
	IsTest                                             bool
	Marketing                                          bool
	// Transient: Raw is checked but not stored; the worker rebuilds it from
	// EmailID at dispatch (RenderRaw).
	Transient bool
}

func (s *Service) Submit(ctx context.Context, in Submission) (Message, error) {
	r := s.SubmitBatch(ctx, []Submission{in})
	return r[0].Message, r[0].Err
}

func (s *Service) persistDomain(ctx context.Context, d *Domain) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Use the same account lock as operator actions, quota reservations and feedback.
		var a Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&a, "team_id = ?", d.TeamID).Error; err != nil {
			return err
		}
		if a.Suspended {
			d.Ready = false
		}
		r := tx.Model(&Domain{}).Where("id = ? AND team_id = ? AND token = ? AND check_id = ?", d.ID, d.TeamID, d.Token, d.CheckID).Updates(map[string]any{"ready": d.Ready, "ownership": d.Ownership, "provisioned": d.Provisioned, "identity_status": d.IdentityStatus, "dkim_status": d.DKIMStatus, "mailfrom_status": d.MAILFROMStatus, "dmarc_status": d.DMARCStatus, "dkim_tokens": d.DKIMTokens, "checked_at": d.CheckedAt})
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected != 1 {
			return ErrDenied
		}
		if d.Ownership {
			if err := QueueMilestone(tx, d.TeamID, d.ID, "ownership_verified"); err != nil {
				return err
			}
		}
		if d.Ready {
			if err := QueueMilestone(tx, d.TeamID, d.ID, "domain_ready"); err != nil {
				return err
			}
		}
		if !d.Ready || !d.Ownership || !d.Provisioned || a.Approved || a.Suspended || a.Paused {
			return nil
		}
		// Auto-approval never raises existing limits or resets consumed quota.
		if err := tx.Model(&a).Updates(map[string]any{
			"approved":              true,
			"daily_limit":           min(a.DailyLimit, autoApprovalDailyLimit),
			"monthly_limit":         min(a.MonthlyLimit, autoApprovalMonthlyLimit),
			"monthly_budget_micros": min(a.MonthlyBudgetMicros, autoApprovalBudgetMicros),
		}).Error; err != nil {
			return err
		}
		if err := tx.Create(&Audit{ID: uuid.NewString(), TeamID: a.TeamID, Actor: "system:dns-verification", Action: "auto_approve", CreatedAt: s.Now()}).Error; err != nil {
			return err
		}
		return QueueMilestone(tx, a.TeamID, d.ID, "approved")
	})
}
