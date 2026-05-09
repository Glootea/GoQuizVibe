package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	client     *redis.Client
	defaultTTL time.Duration
}

func NewCacheService(client *redis.Client, ttl time.Duration) *CacheService {
	return &CacheService{
		client:     client,
		defaultTTL: ttl,
	}
}

func (s *CacheService) Get(ctx context.Context, key string, dest any) bool {
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		return false
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return false
	}
	return true
}

func (s *CacheService) Set(ctx context.Context, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, data, s.defaultTTL).Err()
}

func (s *CacheService) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

func (s *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	n, err := s.client.Exists(ctx, key).Result()
	return n > 0, err
}

func Get[T any](ctx context.Context, s *CacheService, key string) (T, bool) {
	var result T
	if s.Get(ctx, key, &result) {
		return result, true
	}
	var zero T
	return zero, false
}

func Set[T any](ctx context.Context, s *CacheService, key string, value T) error {
	return s.Set(ctx, key, value)
}
