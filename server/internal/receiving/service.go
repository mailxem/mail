package receiving

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/models"
	"kori/internal/sending"
)

var ErrDisabled = errors.New("managed receiving is disabled")
var ErrMailboxExists = errors.New("a mailbox already exists for this address")
var localPattern = regexp.MustCompile(`^[a-z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+$`)

type MX struct {
	Host     string `json:"host"`
	Priority uint16 `json:"priority"`
}
type RuleProvider interface {
	ActiveRuleSet(context.Context) (string, error)
	PutRule(context.Context, string, []string) error
	DeleteRule(context.Context, string) error
	RuleReady(context.Context, string, []string) (bool, error)
}
type Resolver interface {
	LookupMX(context.Context, string) ([]*net.MX, error)
}
type Service struct {
	DB     *gorm.DB
	Config Config
	Rules  RuleProvider
	Store  ObjectStore
	DNS    Resolver
	Now    func() time.Time
}

var defaultService *Service
var defaultMu sync.RWMutex

func RegisterDefault(s *Service) { defaultMu.Lock(); defaultService = s; defaultMu.Unlock() }
func Current() *Service          { defaultMu.RLock(); defer defaultMu.RUnlock(); return defaultService }

func New(db *gorm.DB, c Config, r RuleProvider) *Service {
	return &Service{DB: db, Config: c, Rules: r, DNS: net.DefaultResolver, Now: func() time.Time { return time.Now().UTC() }}
}
func ruleName(domainID string) string { return "xem-inbox-" + strings.ReplaceAll(domainID, "-", "") }

func (s *Service) List(ctx context.Context, team string) ([]sending.Domain, []DomainState, []Mailbox, error) {
	domains := []sending.Domain{}
	states := []DomainState{}
	boxes := []Mailbox{}
	if err := s.DB.WithContext(ctx).Where("team_id = ?", team).Order("created_at,id").Find(&domains).Error; err != nil {
		return nil, nil, nil, err
	}
	if err := s.DB.WithContext(ctx).Where("team_id = ?", team).Find(&states).Error; err != nil {
		return nil, nil, nil, err
	}
	if err := s.DB.WithContext(ctx).Where("team_id = ?", team).Order("address,id").Find(&boxes).Error; err != nil {
		return nil, nil, nil, err
	}
	return domains, states, boxes, nil
}

func normalizeLocal(v string) (string, error) {
	v = strings.ToLower(strings.TrimSpace(v))
	if !localPattern.MatchString(v) || len(v) > 64 || strings.Contains(v, "..") || strings.HasPrefix(v, ".") || strings.HasSuffix(v, ".") {
		return "", errors.New("invalid mailbox local part")
	}
	return v, nil
}

