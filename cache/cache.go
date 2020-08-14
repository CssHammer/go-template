package cache

import (
	"context"
	"time"
)

type Cache interface {
	HealthCheck() error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, val string, ttl time.Duration) error
}
