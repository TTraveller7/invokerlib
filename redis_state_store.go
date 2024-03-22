package invokerlib

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStateStore struct {
	StateStore
	cli *redis.Client
}

var redisConfigs map[string]*RedisConfig = make(map[string]*RedisConfig, 0)

func NewRedisStateStore(name string) (StateStore, error) {
	rc, exists := redisConfigs[name]
	if !exists {
		return nil, fmt.Errorf("redis state store config with name %s does not exist", name)
	}
	cli := redis.NewClient(&redis.Options{
		Addr: rc.Address,
	})
	pingSuccess := false
	for i := 0; i < RedisPingRetryTimes; i++ {
		if err := cli.Ping(context.Background()).Err(); err == nil {
			pingSuccess = true
			break
		}
		time.Sleep(1 * time.Second)
	}
	if !pingSuccess {
		return nil, fmt.Errorf("redis state store %s cannot connect to address %s", rc.Name, rc.Address)
	}
	return &RedisStateStore{
		cli: cli,
	}, nil
}

func (r *RedisStateStore) Get(ctx context.Context, key string) ([]byte, error) {
	strVal, err := r.cli.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis state store Get failed: %v", err)
	}
	return []byte(strVal), nil
}

func (r *RedisStateStore) Put(ctx context.Context, key string, val []byte) error {
	return r.PutWithExpireTime(ctx, key, val, 0)
}

func (r *RedisStateStore) PutWithExpireTime(ctx context.Context, key string, val []byte, expireSeconds int) error {
	if err := r.cli.Set(ctx, key, string(val), time.Duration(expireSeconds)*time.Second).Err(); err != nil {
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

func (r *RedisStateStore) Keys(ctx context.Context, limit int) ([]string, error) {
	keys, err := r.cli.Keys(ctx, "*").Result()
	if err != nil {
		return nil, err
	}
	return keys, nil
}
