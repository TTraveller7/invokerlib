package invokerlib

import (
	"context"

	"github.com/coocood/freecache"
)

type FreeCacheStateStore struct {
	StateStore
	cli *freecache.Cache
}

func NewFreeCacheStateStore() (StateStore, error) {
	c := freecache.NewCache(CacheSize)
	s := &FreeCacheStateStore{
		cli: c,
	}
	return s, nil
}

func (f *FreeCacheStateStore) Get(ctx context.Context, key string) ([]byte, error) {
	return f.cli.Get([]byte(key))
}

func (f *FreeCacheStateStore) Put(ctx context.Context, key string, val []byte) error {
	return f.cli.Set([]byte(key), val, 0)
}

func (f *FreeCacheStateStore) PutWithExpireTime(ctx context.Context, key string, val []byte, expireSeconds int) error {
	return f.cli.Set([]byte(key), val, expireSeconds)
}

func (f *FreeCacheStateStore) Delete(ctx context.Context, key string) error {
	f.cli.Del([]byte(key))
	return nil
}

func (f *FreeCacheStateStore) Keys(ctx context.Context, limit int) ([]string, error) {
	iter := f.cli.NewIterator()
	keys := make([]string, 0)

	l := limit
	if l == 0 {
		l = int(f.cli.EntryCount())
	}
	count := 0
	for entry := iter.Next(); entry != nil; {
		if count == l {
			break
		}
		keys = append(keys, string(entry.Key))
		count++
	}
	return keys, nil
}
