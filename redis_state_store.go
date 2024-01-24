package invokerlib

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisStateStore struct {
	StateStore
	cli *redis.Client
}

func NewRedisStateStore() (StateStore, error) {
	// TODO
	return nil, nil
}

func (r *RedisStateStore) Get(ctx context.Context, key string) ([]byte, error) {
	strVal, err := r.cli.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis state store Get failed: %v", err)
	}
	return []byte(strVal), nil
}

func (r *RedisStateStore) Put(ctx context.Context, key string, val []byte) error {
	if err := r.cli.Set(ctx, key, string(val), 0).Err(); err != nil {
		return err
	}
	return nil
}

func (r *RedisStateStore) Delete(ctx context.Context, key string) error {
	if err := r.cli.Del(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}

func (r *RedisStateStore) Keys(ctx context.Context) ([]string, error) {
	keys, err := r.cli.Keys(ctx, "*").Result()
	if err != nil {
		return nil, err
	}
	return keys, nil
}
