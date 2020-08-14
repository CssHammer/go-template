package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const ServiceName = "redis cache"

// Cache describes connection to Cache server
type Cache struct {
	client *redis.Client
	log    *zap.Logger
}

// New returns the initialized Cache object
func New(ctx context.Context, wg *sync.WaitGroup, log *zap.Logger, redisServer string) (*Cache, error) {
	log = log.Named(ServiceName)
	log.Info("new", zap.String("host", redisServer))

	client := redis.NewClient(&redis.Options{
		Addr: redisServer,
	})

	c := new(Cache)
	c.client = client
	c.log = log

	if err := c.HealthCheck(); err != nil {
		return nil, fmt.Errorf("healthcheck: %w", err)
	}

	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ctx.Done()
		err := c.client.Close()
		if err != nil {
			c.log.Error("close", zap.Error(err))
			return
		}
		c.log.Info("close")
	}()

	return c, nil
}

// HealthCheck checks if connection exists
func (c *Cache) HealthCheck() error {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	_, err := c.client.Ping(ctx).Result()
	return fmt.Errorf("ping: %w", err)
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	res, err := c.client.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return "", fmt.Errorf("get: %w", err)
	}

	return res, nil
}

func (c *Cache) Set(ctx context.Context, key, val string, ttl time.Duration) error {
	err := c.client.Set(ctx, key, val, ttl).Err()
	if err != nil {
		return fmt.Errorf("set: %w", err)
	}

	return nil
}
