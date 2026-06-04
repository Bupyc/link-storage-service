package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/Bupyc/link-storage-service/internal/cache"
	"github.com/Bupyc/link-storage-service/internal/model"
	"github.com/Bupyc/link-storage-service/internal/repository"
)

type LinkService struct {
	repo  repository.LinkRepository
	cache cache.LinkCache
}

func NewLinkService(repo repository.LinkRepository, cache cache.LinkCache) *LinkService {
	return &LinkService{
		repo:  repo,
		cache: cache,
	}
}

const maxCreateAttempts = 5

var ErrCreateLinkFailed = errors.New("failed to create unique short code")

func generateShortCode() (string, error) {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func isValidURL(rawURL string) bool {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}

	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

func (s *LinkService) Create(ctx context.Context, rawURL string) (model.Link, error) {
	rawURL = strings.TrimSpace(rawURL)
	if !isValidURL(rawURL) {
		return model.Link{}, ErrInvalidURL
	}

	for range maxCreateAttempts {
		shortCode, err := generateShortCode()
		if err != nil {
			return model.Link{}, err
		}

		link := model.Link{
			ID:          shortCode,
			OriginalURL: rawURL,
			CreatedAt:   time.Now().UTC(),
			Visits:      0,
		}

		if err := s.repo.Create(ctx, link); err != nil {
			if errors.Is(err, repository.ErrLinkAlreadyExists) {
				continue
			}
			return model.Link{}, err
		}
		return link, nil
	}

	return model.Link{}, ErrCreateLinkFailed
}

var ErrInvalidURL = errors.New("invalid url")

func (s *LinkService) Get(ctx context.Context, shortCode string) (model.Link, error) {
	shortCode = strings.TrimSpace(shortCode)

	var link model.Link

	cachedLink, ok := s.cache.Get(shortCode)
	if ok {
		link = cachedLink
	} else {
		repoLink, err := s.repo.GetByShortCode(ctx, shortCode)
		if err != nil {
			return model.Link{}, err
		}

		link = repoLink
	}

	visits, err := s.repo.IncrementVisits(ctx, shortCode)
	if err != nil {
		return model.Link{}, err
	}

	link.Visits = visits
	s.cache.Set(link)

	return link, nil
}

func (s *LinkService) List(ctx context.Context, limit, offset int) ([]model.Link, error) {
	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, limit, offset)
}

func (s *LinkService) Delete(ctx context.Context, shortCode string) error {
	shortCode = strings.TrimSpace(shortCode)

	if _, err := s.repo.GetByShortCode(ctx, shortCode); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, shortCode); err != nil {
		return err
	}

	s.cache.Delete(shortCode)

	return nil
}

func (s *LinkService) Stats(ctx context.Context, shortCode string) (model.Link, error) {
	shortCode = strings.TrimSpace(shortCode)

	return s.repo.GetByShortCode(ctx, shortCode)
}

func (s *LinkService) Count(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}
