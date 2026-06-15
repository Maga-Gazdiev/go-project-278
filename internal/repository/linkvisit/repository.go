package linkvisit

import (
	"context"

	generated "project-3/db/generated"
	"project-3/internal/model"
)

type Repository struct {
	queries *generated.Queries
}

func New(database generated.DBTX) *Repository {
	return &Repository{
		queries: generated.New(database),
	}
}

func (r *Repository) Create(ctx context.Context, visit model.LinkVisit) (model.LinkVisit, error) {
	createdVisit, err := r.queries.CreateLinkVisit(ctx, generated.CreateLinkVisitParams{
		LinkID:    visit.LinkID,
		Ip:        visit.IP,
		UserAgent: visit.UserAgent,
		Referer:   visit.Referer,
		Status:    int32(visit.Status),
	})
	if err != nil {
		return model.LinkVisit{}, err
	}

	return toModel(createdVisit), nil
}

func (r *Repository) List(ctx context.Context, from int, to int) ([]model.LinkVisit, error) {
	limit := to - from + 1

	visits, err := r.queries.ListLinkVisits(ctx, generated.ListLinkVisitsParams{
		Limit:  int32(limit),
		Offset: int32(from),
	})
	if err != nil {
		return nil, err
	}

	result := make([]model.LinkVisit, 0, len(visits))
	for _, visit := range visits {
		result = append(result, toModel(visit))
	}

	return result, nil
}

func (r *Repository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountLinkVisits(ctx)
}

func toModel(visit generated.LinkVisit) model.LinkVisit {
	return model.LinkVisit{
		ID:        visit.ID,
		LinkID:    visit.LinkID,
		CreatedAt: visit.CreatedAt.Time,
		IP:        visit.Ip,
		UserAgent: visit.UserAgent,
		Referer:   visit.Referer,
		Status:    int(visit.Status),
	}
}
