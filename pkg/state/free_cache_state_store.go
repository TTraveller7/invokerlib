package state

import (
	"context"

	"github.com/TTraveller7/invokerlib/pkg/consts"
	"github.com/coocood/freecache"
)

type FreeCacheStateStore struct {
	StateStore
	cli *freecache.Cache
}

func NewFreeCacheStateStore() (StateStore, error) {
	c := freecache.NewCache(consts.CacheSize)
	s := &FreeCacheStateStore{
		cli: c,
	}
	return s, nil
}

func (f *FreeCacheStateStore) Get(ctx context.Context, key string) ([]byte, error) {
	res, err := f.cli.Get([]byte(key))
	if err == freecache.ErrNotFound {
		return nil, consts.ErrStateStoreKeyNotExist
	} else if err != nil {
		return nil, err
	} else {
		return res, nil
	}
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
	for i := 0; i < l; i++ {
		entry := iter.Next()
		if entry == nil {
			break
		}
		keys = append(keys, string(entry.Key))
	}
	return keys, nil
}
