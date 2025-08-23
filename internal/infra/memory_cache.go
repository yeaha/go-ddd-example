package infra

import (
	"context"
	"time"

	"ddd-example/internal/app/adapter"
	"ddd-example/internal/domain"

	"github.com/pmylund/go-cache"
)

// memoryCache 本地内存缓存
type memoryCache struct {
	values *cache.Cache
}

// NewMemoryCache 内存缓存
func NewMemoryCache() adapter.Cacher {
	return &memoryCache{
		values: cache.New(5*time.Minute, 10*time.Minute),
	}
}

// Get 获取缓存结果
func (mc *memoryCache) Get(_ context.Context, key string) ([]byte, error) {
	if v, ok := mc.values.Get(key); ok {
		return v.([]byte), nil
	}
	return nil, domain.ErrMissingCache
}

// Put 保存缓存
func (mc *memoryCache) Put(_ context.Context, key string, value []byte, expiration time.Duration) error {
	mc.values.Set(key, value, expiration)
	return nil
}

// Delete 删除缓存
func (mc *memoryCache) Delete(_ context.Context, key string) error {
	mc.values.Delete(key)
	return nil
}