func (s *Service) CreateMailbox(ctx context.Context, team, domainID, local, display string) (Mailbox, error) {
	if !s.Config.Enabled {
		return Mailbox{}, ErrDisabled
	}
	local, err := normalizeLocal(local)
	if err != nil {
		return Mailbox{}, err
	}
	display = strings.TrimSpace(display)
	if len(display) > 120 {
		return Mailbox{}, errors.New("display name is too long")
	}
	box := Mailbox{}
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var d sending.Domain
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND team_id = ?", domainID, team).First(&d).Error; e != nil {
			return e
		}
		if !d.Ownership {
			return errors.New("domain ownership must be verified")
		}
		var account sending.Account
		if e := tx.Where("team_id = ?", team).First(&account).Error; e != nil {
			return e
		}
		if account.Suspended {
			return errors.New("domain receiving is unavailable")
		}
		address := local + "@" + d.Name
		if parsed, e := mail.ParseAddress(address); e != nil || parsed.Address != address {
			return errors.New("invalid mailbox address")
		}
		var addressCount int64
		if e := tx.Model(&Mailbox{}).Where("lower(address) = ?", strings.ToLower(address)).Count(&addressCount).Error; e != nil {
			return e
		}
		if addressCount > 0 {
			return ErrMailboxExists
		}
		var count int64
		if e := tx.Model(&Mailbox{}).Where("domain_id = ?", d.ID).Count(&count).Error; e != nil {
			return e
		}
		if count >= int64(s.Config.MaxMailboxesPerDomain) {
			return errors.New("mailbox limit reached")
		}
		id := uuid.NewString()
		senderID := uuid.NewString()
		senderReady := account.Approved && !account.Paused && d.Ready && d.Provisioned && d.CheckedAt != nil && s.Now().Sub(*d.CheckedAt) <= 24*time.Hour
		sender := models.SMTPConfig{Base: models.Base{ID: senderID}, Provider: "MANAGED", Host: "managed.internal", Port: 587, Username: address, FromEmail: address, IsActive: senderReady, SupportsTLS: true, RequiresAuth: true, MaxSendRate: 1, TeamID: team}
		if e := tx.Session(&gorm.Session{SkipHooks: true}).Create(&sender).Error; e != nil {
			return e
		}
		box = Mailbox{ID: id, TeamID: team, DomainID: d.ID, Address: address, DisplayName: display, Active: true, SMTPConfigID: senderID, Status: "provisioning", UIDValidity: uidValidity(id), QuotaBytes: s.Config.MailboxQuotaBytes, CreatedAt: s.Now(), UpdatedAt: s.Now()}
		return tx.Create(&box).Error
	})
	if err != nil {
		message := err.Error()
		if strings.Contains(message, "idx_managed_receiving_mailboxes_address") || strings.Contains(message, "managed_receiving_mailboxes.address") {
			return Mailbox{}, ErrMailboxExists
		}
		return Mailbox{}, err
	}
	if err = s.reconcileDomain(ctx, team, domainID); err != nil {
		if updateErr := s.DB.Model(&Mailbox{}).Where("id = ? AND active = true", box.ID).Updates(map[string]any{"status": "error", "updated_at": s.Now()}).Error; updateErr != nil {
			return box, updateErr
		}
		_ = s.DB.First(&box, "id = ?", box.ID).Error
		return box, nil
	}
	if err := s.DB.First(&box, "id = ? AND team_id = ?", box.ID, team).Error; err != nil {
		return box, err
	}
	return box, nil
}

func uidValidity(id string) uint32 {
	u, _ := uuid.Parse(id)
	n := uint32(u[0])<<24 | uint32(u[1])<<16 | uint32(u[2])<<8 | uint32(u[3])
	if n == 0 {
		n = 1
	}
	return n
}

func (s *Service) SetActive(ctx context.Context, team, id string, active bool) (Mailbox, error) {
	if !s.Config.Enabled {
		return Mailbox{}, ErrDisabled
	}
	var box Mailbox
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var initial Mailbox
		if e := tx.Where("id = ? AND team_id = ?", id, team).First(&initial).Error; e != nil {
			return e
		}
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND team_id = ?", initial.DomainID, team).First(&sending.Domain{}).Error; e != nil {
			return e
		}
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND team_id = ?", id, team).First(&box).Error; e != nil {
			return e
		}
		status := "inactive"
		if active {
			status = "provisioning"
		}
		if e := tx.Model(&box).Updates(map[string]any{"active": active, "status": status, "updated_at": s.Now()}).Error; e != nil {
			return e
		}
		return tx.Session(&gorm.Session{SkipHooks: true}).Model(&models.SMTPConfig{}).Where("id = ? AND team_id = ? AND provider = ?", box.SMTPConfigID, team, "MANAGED").Update("is_active", active).Error
	})
	if err != nil {
		return box, err
	}
	if err = s.reconcileDomain(ctx, team, box.DomainID); err != nil {
		if updateErr := s.DB.Model(&Mailbox{}).Where("id = ? AND team_id = ? AND active = ?", box.ID, team, active).Updates(map[string]any{"status": "error", "updated_at": s.Now()}).Error; updateErr != nil {
			return box, updateErr
		}
		_ = s.DB.First(&box, "id = ? AND team_id = ?", box.ID, team).Error
		return box, nil
	}
	if err := s.DB.First(&box, "id = ? AND team_id = ?", box.ID, team).Error; err != nil {
		return box, err
	}
	return box, nil
}

