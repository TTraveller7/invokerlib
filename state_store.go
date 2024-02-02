package invokerlib

import (
	"context"
)

var GlobalStateStoreTypes = struct {
	Redis string
}{
	Redis: "Redis",
}

type StateStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, val []byte) error
	PutWithExpireTime(ctx context.Context, key string, val []byte, expireSeconds int) error
	Delete(ctx context.Context, key string) error
	Keys(ctx context.Context) ([]string, error)
}

var stateStores map[string]StateStore = make(map[string]StateStore, 0)

func AddStateStore(name string, stateStore StateStore) {
	stateStores[name] = stateStore
}

func cat(ctx context.Context) (ProcessorCatResult, error) {
	resp := make(map[string][]StateStoreEntry, 0)
	for name, stateStore := range stateStores {
		keys, err := stateStore.Keys(ctx)
		if err != nil {
			return nil, err
		}
		entries := make([]StateStoreEntry, 0)
		for _, key := range keys {
			val, err := stateStore.Get(ctx, key)
			if err != nil {
				return nil, err
			}
			entries = append(entries, StateStoreEntry{
				Key: key,
				Val: string(val),
			})
		}
		resp[name] = entries
	}
	return resp, nil
}
