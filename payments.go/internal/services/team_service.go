package services

import (
	"context"

	"github.com/mailxem/payments.go/internal/models"
	"github.com/mailxem/payments.go/internal/repositories"
)

type TeamService interface {
	GetByID(ctx context.Context, id string) (*models.Team, error)
}

type teamService struct {
	teamRepo repositories.TeamRepository
}

func NewTeamService(teamRepo repositories.TeamRepository) TeamService {
	return &teamService{teamRepo: teamRepo}
}

func (s *teamService) GetByID(ctx context.Context, id string) (*models.Team, error) {
	return s.teamRepo.GetByID(ctx, id)
}
