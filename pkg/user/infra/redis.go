package infra

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache redis缓存
type RedisCache struct {
	client redis.Cmdable
}

// NewRedisCache 构造函数
func NewRedisCache(client redis.Cmdable) *RedisCache {
	return &RedisCache{client: client}
}

// Set 写缓存
func (cache *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return cache.client.Set(ctx, key, value, ttl).Err()
}

// Get 读取缓存
func (cache *RedisCache) Get(ctx context.Context, key string) (value []byte, err error) {
	return cache.client.Get(ctx, key).Bytes()
}
