package sending

import (
	"context"
	"errors"
	"strings"

	"kori/internal/models"
)

func (s *Service) SubmitEmail(ctx context.Context, email *models.Email, raw []byte) error {
	var d Domain
	if e := s.DB.WithContext(ctx).First(&d, "smtp_config_id = ? AND team_id = ? AND ready = true", email.SMTPConfigID, email.TeamID).Error; e != nil {
		return ErrDenied
	}
	recipients := []string{email.To}
	for _, list := range []string{email.CC, email.BCC} {
		if list != "" {
			recipients = append(recipients, strings.Split(list, ",")...)
		}
	}
	_, err := s.Submit(ctx, Submission{TeamID: email.TeamID, DomainID: d.ID, EmailID: email.ID, Key: "email-" + email.ID, From: email.From, Recipients: recipients, Raw: raw, Marketing: email.CampaignID != "" || email.ContactID != ""})
	return err
}
func (s *Service) RecipientPolicy(ctx context.Context, team string, recipients []string) error {
	for _, r := range recipients {
		a, e := address(r)
		if e != nil {
			return e
		}
		blocked, e := Suppressed(s.DB.WithContext(ctx), team, a, true)
		if e != nil {
			return errors.New("recipient policy unavailable")
		}
		if blocked {
			return ErrSuppressed
		}
	}
	return nil
}
