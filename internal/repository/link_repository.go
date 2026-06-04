package repository

import (
	"context"

	"github.com/Bupyc/link-storage-service/internal/model"
)

type LinkRepository interface {
	Create(ctx context.Context, link model.Link) error
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
	GetByShortCode(ctx context.Context, shortCode string) (model.Link, error)
	List(ctx context.Context, limit, offset int) ([]model.Link, error)
	Delete(ctx context.Context, shortCode string) error
	IncrementVisits(ctx context.Context, shortCode string) (int64, error)
	Count(ctx context.Context) (int64, error)
}
