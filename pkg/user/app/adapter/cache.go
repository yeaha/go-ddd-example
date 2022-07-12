package adapter

import (
	"context"
	"time"
)

// Cacher 缓存接口
type Cacher interface {
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Get(ctx context.Context, key string) (value []byte, err error)
}