func (s *Service) reconcileDomain(ctx context.Context, team, domainID string) error {
	if s.Rules == nil {
		return errors.New("receiving rule provider unavailable")
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var d sending.Domain
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND team_id = ?", domainID, team).First(&d).Error; e != nil {
			return e
		}
		var account sending.Account
		if e := tx.Where("team_id = ?", team).First(&account).Error; e != nil {
			return e
		}
		if !d.Ownership || account.Suspended {
			return s.Rules.DeleteRule(ctx, ruleName(domainID))
		}
		active, e := s.Rules.ActiveRuleSet(ctx)
		if e != nil {
			return e
		}
		if active != s.Config.RuleSet {
			return fmt.Errorf("configured receipt rule set is not active")
		}
		boxes := []Mailbox{}
		if e = tx.Where("team_id = ? AND domain_id = ? AND active = true AND status <> ?", team, domainID, "storage_full").Order("address").Find(&boxes).Error; e != nil {
			return e
		}
		recipients := make([]string, 0, len(boxes))
		senderReady := account.Approved && !account.Paused && d.Ready && d.Provisioned && d.CheckedAt != nil && s.Now().Sub(*d.CheckedAt) <= 24*time.Hour
		finalize := func() error {
			if err := tx.Model(&Mailbox{}).Where("team_id = ? AND domain_id = ? AND active = true AND status <> ?", team, domainID, "storage_full").Updates(map[string]any{"status": "active", "updated_at": s.Now()}).Error; err != nil {
				return err
			}
			ids := make([]string, 0, len(boxes))
			for _, b := range boxes {
				ids = append(ids, b.SMTPConfigID)
			}
			if len(ids) > 0 {
				return tx.Session(&gorm.Session{SkipHooks: true}).Model(&models.SMTPConfig{}).Where("id IN ? AND team_id = ? AND provider = ?", ids, team, "MANAGED").Update("is_active", senderReady).Error
			}
			return nil
		}
		for _, b := range boxes {
			recipients = append(recipients, b.Address)
		}
		if len(recipients) == 0 {
			return s.Rules.DeleteRule(ctx, ruleName(domainID))
		}
		if ready, e := s.Rules.RuleReady(ctx, ruleName(domainID), recipients); e != nil {
			return e
		} else if ready {
			return finalize()
		}
		if err := s.Rules.PutRule(ctx, ruleName(domainID), recipients); err != nil {
			return err
		}
		return finalize()
	})
}

func (s *Service) ReconcileAll(ctx context.Context) error {
	if !s.Config.Enabled {
		return nil
	}
	var domains []struct {
		TeamID   string
		DomainID string
	}
	if err := s.DB.WithContext(ctx).Model(&Mailbox{}).Distinct("team_id", "domain_id").Scan(&domains).Error; err != nil {
		return err
	}
	var joined error
	for _, d := range domains {
		attemptStarted := s.Now()
		domainCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		err := s.reconcileDomain(domainCtx, d.TeamID, d.DomainID)
		if err == nil {
			_, _, err = s.CheckDomain(domainCtx, d.TeamID, d.DomainID)
		}
		cancel()
		if err != nil {
			joined = errors.Join(joined, err)
			now := s.Now()
			var ids []string
			_ = s.DB.WithContext(ctx).Model(&Mailbox{}).Where("team_id = ? AND domain_id = ? AND updated_at <= ?", d.TeamID, d.DomainID, attemptStarted).Pluck("smtp_config_id", &ids).Error
			_ = s.DB.WithContext(ctx).Model(&Mailbox{}).Where("team_id = ? AND domain_id = ? AND active = true AND status <> ? AND updated_at <= ?", d.TeamID, d.DomainID, "storage_full", attemptStarted).Updates(map[string]any{"status": "error", "updated_at": now}).Error
			if len(ids) > 0 {
				_ = s.DB.WithContext(ctx).Session(&gorm.Session{SkipHooks: true}).Model(&models.SMTPConfig{}).Where("id IN ? AND updated_at <= ?", ids, attemptStarted).Update("is_active", false).Error
			}
			updated := s.DB.WithContext(ctx).Model(&DomainState{}).Where("domain_id = ? AND updated_at <= ?", d.DomainID, attemptStarted).Updates(map[string]any{"status": "error", "detail": "Receipt rule reconciliation failed", "updated_at": now})
			if updated.Error == nil && updated.RowsAffected == 0 {
				_ = s.DB.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&DomainState{DomainID: d.DomainID, TeamID: d.TeamID, Status: "error", Detail: "Receipt rule reconciliation failed", CreatedAt: now, UpdatedAt: now}).Error
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1100 * time.Millisecond):
		}
	}
	return joined
}

