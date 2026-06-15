package link

import (
	"context"
	"crypto/rand"
	stderrors "errors"
	"fmt"
	"math/big"
	"strings"

	apperrors "project-3/internal/errors"
	"project-3/internal/model"
)

type LinkRepositoryInterface interface {
	GetByID(ctx context.Context, id int64) (model.Link, error)
	GetByShortName(ctx context.Context, shortName string) (model.Link, error)
	List(ctx context.Context, from int, to int) ([]model.Link, error)
	Count(ctx context.Context) (int64, error)
	Create(ctx context.Context, link model.Link) (model.Link, error)
	Update(ctx context.Context, link model.Link) (model.Link, error)
	Delete(ctx context.Context, id int64) error
}

type LinkService struct {
	repository LinkRepositoryInterface
	baseURL    string
}

func NewService(repository LinkRepositoryInterface, baseURL string) *LinkService {
	return &LinkService{
		repository: repository,
		baseURL:    redirectBaseURL(baseURL),
	}
}

func (s *LinkService) GetByID(ctx context.Context, id int64) (model.Link, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *LinkService) GetByShortName(ctx context.Context, shortName string) (model.Link, error) {
	return s.repository.GetByShortName(ctx, shortName)
}

func (s *LinkService) List(ctx context.Context, from int, to int) ([]model.Link, int64, error) {
	if from < 0 || to < 0 || from > to {
		return nil, 0, apperrors.ErrNotValidQuery
	}

	links, err := s.repository.List(ctx, from, to)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repository.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return links, total, nil
}

func (s *LinkService) Create(ctx context.Context, link model.Link) (model.Link, error) {
	if link.ShortName != "" {
		link.ShortUrl = s.shortURL(link.ShortName)
		return s.repository.Create(ctx, link)
	}

	return s.createWithGeneratedShortName(ctx, link)
}

func (s *LinkService) createWithGeneratedShortName(ctx context.Context, link model.Link) (model.Link, error) {
	const attempts = 5

	for range attempts {
		shortName, err := generateShortName()
		if err != nil {
			return model.Link{}, err
		}

		link.ShortName = shortName
		link.ShortUrl = s.shortURL(shortName)

		createdLink, err := s.repository.Create(ctx, link)
		if err == nil {
			return createdLink, nil
		}
		if !stderrors.Is(err, apperrors.ErrShortNameTaken) {
			return model.Link{}, err
		}
	}

	return model.Link{}, apperrors.ErrShortNameTaken
}

func (s *LinkService) Update(ctx context.Context, link model.Link) (model.Link, error) {
	if link.ShortName == "" {
		shortName, err := generateShortName()
		if err != nil {
			return model.Link{}, err
		}
		link.ShortName = shortName
	}

	link.ShortUrl = s.shortURL(link.ShortName)

	return s.repository.Update(ctx, link)
}

func (s *LinkService) Delete(ctx context.Context, id int64) error {
	return s.repository.Delete(ctx, id)
}

func (s *LinkService) shortURL(shortName string) string {
	return s.baseURL + "/" + shortName
}

func redirectBaseURL(baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(baseURL, "/r") {
		return baseURL
	}

	return baseURL + "/r"
}

func generateShortName() (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const size = 8

	result := make([]byte, size)
	for i := range result {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", fmt.Errorf("generate short name: %w", err)
		}
		result[i] = alphabet[index.Int64()]
	}

	return string(result), nil
}
