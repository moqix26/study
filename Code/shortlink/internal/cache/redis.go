package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type LinkCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewLinkCache(rdb *redis.Client, ttl time.Duration) *LinkCache {
	return &LinkCache{rdb: rdb, ttl: ttl}
}

func key(code string) string {
	return "link:" + code
}

func (c *LinkCache) Get(ctx context.Context, code string) (string, bool, error) {
	val, err := c.rdb.Get(ctx, key(code)).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if val == "" {
		return "", false, nil
	}
	return val, true, nil
}

func (c *LinkCache) Set(ctx context.Context, code, longURL string) error {
	return c.rdb.Set(ctx, key(code), longURL, c.ttl).Err()
}

func (c *LinkCache) Del(ctx context.Context, code string) error {
	return c.rdb.Del(ctx, key(code)).Err()
}
