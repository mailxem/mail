package repositories

import (
	"context"

	"github.com/mailxem/payments.go/internal/models"
	"gorm.io/gorm"
)

type TeamRepository interface {
	GetByID(ctx context.Context, id string) (*models.Team, error)
}

type teamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) GetByID(ctx context.Context, id string) (*models.Team, error) {
	var team models.Team
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&team).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}
