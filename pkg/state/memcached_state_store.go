package state

import (
	"context"
	"fmt"

	"github.com/TTraveller7/invokerlib/pkg/conf"
	"github.com/bradfitz/gomemcache/memcache"
)

type MemcachedStateStore struct {
	StateStore
	cli *memcache.Client
}

func NewMemcachedStateStore(name string) (StateStore, error) {
	mc := conf.GetMemcachedConfigByName(name)
	cli := memcache.New(mc.Addresses...)
	if err := cli.Ping(); err != nil {
		return nil, err
	} else {
		return &MemcachedStateStore{
			cli: cli,
		}, nil
	}
}

func (m *MemcachedStateStore) Get(ctx context.Context, key string) ([]byte, error) {
	item, err := m.cli.Get(key)
	if err != nil {
		return nil, fmt.Errorf("memcached state store Get failed: %v", err)
	} else {
		return item.Value, nil
	}
}

func (m *MemcachedStateStore) Put(ctx context.Context, key string, val []byte) error {
	return m.PutWithExpireTime(ctx, key, val, 0)
}

func (m *MemcachedStateStore) PutWithExpireTime(ctx context.Context, key string, val []byte, expireSeconds int) error {
	item := &memcache.Item{
		Key:        key,
		Value:      val,
		Expiration: int32(expireSeconds),
	}
	if err := m.cli.Set(item); err != nil {
		return err
	} else {
		return nil
	}
}

func (m *MemcachedStateStore) Delete(ctx context.Context, key string) error {
	if err := m.cli.Delete(key); err != nil && err != memcache.ErrCacheMiss {
		return err
	} else {
		return nil
	}
}

func (m *MemcachedStateStore) Keys(ctx context.Context, limit int) ([]string, error) {
	return nil, ErrNotImplemented
}
