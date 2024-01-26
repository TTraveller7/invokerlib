package invokerlib

import (
	"context"
	"fmt"

	"github.com/allegro/bigcache/v3"
)

type BigCacheStateStore struct {
	StateStore
	cli *bigcache.BigCache
}

func NewBigCacheStateStore() (StateStore, error) {
	config := bigcache.Config{
		// number of shards (must be a power of 2)
		Shards: 1024,

		// time after which entry can be evicted
		LifeWindow: 0,

		// Interval between removing expired entries (clean up).
		// If set to <= 0 then no action is performed.
		// Setting to < 1 second is counterproductive — bigcache has a one second resolution.
		CleanWindow: 0,

		// cache will not allocate more memory than this limit, value in MB
		// if value is reached then the oldest entries can be overridden for the new ones
		// 0 value means no size limit
		HardMaxCacheSize: 512,

		// rps * lifeWindow, used only in initial memory allocation
		MaxEntriesInWindow: 1000 * 10 * 60,
		// max entry size in bytes, used only in initial memory allocation
		MaxEntrySize: 500,

		// prints information about additional memory allocation
		Verbose: true,
		Logger:  logs,
	}

	b, err := bigcache.New(processorCtx, config)
	if err != nil {
		return nil, err
	}
	s := &BigCacheStateStore{
		cli: b,
	}
	return s, nil
}

func (b *BigCacheStateStore) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := b.cli.Get(key)
	if err != nil {
		return nil, fmt.Errorf("big cache state store Get failed: %v", err)
	}
	return val, nil
}

func (b *BigCacheStateStore) Put(ctx context.Context, key string, val []byte) error {
	if err := b.cli.Set(key, val); err != nil {
		return err
	}
	return nil
}

func (b *BigCacheStateStore) Delete(ctx context.Context, key string) error {
	if err := b.cli.Delete(key); err != nil {
		return err
	}
	return nil
}

func (b *BigCacheStateStore) Keys(ctx context.Context) ([]string, error) {
	keys := make([]string, 0)
	if b.cli.Len() == 0 {
		return keys, nil
	}

	iter := b.cli.Iterator()
	for {
		entry, err := iter.Value()
		if err != nil {
			return nil, err
		}
		keys = append(keys, entry.Key())
		if !iter.SetNext() {
			break
		}
	}
	return keys, nil
}