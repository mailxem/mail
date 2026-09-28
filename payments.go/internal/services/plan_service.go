package services

import (
	"context"
	"fmt"

	"github.com/mailxem/payments.go/internal/models"
	"github.com/mailxem/payments.go/internal/repositories"
)

type PlanService interface {
	CreatePlan(ctx context.Context, req *models.CreatePlanRequest) (*models.PlanResponse, error)
	UpdatePlan(ctx context.Context, planID string, req *models.UpdatePlanRequest) (*models.PlanResponse, error)
	GetPlan(ctx context.Context, planID string) (*models.PlanResponse, error)
	GetPlanByDodoID(ctx context.Context, dodoPlanID string) (*models.PlanResponse, error)
	ListPlans(ctx context.Context, includeInactive bool) ([]*models.PlanResponse, error)
	ListPublicPlans(ctx context.Context) ([]*models.PlanResponse, error)
	DeletePlan(ctx context.Context, planID string) error
	ArchivePlan(ctx context.Context, planID string) error
	GetPlanWithStats(ctx context.Context, planID string) (*models.PlanResponse, error)
}

type planService struct {
	planRepo         repositories.PlanRepository
	planFeaturesRepo repositories.PlanFeaturesRepository
}

func NewPlanService(planRepo repositories.PlanRepository, planFeaturesRepo repositories.PlanFeaturesRepository) PlanService {
	return &planService{
		planRepo:         planRepo,
		planFeaturesRepo: planFeaturesRepo,
	}
}

func (s *planService) CreatePlan(ctx context.Context, req *models.CreatePlanRequest) (*models.PlanResponse, error) {
	// Create plan
	plan := &models.Plan{
		DodoPlanID:      req.DodoPlanID,
		Name:            req.Name,
		Description:     req.Description,
		Price:           req.Price,
		Currency:        req.Currency,
		Status:          models.PlanStatusActive,
		TrialPeriodDays: req.TrialPeriodDays,
		TrialEnabled:    req.TrialEnabled,
		DisplayOrder:    req.DisplayOrder,
		IsPopular:       req.IsPopular,
		IsEnterprise:    req.IsEnterprise,
		Metadata:        req.Metadata,
	}

	// Create plan in database
	if err := s.planRepo.Create(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	return plan.ToResponse(), nil
}

func (s *planService) UpdatePlan(ctx context.Context, planID string, req *models.UpdatePlanRequest) (*models.PlanResponse, error) {
	// Get existing plan
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		plan.Name = *req.Name
	}
	if req.Description != nil {
		plan.Description = req.Description
	}
	if req.Status != nil {
		plan.Status = *req.Status
	}
	if req.Price != nil {
		plan.Price = *req.Price
	}
	if req.Currency != nil {
		plan.Currency = *req.Currency
	}
	if req.TrialPeriodDays != nil {
		plan.TrialPeriodDays = *req.TrialPeriodDays
	}
	if req.TrialEnabled != nil {
		plan.TrialEnabled = *req.TrialEnabled
	}
	if req.DisplayOrder != nil {
		plan.DisplayOrder = *req.DisplayOrder
	}
	if req.IsPopular != nil {
		plan.IsPopular = *req.IsPopular
	}
	if req.IsEnterprise != nil {
		plan.IsEnterprise = *req.IsEnterprise
	}
	if req.Metadata != nil {
		plan.Metadata = req.Metadata
	}

	// Save updated plan
	if err := s.planRepo.Update(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to update plan: %w", err)
	}

	return plan.ToResponse(), nil
}

func (s *planService) GetPlan(ctx context.Context, planID string) (*models.PlanResponse, error) {
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	return plan.ToResponse(), nil
}

func (s *planService) GetPlanByDodoID(ctx context.Context, dodoPlanID string) (*models.PlanResponse, error) {
	plan, err := s.planRepo.GetByDodoPlanID(ctx, dodoPlanID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	return plan.ToResponse(), nil
}

func (s *planService) ListPlans(ctx context.Context, includeInactive bool) ([]*models.PlanResponse, error) {
	return s.planRepo.ListWithSubscriptionCounts(ctx, includeInactive, true)
}

func (s *planService) ListPublicPlans(ctx context.Context) ([]*models.PlanResponse, error) {
	// Only return active plans for public consumption
	return s.planRepo.ListWithSubscriptionCounts(ctx, false, false)
}

func (s *planService) DeletePlan(ctx context.Context, planID string) error {
	// Check if plan exists
	_, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return fmt.Errorf("plan not found: %w", err)
	}

	// Check if plan has active subscriptions
	_, subscriptionCount, err := s.planRepo.GetWithSubscriptionCount(ctx, planID)
	if err != nil {
		return fmt.Errorf("failed to check subscription count: %w", err)
	}

	if subscriptionCount > 0 {
		return fmt.Errorf("cannot delete plan with %d active subscriptions", subscriptionCount)
	}

	// Delete the plan
	if err := s.planRepo.Delete(ctx, planID); err != nil {
		return fmt.Errorf("failed to delete plan: %w", err)
	}

	return nil
}

func (s *planService) ArchivePlan(ctx context.Context, planID string) error {
	// Get existing plan
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return fmt.Errorf("plan not found: %w", err)
	}

	// Set status to archived
	plan.Status = models.PlanStatusArchived

	// Save to database
	if err := s.planRepo.Update(ctx, plan); err != nil {
		return fmt.Errorf("failed to archive plan: %w", err)
	}

	return nil
}

func (s *planService) GetPlanWithStats(ctx context.Context, planID string) (*models.PlanResponse, error) {
	plan, subscriptionCount, err := s.planRepo.GetWithSubscriptionCount(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	response := plan.ToResponse()
	response.ActiveSubscriptions = int(subscriptionCount)

	return response, nil
}
