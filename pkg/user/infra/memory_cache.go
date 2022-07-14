package infra

import (
	"context"
	"time"

	"github.com/pmylund/go-cache"
	"gitlab.haochang.tv/yangyi/examine-code/pkg/user/domain"
)

// MemoryCache 本地内存缓存
type MemoryCache struct {
	values *cache.Cache
}

// NewMemoryCache 内存缓存
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		values: cache.New(5*time.Minute, 10*time.Minute),
	}
}

// Get 获取缓存结果
func (mc *MemoryCache) Get(_ context.Context, key string) ([]byte, error) {
	if v, ok := mc.values.Get(key); ok {
		return v.([]byte), nil
	}
	return nil, domain.ErrMissingCache
}

// Put 保存缓存
func (mc *MemoryCache) Put(_ context.Context, key string, value []byte, expiration time.Duration) error {
	mc.values.Set(key, value, expiration)
	return nil
}

// Delete 删除缓存
func (mc *MemoryCache) Delete(_ context.Context, key string) error {
	mc.values.Delete(key)
	return nil
}
