package invokerlib

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisStateStore struct {
	StateStore
	conf *RedisStateStoreConfig
	cli  *redis.Client
}

func NewRedisStateStore() (StateStore, error) {
	if conf == nil {
		return nil, ErrConfigNotInitialized
	}
	if conf.RedisConf == nil {
		return nil, ErrRedisConfNotInitialized
	}

	redisConf := conf.RedisConf
	cli := redis.NewClient(&redis.Options{
		Addr:     redisConf.Address,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	if err := cli.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &RedisStateStore{
		conf: redisConf,
		cli:  cli,
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
