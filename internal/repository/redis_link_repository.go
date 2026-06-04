package repository

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Bupyc/link-storage-service/internal/model"
)

const linksIndexKey = "links:index"

var (
	ErrLinkAlreadyExists = errors.New("link already exists")
	ErrLinkNotFound      = errors.New("link not found")
)

type RedisLinkRepository struct {
	client *redis.Client
}

func NewRedisLinkRepository(client *redis.Client) *RedisLinkRepository {
	return &RedisLinkRepository{
		client: client,
	}
}

func linkKey(id string) string {
	return "link:" + id
}

func (r *RedisLinkRepository) Create(ctx context.Context, link model.Link) error {
	exists, err := r.client.Exists(ctx, linkKey(link.ID)).Result()
	if err != nil {
		return err
	}

	if exists > 0 {
		return ErrLinkAlreadyExists
	}

	pipe := r.client.TxPipeline()

	pipe.HSet(ctx, linkKey(link.ID), map[string]any{
		"id":           link.ID,
		"original_url": link.OriginalURL,
		"created_at":   link.CreatedAt.Format(time.RFC3339),
		"visits":       link.Visits,
	})
	pipe.RPush(ctx, linksIndexKey, link.ID)

	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisLinkRepository) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	return r.client.HGet(ctx, linkKey(shortCode), "original_url").Result()
}

func (r *RedisLinkRepository) GetByShortCode(ctx context.Context, shortCode string) (model.Link, error) {
	values, err := r.client.HGetAll(ctx, linkKey(shortCode)).Result()
	if err != nil {
		return model.Link{}, err
	}

	if len(values) == 0 {
		return model.Link{}, ErrLinkNotFound
	}

	createdAt, err := time.Parse(time.RFC3339, values["created_at"])
	if err != nil {
		return model.Link{}, err
	}

	visits, err := strconv.ParseInt(values["visits"], 10, 64)
	if err != nil {
		return model.Link{}, err
	}

	return model.Link{
		ID:          values["id"],
		OriginalURL: values["original_url"],
		CreatedAt:   createdAt,
		Visits:      visits,
	}, nil
}

func (r *RedisLinkRepository) List(ctx context.Context, limit, offset int) ([]model.Link, error) {
	ids, err := r.client.LRange(ctx, linksIndexKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, err
	}

	links := make([]model.Link, 0, len(ids))

	for _, id := range ids {
		link, err := r.GetByShortCode(ctx, id)
		if err != nil {
			return nil, err
		}

		links = append(links, link)
	}

	return links, nil
}

func (r *RedisLinkRepository) Delete(ctx context.Context, shortCode string) error {
	pipe := r.client.TxPipeline()

	pipe.Del(ctx, linkKey(shortCode))
	pipe.LRem(ctx, linksIndexKey, 0, shortCode)

	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisLinkRepository) IncrementVisits(ctx context.Context, shortCode string) (int64, error) {
	return r.client.HIncrBy(ctx, linkKey(shortCode), "visits", 1).Result()
}

func (r *RedisLinkRepository) Count(ctx context.Context) (int64, error) {
	return r.client.LLen(ctx, linksIndexKey).Result()
}
