package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/goquizvibe/metrics"
	"github.com/prometheus/client_golang/prometheus"
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

func GetOrFetch[T any](ctx context.Context, cache *CacheService, key string, fetchFunc func() (T, error), ttl ...time.Duration) (T, error) {
	if cache == nil {
		return fetchFunc()
	}

	queryName := normalizeCacheKey(key)

	var zero T
	var result T
	if cache.Get(ctx, key, &result, queryName) {
		return result, nil
	}

	metrics.CacheMissesTotal.With(prometheus.Labels{"query": queryName}).Inc()

	fetched, err := fetchFunc()
	if err != nil {
		return zero, err
	}

	useTTL := cache.defaultTTL
	if len(ttl) > 0 {
		useTTL = ttl[0]
	}

	go func() {
		if cacheErr := cache.Set(ctx, key, fetched, useTTL, queryName); cacheErr != nil {
			log.Printf("cache set error for key %s: %v", key, cacheErr)
		}
	}()

	return fetched, nil
}

func SaveOrUpdate[T any](ctx context.Context, cache *CacheService, key string, saveFunc func() (T, error), ttl ...time.Duration) (T, error) {
	if cache == nil {
		return saveFunc()
	}

	queryName := normalizeCacheKey(key)

	var zero T
	result, err := saveFunc()
	if err != nil {
		return zero, err
	}

	useTTL := cache.defaultTTL
	if len(ttl) > 0 {
		useTTL = ttl[0]
	}

	if err := cache.Set(ctx, key, result, useTTL, queryName); err != nil {
		log.Printf("cache set error for key %s: %v", key, err)
	}

	return result, nil
}

func Delete(ctx context.Context, cache *CacheService, key string, deleteFunc func() error) error {
	if err := deleteFunc(); err != nil {
		return err
	}

	if cache == nil {
		return nil
	}

	if err := cache.Delete(ctx, key, normalizeCacheKey(key)); err != nil {
		log.Printf("cache delete error for key %s: %v", key, err)
	}

	return nil
}

func InvalidateCache(ctx context.Context, cache *CacheService, keys ...string) {
	if cache == nil {
		return
	}
	for _, key := range keys {
		if err := cache.Delete(ctx, key, ""); err != nil {
			log.Printf("cache delete error for key %s: %v", key, err)
		}
	}
}

func (s *CacheService) Get(ctx context.Context, key string, dest any, queryName string) bool {
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		metrics.RedisOperationsTotal.With(prometheus.Labels{
			"operation": "get",
			"status":    "error",
			"query":     queryName,
		}).Inc()
		return false
	}
	if err := json.Unmarshal(data, dest); err != nil {
		metrics.RedisOperationsTotal.With(prometheus.Labels{
			"operation": "get",
			"status":    "error",
			"query":     queryName,
		}).Inc()
		return false
	}
	metrics.RedisOperationsTotal.With(prometheus.Labels{
		"operation": "get",
		"status":    "success",
		"query":     queryName,
	}).Inc()
	metrics.CacheHitsTotal.With(prometheus.Labels{"query": queryName}).Inc()
	return true
}

func (s *CacheService) Set(ctx context.Context, key string, value any, ttl time.Duration, queryName string) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = s.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		metrics.RedisOperationsTotal.With(prometheus.Labels{
			"operation": "set",
			"status":    "error",
			"query":     queryName,
		}).Inc()
		return err
	}
	metrics.RedisOperationsTotal.With(prometheus.Labels{
		"operation": "set",
		"status":    "success",
		"query":     queryName,
	}).Inc()
	return nil
}

func (s *CacheService) SetDefault(ctx context.Context, key string, value any, queryName string) error {
	return s.Set(ctx, key, value, s.defaultTTL, queryName)
}

func (s *CacheService) Delete(ctx context.Context, key string, queryName string) error {
	err := s.client.Del(ctx, key).Err()
	if err != nil {
		metrics.RedisOperationsTotal.With(prometheus.Labels{
			"operation": "del",
			"status":    "error",
			"query":     queryName,
		}).Inc()
		return err
	}
	metrics.RedisOperationsTotal.With(prometheus.Labels{
		"operation": "del",
		"status":    "success",
		"query":     queryName,
	}).Inc()
	return nil
}

func (s *CacheService) Exists(ctx context.Context, key string, queryName string) (bool, error) {
	n, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		metrics.RedisOperationsTotal.With(prometheus.Labels{
			"operation": "exists",
			"status":    "error",
			"query":     queryName,
		}).Inc()
		return false, err
	}
	metrics.RedisOperationsTotal.With(prometheus.Labels{
		"operation": "exists",
		"status":    "success",
		"query":     queryName,
	}).Inc()
	return n > 0, nil
}

func Get[T any](ctx context.Context, s *CacheService, key string) (T, bool) {
	var result T
	if s.Get(ctx, key, &result, "") {
		return result, true
	}
	var zero T
	return zero, false
}

func Set[T any](ctx context.Context, s *CacheService, key string, value T, ttl time.Duration) error {
	return s.Set(ctx, key, value, ttl, "")
}
