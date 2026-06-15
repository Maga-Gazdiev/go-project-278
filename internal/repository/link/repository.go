package link

import (
	"context"
	stderrors "errors"

	generated "project-3/db/generated"
	apperrors "project-3/internal/errors"
	"project-3/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
	queries *generated.Queries
}

func New(database generated.DBTX) *Repository {
	return &Repository{
		queries: generated.New(database),
	}
}

func (r *Repository) GetByID(ctx context.Context, id int64) (model.Link, error) {
	link, err := r.queries.GetLink(ctx, id)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return model.Link{}, apperrors.ErrLinkNotFound
		}
		return model.Link{}, err
	}

	return toModel(link), nil
}

func (r *Repository) GetByShortName(ctx context.Context, shortName string) (model.Link, error) {
	link, err := r.queries.GetLinkByShortName(ctx, shortName)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return model.Link{}, apperrors.ErrLinkNotFound
		}
		return model.Link{}, err
	}

	return toModel(link), nil
}

func (r *Repository) List(ctx context.Context, from int, to int) ([]model.Link, error) {
	limit := to - from + 1

	links, err := r.queries.ListLinks(ctx, generated.ListLinksParams{
		Limit:  int32(limit),
		Offset: int32(from),
	})
	if err != nil {
		return nil, err
	}

	result := make([]model.Link, 0, len(links))
	for _, link := range links {
		result = append(result, toModel(link))
	}

	return result, nil
}

func (r *Repository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountLinks(ctx)
}

func (r *Repository) Create(ctx context.Context, link model.Link) (model.Link, error) {
	createdLink, err := r.queries.CreateLink(ctx, generated.CreateLinkParams{
		OriginalUrl: link.OriginalUrl,
		ShortName:   link.ShortName,
		ShortUrl:    link.ShortUrl,
	})
	if err != nil {
		return model.Link{}, mapConstraintError(err, apperrors.ErrLinkNotCreated)
	}

	return toModel(createdLink), nil
}

func (r *Repository) Update(ctx context.Context, link model.Link) (model.Link, error) {
	updatedLink, err := r.queries.UpdateLink(ctx, generated.UpdateLinkParams{
		ID:          link.ID,
		OriginalUrl: link.OriginalUrl,
		ShortName:   link.ShortName,
		ShortUrl:    link.ShortUrl,
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return model.Link{}, apperrors.ErrLinkNotFound
		}
		return model.Link{}, mapConstraintError(err, apperrors.ErrLinkNotUpdated)
	}

	return toModel(updatedLink), nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	rowsAffected, err := r.queries.DeleteLink(ctx, id)
	if err != nil {
		return apperrors.ErrLinkNotDeleted
	}
	if rowsAffected == 0 {
		return apperrors.ErrLinkNotFound
	}

	return nil
}

func mapConstraintError(err error, fallback error) error {
	var pgErr *pgconn.PgError
	if stderrors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperrors.ErrShortNameTaken
	}

	return fallback
}

func toModel(link generated.Link) model.Link {
	return model.Link{
		ID:          link.ID,
		OriginalUrl: link.OriginalUrl,
		ShortName:   link.ShortName,
		ShortUrl:    link.ShortUrl,
	}
}
