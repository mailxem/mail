package sending

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
	"kori/internal/models"
)

// RenderRaw rebuilds the MIME of an in-process email (utils.RenderRaw). Those
// messages are queued without their MIME, which is only needed at dispatch.
var RenderRaw func(ctx context.Context, emailID string) ([]byte, error)

func (s *Service) SubmitEmail(ctx context.Context, email *models.Email, raw []byte) error {
	domainID, err := s.emailDomain(ctx, email)
	if err != nil {
		return err
	}
	ref, kind := s.feedbackRef(ctx, email.TeamID, email.ID, email.CampaignID)
	_, err = s.Submit(ctx, emailSubmission(email, domainID, raw, ref, kind))
	return err
}

// SubmitEmails is SubmitEmail for a page of one team's emails, reserving
// quota for the page at once (SubmitBatch). It returns each email's error.
func (s *Service) SubmitEmails(ctx context.Context, emails []*models.Email, raws [][]byte) []error {
	errs := make([]error, len(emails))
	type resolved struct {
		id  string
		err error
	}
	domains := map[string]resolved{} // a page shares its sender
	refs := map[string][2]string{}   // and its campaign
	var subs []Submission
	var at []int
	for i, email := range emails {
		k := email.TeamID + ":" + email.SMTPConfigID + ":" + email.From
		d, ok := domains[k]
		if !ok {
			d.id, d.err = s.emailDomain(ctx, email)
			domains[k] = d
		}
		if d.err != nil {
			errs[i] = d.err
			continue
		}
		ref, ok := refs[email.CampaignID]
		if !ok {
			ref[0], ref[1] = s.feedbackRef(ctx, email.TeamID, email.ID, email.CampaignID)
			if email.CampaignID != "" {
				refs[email.CampaignID] = ref
			}
		}
		subs = append(subs, emailSubmission(email, d.id, raws[i], ref[0], ref[1]))
		at = append(at, i)
	}
	for j, r := range s.SubmitBatch(ctx, subs) {
		errs[at[j]] = r.Err
	}
	return errs
}

func emailSubmission(email *models.Email, domainID string, raw []byte, ref, kind string) Submission {
	recipients := []string{email.To}
	for _, list := range []string{email.CC, email.BCC} {
		if list != "" {
			recipients = append(recipients, strings.Split(list, ",")...)
		}
	}
	return Submission{TeamID: email.TeamID, DomainID: domainID, EmailID: email.ID, Key: "email-" + email.ID, From: email.From, Recipients: recipients, Raw: raw, Marketing: email.CampaignID != "" || email.ContactID != "", FeedbackRef: ref, FeedbackKind: kind, Transient: RenderRaw != nil}
}

// feedbackRef is the first Feedback-ID field of an in-process email: its
// newsletter, campaign or own ID.
func (s *Service) feedbackRef(ctx context.Context, team, emailID, campaignID string) (string, string) {
	if campaignID == "" {
		return emailID, "email"
	}
	var c models.Campaign
	if s.DB.WithContext(ctx).Select("newsletter_id").First(&c, "id = ? AND team_id = ?", campaignID, team).Error == nil && c.NewsletterID != nil {
		return *c.NewsletterID, "newsletter"
	}
	return campaignID, "campaign"
}

// emailDomain finds the managed domain an email's sender may use.
func (s *Service) emailDomain(ctx context.Context, email *models.Email) (string, error) {
	var d Domain
	var managedMailbox struct {
		DomainID, Address, Status string
		Active                    bool
	}
	managedQuery := &gorm.DB{Error: gorm.ErrRecordNotFound}
	if s.DB.Migrator().HasTable("managed_receiving_mailboxes") {
		managedQuery = s.DB.WithContext(ctx).Table("managed_receiving_mailboxes").Select("domain_id,address,status,active").Where("smtp_config_id = ? AND team_id = ?", email.SMTPConfigID, email.TeamID).Take(&managedMailbox)
	}
	if managedQuery.Error == nil {
		// A full mailbox cannot receive, but replies from it still send.
		if !managedMailbox.Active || (managedMailbox.Status != "active" && managedMailbox.Status != "storage_full") || email.From != managedMailbox.Address {
			return "", ErrDenied
		}
		if e := s.DB.WithContext(ctx).First(&d, "id = ? AND team_id = ? AND ready = true AND ownership = true AND provisioned = true", managedMailbox.DomainID, email.TeamID).Error; e != nil {
			return "", ErrDenied
		}
		var account Account
		if e := s.DB.WithContext(ctx).First(&account, "team_id = ? AND approved = true AND paused = false AND suspended = false", email.TeamID).Error; e != nil {
			return "", ErrDenied
		}
	} else {
		if !errors.Is(managedQuery.Error, gorm.ErrRecordNotFound) {
			return "", ErrDenied
		}
		if e := s.DB.WithContext(ctx).First(&d, "smtp_config_id = ? AND team_id = ? AND ready = true", email.SMTPConfigID, email.TeamID).Error; e != nil {
			return "", ErrDenied
		}
	}
	return d.ID, nil
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
