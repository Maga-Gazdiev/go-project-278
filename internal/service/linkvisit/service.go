package linkvisit

import (
	"context"

	apperrors "project-3/internal/errors"
	"project-3/internal/model"
)

type RepositoryInterface interface {
	Create(ctx context.Context, visit model.LinkVisit) (model.LinkVisit, error)
	List(ctx context.Context, from int, to int) ([]model.LinkVisit, error)
	Count(ctx context.Context) (int64, error)
}

type Service struct {
	repository RepositoryInterface
}

func NewService(repository RepositoryInterface) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, visit model.LinkVisit) (model.LinkVisit, error) {
	return s.repository.Create(ctx, visit)
}

func (s *Service) List(ctx context.Context, from int, to int) ([]model.LinkVisit, int64, error) {
	if from < 0 || to < 0 || from > to {
		return nil, 0, apperrors.ErrNotValidQuery
	}

	visits, err := s.repository.List(ctx, from, to)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repository.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return visits, total, nil
}
