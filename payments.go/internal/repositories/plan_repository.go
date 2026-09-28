package repositories

import (
	"context"
	"fmt"

	"github.com/dodopayments/dodopayments-go"
	"github.com/mailxem/payments.go/internal/models"
	"gorm.io/gorm"
)

type PlanRepository interface {
	Create(ctx context.Context, plan *models.Plan) error
	GetByID(ctx context.Context, id string) (*models.Plan, error)
	GetByDodoPlanID(ctx context.Context, dodoPlanID string) (*models.Plan, error)
	Update(ctx context.Context, plan *models.Plan) error
	Delete(ctx context.Context, id string) error
	GetWithSubscriptionCount(ctx context.Context, id string) (*models.Plan, int64, error)
	ListWithSubscriptionCounts(ctx context.Context, includeInactive bool, isAdmin bool) ([]*models.PlanResponse, error)
}

type planRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) PlanRepository {
	return &planRepository{db: db}
}

func (r *planRepository) Create(ctx context.Context, plan *models.Plan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *planRepository) GetByID(ctx context.Context, id string) (*models.Plan, error) {
	var plan models.Plan
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&plan).Error
	if err != nil {
		return nil, err
	}

	plan.Features = []models.PlanFeatures{}
	plan.TrailFeatures = []models.PlanFeatures{}

	err = r.db.WithContext(ctx).Where("plan_id = ? AND trail = ?", plan.ID, false).Find(&plan.Features).Error
	if err != nil {
		return nil, err
	}
	err = r.db.WithContext(ctx).Where("plan_id = ? AND trail = ?", plan.ID, true).Find(&plan.TrailFeatures).Error
	if err != nil {
		return nil, err
	}

	return &plan, nil
}

func (r *planRepository) GetByDodoPlanID(ctx context.Context, dodoPlanID string) (*models.Plan, error) {
	var plan models.Plan
	err := r.db.WithContext(ctx).Where("dodo_plan_id = ?", dodoPlanID).First(&plan).Error
	if err != nil {
		return nil, err
	}
	plan.Features = []models.PlanFeatures{}
	plan.TrailFeatures = []models.PlanFeatures{}

	err = r.db.WithContext(ctx).Where("plan_id = ? AND trail = ?", plan.ID, false).Find(&plan.Features).Error
	if err != nil {
		return nil, err
	}
	err = r.db.WithContext(ctx).Where("plan_id = ? AND trail = ?", plan.ID, true).Find(&plan.TrailFeatures).Error
	if err != nil {
		return nil, err
	}

	return &plan, nil
}

func (r *planRepository) Update(ctx context.Context, plan *models.Plan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *planRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Plan{}, "id = ?", id).Error
}

func (r *planRepository) GetWithSubscriptionCount(ctx context.Context, id string) (*models.Plan, int64, error) {
	var plan models.Plan
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&plan).Error
	if err != nil {
		return nil, 0, err
	}

	var count int64
	err = r.db.WithContext(ctx).Model(&models.Subscription{}).
		Where("plan_id = ? AND status IN ?", plan.ID, []dodopayments.SubscriptionStatus{
			dodopayments.SubscriptionStatusActive,
			dodopayments.SubscriptionStatusPending,
		}).
		Count(&count).Error

	return &plan, count, err
}
func (r *planRepository) ListWithSubscriptionCounts(ctx context.Context, includeInactive bool, isAdmin bool) ([]*models.PlanResponse, error) {
	var plans []*models.Plan
	query := r.db.WithContext(ctx)

	if !includeInactive {
		query = query.Where("status = ?", models.PlanStatusActive)
	}

	err := query.Order("display_order ASC, created_at ASC").Find(&plans).Error
	if err != nil {
		return nil, err
	}

	for _, plan := range plans {
		plan.Features = []models.PlanFeatures{}
		plan.TrailFeatures = []models.PlanFeatures{}
	}
	// Load features separately to avoid the scan error
	for i := range plans {
		err = r.db.WithContext(ctx).Where("plan_id = ? AND trail = ?", plans[i].ID, false).Find(&plans[i].Features).Error
		if err != nil {
			fmt.Println("error", err)
			return nil, err
		}
		err = r.db.WithContext(ctx).Where("plan_id = ? AND trail = ?", plans[i].ID, true).Find(&plans[i].TrailFeatures).Error
		if err != nil {
			fmt.Println("error", err)
			return nil, err
		}
	}

	responses := make([]*models.PlanResponse, 0, len(plans))

	if isAdmin {
		for _, plan := range plans {
			// Get subscription count for this plan
			var count int64
			err = r.db.WithContext(ctx).Model(&models.Subscription{}).
				Where("plan_id = ? AND status IN ?", plan.ID, []dodopayments.SubscriptionStatus{
					dodopayments.SubscriptionStatusActive,
					dodopayments.SubscriptionStatusPending,
				}).
				Count(&count).Error
			if err != nil {
				return nil, err
			}

			response := plan.ToResponse()
			response.ActiveSubscriptions = int(count)
			responses = append(responses, response)
		}
	} else {
		for _, plan := range plans {
			response := plan.ToResponse()
			responses = append(responses, response)
		}
	}

	return responses, nil
}

// PlanFeaturesRepository handles operations for plan features
type PlanFeaturesRepository interface {
	Create(ctx context.Context, feature *models.PlanFeatures) error
	GetByPlanID(ctx context.Context, planID string) ([]*models.PlanFeatures, error)
	UpdateQuota(ctx context.Context, id string, quota int) error
	Delete(ctx context.Context, id string) error
	DeleteByPlanID(ctx context.Context, planID string) error
}

type planFeaturesRepository struct {
	db *gorm.DB
}

func NewPlanFeaturesRepository(db *gorm.DB) PlanFeaturesRepository {
	return &planFeaturesRepository{db: db}
}

func (r *planFeaturesRepository) Create(ctx context.Context, feature *models.PlanFeatures) error {
	return r.db.WithContext(ctx).Create(feature).Error
}

func (r *planFeaturesRepository) GetByPlanID(ctx context.Context, planID string) ([]*models.PlanFeatures, error) {
	var features []*models.PlanFeatures
	err := r.db.WithContext(ctx).Where("plan_id = ?", planID).Find(&features).Error
	return features, err
}

func (r *planFeaturesRepository) UpdateQuota(ctx context.Context, id string, quota int) error {
	return r.db.WithContext(ctx).Model(&models.PlanFeatures{}).
		Where("id = ?", id).
		Update("quota", quota).Error
}

func (r *planFeaturesRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.PlanFeatures{}, "id = ?", id).Error
}

func (r *planFeaturesRepository) DeleteByPlanID(ctx context.Context, planID string) error {
	return r.db.WithContext(ctx).Where("plan_id = ?", planID).Delete(&models.PlanFeatures{}).Error
}
