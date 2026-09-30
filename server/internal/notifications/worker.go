// Package notifications dispatches account/security and sending-alert emails
// through platform SMTP, independently of tenant sending and marketing jobs.
package notifications

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"kori/internal/models"
	"kori/internal/onboardingemails"
)

type Sender func(context.Context, string, string, string, string, string) (bool, error)
type Worker struct {
	DB         *gorm.DB
	BaseURL    string
	Send       Sender
	Suppressed func(context.Context, string, string) (bool, error)
	Now        func() time.Time
}

func New(db *gorm.DB, base string, send Sender) (*Worker, error) {
	u, err := url.Parse(base)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
		return nil, fmt.Errorf("service emails require an HTTPS DASHBOARD_URL")
	}
	u.Path, u.RawPath, u.RawQuery, u.Fragment = "", "", "", ""
	return &Worker{DB: db, BaseURL: u.String(), Send: send, Now: time.Now}, nil
}

var errSkip = errors.New("notice is no longer eligible")

func (w *Worker) recipient(ctx context.Context, n models.ServiceNotice) (string, onboardingemails.ServiceData, string, error) {
	d := onboardingemails.ServiceData{Time: n.CreatedAt.UTC().Format("02 Jan 2006, 15:04 MST"), Count: fmt.Sprint(n.EventCount)}
	var team models.Team
	if err := w.DB.WithContext(ctx).Where("id = ? AND is_deleted = false", n.TeamID).First(&team).Error; err != nil {
		return "", d, "", err
	}
	d.Workspace = team.Name
	var user models.User
	q := w.DB.WithContext(ctx).Where("team_id = ? AND is_deleted = false", n.TeamID)
	if n.UserID != "" {
		q = q.Where("id = ?", n.UserID)
	} else {
		q = q.Where("role IN ?", []models.UserRole{models.UserRoleAdmin, models.UserRoleSuperAdmin})
		if team.OwnerUserID != "" {
			q = q.Where("id = ?", team.OwnerUserID)
		} else {
			q = q.Order("created_at, id")
		}
	}
	if err := q.First(&user).Error; err != nil {
		return "", d, "", err
	}
	if n.UserID != "" && !strings.EqualFold(user.Email, n.AccountEmail) {
		return "", d, "", errSkip
	}
	address, err := mail.ParseAddress(user.Email)
	if err != nil || strings.ContainsAny(user.Email, "\r\n") {
		return "", d, "", errSkip
	}
	address.Address = strings.ToLower(address.Address)
	d.Name, d.Email = user.FirstName, address.Address
	template, ok := onboardingemails.FindService(n.Kind)
	if !ok {
		return "", d, "", errSkip
	}
	link := w.BaseURL + template.Path
	if n.Kind == "password_reset" {
		var reset models.PasswordReset
		if err := w.DB.WithContext(ctx).Where("id = ? AND user_id = ? AND used = false AND is_deleted = false AND expires_at > ?", n.ReferenceID, user.ID, w.Now().Add(30*time.Second)).First(&reset).Error; err != nil {
			return "", d, "", err
		}
		link = w.BaseURL + "/auth/reset-password/" + url.PathEscape(reset.Code)
	}
	// Account security mail is transactional and must not follow marketing opt-outs.
	// Workspace delivery alerts honor suppression to avoid repeated bounces.
	if !template.Account && w.Suppressed != nil {
		blocked, err := w.Suppressed(ctx, n.TeamID, address.Address)
		if err != nil {
			return "", d, "", err
		}
		if blocked {
			return "", d, "", errSkip
		}
	}
	return address.Address, d, link, nil
}

func (w *Worker) ProcessOne(ctx context.Context) error {
	if w.Send == nil {
		return nil
	}
	now := w.Now()
	if err := w.DB.WithContext(ctx).Model(&models.ServiceNotice{}).Where("status = ? AND updated_at < ?", "SENDING", now.Add(-5*time.Minute)).Update("status", "DELIVERY_UNKNOWN").Error; err != nil {
		return err
	}
	var n models.ServiceNotice
	err := w.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where("status = ? AND next_attempt_at <= ?", "QUEUED", now).Order("CASE WHEN kind = 'password_reset' THEN 0 WHEN kind = 'password_changed' THEN 1 ELSE 2 END, created_at, key").First(&n).Error; err != nil {
			return err
		}
		r := tx.Model(&models.ServiceNotice{}).Where("key = ? AND status = ?", n.Key, "QUEUED").Updates(map[string]any{"status": "SENDING", "attempts": n.Attempts + 1, "updated_at": now})
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		return err
	}
	status := "SKIPPED"
	if n.CreatedAt.After(now.Add(-48 * time.Hour)) {
		recipient, data, link, err := w.recipient(ctx, n)
		if err == nil {
			t, _ := onboardingemails.FindService(n.Kind)
			html, text := t.Render(data, link)
			sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			uncertain, sendErr := w.Send(sendCtx, n.Key, recipient, t.Subject, html, text)
			cancel()
			status = "ACCEPTED"
			if sendErr != nil {
				status = "QUEUED"
				if uncertain {
					status = "DELIVERY_UNKNOWN"
				}
			}
		} else if !errors.Is(err, errSkip) && !errors.Is(err, gorm.ErrRecordNotFound) {
			status = "QUEUED"
		}
	}
	if status == "QUEUED" && n.Attempts >= 4 {
		status = "FAILED"
	}
	record, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	return w.DB.WithContext(record).Model(&models.ServiceNotice{}).Where("key = ? AND status = ?", n.Key, "SENDING").Updates(map[string]any{"status": status, "next_attempt_at": now.Add(time.Minute * time.Duration(1<<min(n.Attempts, 4))), "updated_at": w.Now()}).Error
}

func (w *Worker) Run(ctx context.Context) {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if err := w.ProcessOne(ctx); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Print("service email worker failed; inspect outbox and SMTP availability")
			}
		}
	}
}