func (s *Service) CheckDomain(ctx context.Context, team, id string) (DomainState, []MX, error) {
	var existingMX []MX
	var d sending.Domain
	if err := s.DB.WithContext(ctx).Where("id = ? AND team_id = ?", id, team).First(&d).Error; err != nil {
		return DomainState{}, nil, err
	}
	state := DomainState{DomainID: id, TeamID: team, Status: "unconfigured", CreatedAt: s.Now(), UpdatedAt: s.Now()}
	if !s.Config.Enabled {
		return state, existingMX, nil
	}
	var account sending.Account
	if err := s.DB.WithContext(ctx).Where("team_id = ?", team).First(&account).Error; err != nil {
		return state, nil, err
	}
	if account.Suspended {
		state.Status = "error"
		state.Detail = "Workspace receiving is suspended"
		return state, existingMX, s.persistDomainState(ctx, &state, existingMX)
	}
	var existing DomainState
	if s.DB.Where("domain_id = ? AND team_id = ?", id, team).First(&existing).Error == nil {
		state.CreatedAt = existing.CreatedAt
	}
	if !d.Ownership {
		state.Status = "needs_verification"
	} else {
		boxes := []Mailbox{}
		if err := s.DB.Where("domain_id = ? AND team_id = ? AND active = true AND status <> ?", id, team, "storage_full").Order("address").Find(&boxes).Error; err != nil {
			return state, nil, err
		}
		if len(boxes) == 0 {
			var mailboxCount int64
			if err := s.DB.Model(&Mailbox{}).Where("domain_id = ? AND team_id = ?", id, team).Count(&mailboxCount).Error; err != nil {
				return state, nil, err
			}
			if mailboxCount == 0 {
				state.Status = "needs_mailbox"
			} else {
				state.Status = "provisioning"
				state.Detail = "No mailbox is currently accepting new mail"
			}
		} else {
			records, err := s.DNS.LookupMX(ctx, d.Name)
			if err != nil {
				state.Status = "pending_mx"
				state.Detail = "MX lookup failed"
			} else {
				mx := make([]MX, 0, len(records))
				expected := "inbound-smtp." + s.Config.Region + ".amazonaws.com"
				found := false
				conflict := false
				for _, r := range records {
					host := strings.TrimSuffix(strings.ToLower(r.Host), ".")
					mx = append(mx, MX{host, r.Pref})
					if host == expected {
						found = true
					} else {
						conflict = true
					}
				}
				sort.Slice(mx, func(i, j int) bool { return mx[i].Priority < mx[j].Priority })
				if conflict {
					state.Status = "mx_conflict"
				} else if found {
					recipients := make([]string, 0, len(boxes))
					for _, b := range boxes {
						recipients = append(recipients, b.Address)
					}
					active, e := s.Rules.ActiveRuleSet(ctx)
					if e != nil || active != s.Config.RuleSet {
						state.Status = "provisioning"
					} else if ready, e := s.Rules.RuleReady(ctx, ruleName(id), recipients); e != nil || !ready {
						state.Status = "provisioning"
					} else {
						state.Status = "ready"
					}
				} else {
					state.Status = "pending_mx"
				}
				now := s.Now()
				state.CheckedAt = &now
				existingMX = mx
			}
		}
	}
	return state, existingMX, s.persistDomainState(ctx, &state, existingMX)
}

func (s *Service) persistDomainState(ctx context.Context, state *DomainState, mx []MX) error {
	state.ExistingMXJSON, _ = json.Marshal(mx)
	return s.DB.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "domain_id"}}, DoUpdates: clause.AssignmentColumns([]string{"status", "detail", "existing_mx_json", "checked_at", "updated_at"})}).Create(state).Error
}
