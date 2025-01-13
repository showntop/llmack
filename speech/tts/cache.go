package tts

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
)

var onceCache sync.Once

// redisCache is the struct for the redis
type redisCache struct {
	client *redis.Client
}

var defaultCache *redisCache

// Cache creates a new redis cache
func Cache(cli *redis.Client) *redisCache {
	onceCache.Do(func() {
		cli := redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		})
		defaultCache = &redisCache{client: cli}
	})
	return defaultCache
}

func (c *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	return c.client.Get(ctx, key).Bytes()
}

func (c *redisCache) Set(ctx context.Context, key string, raw []byte) error {
	return c.client.Set(ctx, key, raw, 0).Err()
}
