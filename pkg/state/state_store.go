package state

import (
	"context"
	"fmt"
)

var ErrNotImplemented error = fmt.Errorf("not Implemented")

type StateStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, val []byte) error
	PutWithExpireTime(ctx context.Context, key string, val []byte, expireSeconds int) error
	Delete(ctx context.Context, key string) error
	Keys(ctx context.Context, limit int) ([]string, error)
}

var stateStores map[string]StateStore = make(map[string]StateStore, 0)

func AddStateStore(name string, stateStore StateStore) {
	stateStores[name] = stateStore
}

func StateStores() map[string]StateStore {
	return stateStores
}
