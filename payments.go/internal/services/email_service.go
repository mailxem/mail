package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dodopayments/dodopayments-go"
	"github.com/mailxem/payments.go/internal/config"
	"github.com/mailxem/payments.go/internal/models"
)

type EmailService interface {
	SendEmail(ctx context.Context, to string, subject string, templateID *string, html *string, data *map[string]any, scheduleAt *time.Time) error
	ScheduleFreeTrialEmails(ctx context.Context, subscription *models.Subscription, trialEnd time.Time) error
	SendFreeTrialStartEmail(ctx context.Context, customerEmail, customerName, successURL string, plan *models.PlanResponse, trialEnd time.Time) error
	SendFreeTrialEndingSoonEmail(ctx context.Context, customerEmail, customerName, successURL string, plan *models.PlanResponse, trialEnd time.Time) error
	SendFreeTrialEndedEmail(ctx context.Context, customerEmail, customerName, successURL string, plan *models.PlanResponse, trialEnd time.Time) error
	sendReminderEmail(ctx context.Context, subscription *models.Subscription) error
}

type emailService struct {
	cfg         *config.Config
	teamService TeamService
	planService PlanService
}

func NewEmailService(cfg *config.Config, teamService TeamService, planService PlanService) EmailService {
	return &emailService{cfg: cfg, teamService: teamService, planService: planService}
}

func (s *emailService) SendFreeTrialStartEmail(ctx context.Context, customerEmail, customerName, successURL string, plan *models.PlanResponse, trialEnd time.Time) error {
	templateID := s.cfg.Templates.FreeTrialStart

	return s.SendEmail(ctx, customerEmail, "Your 14-Day Free Pro Trial has Started! Unlock More with mailxem 🚀", &templateID, nil, &map[string]any{
		"name":         customerName,
		"upgrade_url":  successURL,
		"trial_period": plan.TrialPeriodDays,
		"trial_end":    trialEnd.Format("2006-01-02"),
	}, nil)
}

func (s *emailService) SendFreeTrialEndingSoonEmail(ctx context.Context, customerEmail, customerName, successURL string, plan *models.PlanResponse, trialEnd time.Time) error {
	trialEndMinusOneHalfWeek := trialEnd.AddDate(0, 0, -3)
	templateID := s.cfg.Templates.FreeTrialEndIsGoingToEndSoon

	return s.SendEmail(ctx, customerEmail, "Your 14-Day Free Pro Trial is going to end soon! Unlock More with mailxem 🚀", &templateID, nil, &map[string]any{
		"name":         customerName,
		"upgrade_url":  successURL,
		"trial_period": plan.TrialPeriodDays,
		"trial_end":    trialEnd.Format("2006-01-02"),
	}, &trialEndMinusOneHalfWeek)
}

func (s *emailService) SendFreeTrialEndedEmail(ctx context.Context, customerEmail, customerName, successURL string, plan *models.PlanResponse, trialEnd time.Time) error {
	trialEndedTemplateID := s.cfg.Templates.FreeTrialEnded
	trialEnded := trialEnd.Add(1 * time.Hour)

	return s.SendEmail(ctx, customerEmail, "Your 14-Day Free Pro Trial has ended! Unlock More with mailxem 🚀", &trialEndedTemplateID, nil, &map[string]any{
		"name":         customerName,
		"upgrade_url":  successURL,
		"trial_period": plan.TrialPeriodDays,
		"text":         "Don't let your progress pause—stay ahead with the power of Scale. Just click the button below to upgrade anytime:",
		"trial_end":    trialEnd.Format("2006-01-02"),
	}, &trialEnded)
}

func (s *emailService) ScheduleFreeTrialEmails(ctx context.Context, subscription *models.Subscription, trialEnd time.Time) error {
	if subscription.PlanID == s.cfg.Plans.FreePlanID {
		return nil
	}

	if subscription.TrialEndsAt == nil {
		return fmt.Errorf("trial ends at is not set or plan is not enabled for trials")
	}

	team, err := s.teamService.GetByID(ctx, subscription.TeamID)
	if err != nil {
		return fmt.Errorf("team not found: %w", err)
	}

	plan, err := s.planService.GetPlan(ctx, subscription.PlanID)
	if err != nil {
		return fmt.Errorf("plan not found: %w", err)
	}

	// call line by line
	s.SendFreeTrialStartEmail(ctx, team.Email, team.Name, s.cfg.DashboardURL, plan, trialEnd)
	s.SendFreeTrialEndingSoonEmail(ctx, team.Email, team.Name, s.cfg.DashboardURL, plan, trialEnd)
	s.SendFreeTrialEndedEmail(ctx, team.Email, team.Name, s.cfg.DashboardURL, plan, trialEnd)

	return nil
}

func (s *emailService) sendReminderEmail(ctx context.Context, subscription *models.Subscription) error {
	if subscription.PlanID == s.cfg.Plans.FreePlanID {
		return nil
	}

	templateID := s.cfg.Templates.SubscriptionStatusChanged

	if subscription.Status == dodopayments.SubscriptionStatusActive {
		templateID = s.cfg.Templates.ProPlanActivated
	}

	team, err := s.teamService.GetByID(ctx, subscription.TeamID)
	if err != nil {
		return fmt.Errorf("team not found: %w", err)
	}

	subject := "Your team's subscription has been " + string(subscription.Status)

	if subscription.Status == dodopayments.SubscriptionStatusActive {
		subject = "🎉 Welcome to Scale " + string(subscription.BillingCycle) + " – You're officially upgraded!"
	}

	s.SendEmail(ctx, team.Email, subject, &templateID, nil, &map[string]any{
		"name":            team.Name,
		"subscription_id": subscription.ID,
		"status":          subscription.Status,
		"cancelled_at":    time.Now().Format(time.RFC3339),
	}, nil)

	return nil
}

func (s *emailService) SendEmail(ctx context.Context, to string, subject string, templateID *string, html *string, data *map[string]any, scheduleAt *time.Time) error {
	if s.cfg.KoriEmailAPIURL == "" {
		return nil
	}

	if templateID == nil && html == nil {
		return fmt.Errorf("templateID or html is required")
	}

	var v map[string]any = make(map[string]any)
	if data != nil {
		v["data"] = *data
	}

	if templateID != nil {
		v["templateId"] = *templateID
	}

	if html != nil {
		v["html"] = *html
	}

	v["to"] = to
	v["subject"] = subject
	v["provider"] = s.cfg.KoriEmailProvider
	v["scheduleAt"] = scheduleAt

	jsonBody, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	request, err := http.NewRequest("POST", s.cfg.KoriEmailAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-api-key", s.cfg.KoriEmailAPIKey)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send email: %s", response.Status)
	}

	return nil
}
