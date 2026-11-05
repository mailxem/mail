// Package affiliates runs the Xem Cloud affiliate program: an affiliate earns
// 30% of every plan payment a referred workspace makes, for as long as it pays.
// Each commission unlocks 60 days after its payment and is paid out by bank
// transfer once unlocked commissions reach $100. Refunded or disputed payments
// earn nothing, and commissions already paid for them come off the next payout.
package affiliates

import (
	"crypto/rand"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"kori/internal/models"
)

const (
	Percent      = 30
	Hold         = 60 * 24 * time.Hour
	MinPayout    = 100_00 // in cents
	signupWindow = 24 * time.Hour
)

var ErrBelowMinimum = errors.New("unlocked commissions are below the payout minimum")

// Commission is the affiliate's share of a payment, rounded down to the cent.
func Commission(base int64) int64 { return base * Percent / 100 }

const codeAlphabet = "abcdefghjkmnpqrstuvwxyz23456789"

func newCode() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = codeAlphabet[int(b[i])%len(codeAlphabet)]
	}
	return string(b)
}

// Join makes the user an affiliate, or returns their existing affiliate account.
func Join(db *gorm.DB, userID string) (*models.Affiliate, error) {
	var a models.Affiliate
	if err := db.Where("user_id = ?", userID).First(&a).Error; err == nil {
		return &a, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	var err error
	for range 5 {
		a = models.Affiliate{ID: uuid.NewString(), UserID: userID, Code: newCode()}
		if err = db.Create(&a).Error; err == nil {
			return &a, nil
		}
		// A concurrent join for the same user wins; a code clash retries.
		if existing, findErr := find(db, "user_id = ?", userID); findErr == nil {
			return existing, nil
		}
	}
	return nil, err
}

func find(db *gorm.DB, query string, arg any) (*models.Affiliate, error) {
	var a models.Affiliate
	if err := db.Where(query, arg).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func byCode(db *gorm.DB, code string) (*models.Affiliate, error) {
	return find(db, "code = ?", strings.ToLower(strings.TrimSpace(code)))
}

// Click counts a visit through an affiliate link. Unknown codes are ignored.
func Click(db *gorm.DB, code string) error {
	a, err := byCode(db, code)
	if err != nil {
		return nil
	}
	return db.Create(&models.AffiliateClick{AffiliateID: a.ID}).Error
}

// Attribute credits a new workspace to the affiliate whose link it came
// through. Only workspaces created in the last day can be attributed, each only
// once, and never to an affiliate in the same workspace. Anything else is
// silently ignored so the endpoint can't be used to probe codes.
func Attribute(db *gorm.DB, teamID, userID, code string, now time.Time) error {
	a, err := byCode(db, code)
	if err != nil || a.UserID == userID {
		return nil
	}
	var team models.Team
	if err := db.Select("id", "created_at").Where("id = ?", teamID).First(&team).Error; err != nil || now.Sub(team.CreatedAt) > signupWindow {
		return nil
	}
	var same int64
	if err := db.Model(&models.User{}).Where("id = ? AND team_id = ?", a.UserID, teamID).Count(&same).Error; err != nil {
		return err
	}
	if same > 0 {
		return nil
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&models.AffiliateReferral{TeamID: teamID, AffiliateID: a.ID}).Error
}

// Payment is a successful plan payment, in the settlement currency's smallest
// unit, after discounts and without tax.
type Payment struct {
	ID       string    `json:"payment_id"`
	TeamID   string    `json:"team_id"`
	Base     int64     `json:"base"`
	Currency string    `json:"currency"`
	PaidAt   time.Time `json:"paid_at"`
}

// RecordPayment creates the commission for a referred workspace's payment.
// Payments from workspaces nobody referred are ignored; repeats are no-ops.
func RecordPayment(db *gorm.DB, p Payment) error {
	if p.ID == "" || p.TeamID == "" || p.Base <= 0 || p.Currency == "" {
		return errors.New("payment needs an id, a workspace, a positive amount and a currency")
	}
	var ref models.AffiliateReferral
	if err := db.Where("team_id = ?", p.TeamID).First(&ref).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if p.PaidAt.IsZero() {
		p.PaidAt = time.Now()
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.AffiliateCommission{
		ID: uuid.NewString(), AffiliateID: ref.AffiliateID, TeamID: p.TeamID, PaymentID: p.ID,
		Base: p.Base, Amount: Commission(p.Base), Currency: strings.ToUpper(p.Currency),
		PaidAt: p.PaidAt, UnlocksAt: p.PaidAt.Add(Hold),
	}).Error
}

// Reverse cancels the commission for a refunded or disputed payment.
func Reverse(db *gorm.DB, paymentID string, now time.Time) error {
	return db.Model(&models.AffiliateCommission{}).
		Where("payment_id = ? AND reversed_at IS NULL", paymentID).
		Update("reversed_at", now).Error
}

// Status is where a commission stands: pending, unlocked, paid or reversed.
func Status(c models.AffiliateCommission, now time.Time) string {
	switch {
	case c.ReversedAt != nil:
		return "reversed"
	case c.PayoutID != nil:
		return "paid"
	case !now.Before(c.UnlocksAt):
		return "unlocked"
	default:
		return "pending"
	}
}

// Line is one commission as the affiliate sees it; the customer stays anonymous.
type Line struct {
	Customer  string    `json:"customer"`
	Amount    int64     `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	PaidAt    time.Time `json:"paidAt"`
	UnlocksAt time.Time `json:"unlocksAt"`
}

type Summary struct {
	Code        string           `json:"code"`
	Clicks      int64            `json:"clicks"`
	Signups     int64            `json:"signups"`
	Customers   int              `json:"customers"`
	Currency    string           `json:"currency"`
	Totals      map[string]int64 `json:"totals"`
	Commissions []Line           `json:"commissions"`
}

// Summarize returns the affiliate's link stats, totals by status and recent commissions.
func Summarize(db *gorm.DB, a *models.Affiliate, now time.Time) (*Summary, error) {
	s := &Summary{Code: a.Code, Currency: "USD", Totals: map[string]int64{"pending": 0, "unlocked": 0, "paid": 0, "reversed": 0}, Commissions: []Line{}}
	if err := db.Model(&models.AffiliateClick{}).Where("affiliate_id = ?", a.ID).Count(&s.Clicks).Error; err != nil {
		return nil, err
	}
	var refs []models.AffiliateReferral
	if err := db.Where("affiliate_id = ?", a.ID).Order("created_at").Find(&refs).Error; err != nil {
		return nil, err
	}
	s.Signups = int64(len(refs))
	names := map[string]string{}
	for i, r := range refs {
		names[r.TeamID] = "Customer " + strconv.Itoa(i+1)
	}
	var rows []models.AffiliateCommission
	if err := db.Where("affiliate_id = ?", a.ID).Order("paid_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	paying := map[string]bool{}
	for i, c := range rows {
		st := Status(c, now)
		s.Totals[st] += c.Amount
		s.Currency = c.Currency
		if st != "reversed" {
			paying[c.TeamID] = true
		}
		if i < 50 {
			s.Commissions = append(s.Commissions, Line{Customer: names[c.TeamID], Amount: c.Amount, Currency: c.Currency, Status: st, PaidAt: c.PaidAt, UnlocksAt: c.UnlocksAt})
		}
	}
	s.Customers = len(paying)
	return s, nil
}

// Due is what one affiliate is owed now in one currency: unlocked commissions
// minus commissions already paid that were later reversed.
type Due struct {
	AffiliateID string `json:"affiliate_id"`
	Email       string `json:"email"`
	Currency    string `json:"currency"`
	Amount      int64  `json:"amount"`
}

func owed(db *gorm.DB, now time.Time) *gorm.DB {
	return db.Model(&models.AffiliateCommission{}).Where(
		"(payout_id IS NULL AND reversed_at IS NULL AND unlocks_at <= ?) OR (payout_id IS NOT NULL AND reversed_at IS NOT NULL AND clawback_payout_id IS NULL)", now)
}

const owedSum = "SUM(CASE WHEN payout_id IS NULL THEN amount ELSE -amount END)"

// DuePayouts lists affiliates owed at least the payout minimum.
func DuePayouts(db *gorm.DB, now time.Time) ([]Due, error) {
	var out []Due
	err := owed(db, now).
		Select("affiliate_commissions.affiliate_id, users.email, affiliate_commissions.currency, "+owedSum+" AS amount").
		Joins("JOIN affiliates ON affiliates.id = affiliate_commissions.affiliate_id").
		Joins("JOIN users ON users.id = affiliates.user_id").
		Group("affiliate_commissions.affiliate_id, users.email, affiliate_commissions.currency").
		Having(owedSum+" >= ?", MinPayout).
		Scan(&out).Error
	return out, err
}

// Pay records a bank transfer of everything the affiliate is owed in a
// currency, after the transfer has been made. It refuses amounts under the minimum.
func Pay(db *gorm.DB, affiliateID, currency string, now time.Time) (*models.AffiliatePayout, error) {
	var payout *models.AffiliatePayout
	err := db.Transaction(func(tx *gorm.DB) error {
		var rows []models.AffiliateCommission
		if err := owed(tx, now).Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("affiliate_id = ? AND currency = ?", affiliateID, strings.ToUpper(currency)).Find(&rows).Error; err != nil {
			return err
		}
		var total int64
		var credit, clawback []string
		for _, c := range rows {
			if c.PayoutID == nil {
				total += c.Amount
				credit = append(credit, c.ID)
			} else {
				total -= c.Amount
				clawback = append(clawback, c.ID)
			}
		}
		if total < MinPayout {
			return ErrBelowMinimum
		}
		payout = &models.AffiliatePayout{ID: uuid.NewString(), AffiliateID: affiliateID, Amount: total, Currency: strings.ToUpper(currency), CreatedAt: now}
		if err := tx.Create(payout).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.AffiliateCommission{}).Where("id IN ?", credit).Update("payout_id", payout.ID).Error; err != nil {
			return err
		}
		if len(clawback) > 0 {
			return tx.Model(&models.AffiliateCommission{}).Where("id IN ?", clawback).Update("clawback_payout_id", payout.ID).Error
		}
		return nil
	})
	return payout, err
}
