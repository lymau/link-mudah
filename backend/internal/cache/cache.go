package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const DefaultTTL = 1 * time.Hour

type Cache struct {
	Client *redis.Client
}

func New(client *redis.Client) *Cache {
	return &Cache{Client: client}
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return Get(ctx, c.Client, key)
}

func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return Set(ctx, c.Client, key, value, ttl)
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return Delete(ctx, c.Client, key)
}

func UserPageKey(username string) string {
	return "page:" + username
}

func Get(ctx context.Context, client *redis.Client, key string) (string, error) {
	if client == nil {
		return "", nil
	}

	value, err := client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	return value, err
}

func Set(ctx context.Context, client *redis.Client, key string, value any, ttl time.Duration) error {
	if client == nil {
		return nil
	}

	payload, err := marshalValue(value)
	if err != nil {
		return err
	}
	return client.Set(ctx, key, payload, ttl).Err()
}

func Delete(ctx context.Context, client *redis.Client, key string) error {
	if client == nil {
		return nil
	}
	return client.Del(ctx, key).Err()
}

func marshalValue(value any) ([]byte, error) {
	switch v := value.(type) {
	case nil:
		return []byte("null"), nil
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return json.Marshal(v)
	}
}
