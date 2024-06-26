package state

import (
	"context"
	"fmt"
	"time"

	"github.com/TTraveller7/invokerlib/pkg/conf"
	"github.com/TTraveller7/invokerlib/pkg/consts"
	"github.com/redis/go-redis/v9"
)

type RedisStateStore struct {
	StateStore
	cli *redis.Client
}

func NewRedisStateStore(name string) (StateStore, error) {
	rc := conf.GetRedisConfigByName(name)
	if rc == nil {
		return nil, fmt.Errorf("redis config with name %s not found", name)
	}
	cli := redis.NewClient(&redis.Options{
		Addr: rc.Address,
	})
	pingSuccess := false
	for i := 0; i < consts.RedisPingRetryTimes; i++ {
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
	if err == redis.Nil {
		return nil, consts.ErrStateStoreKeyNotExist
	} else if err != nil {
		return nil, fmt.Errorf("redis state store Get failed: %v", err)
	} else {
		return []byte(strVal), nil
	}
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
	if len(keys) > consts.DefaultCatLimit {
		keys = keys[:consts.DefaultCatLimit]
	}
	return keys, nil
}
