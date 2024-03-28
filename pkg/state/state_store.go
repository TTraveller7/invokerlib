package state

import (
	"context"
)

type StateStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, val []byte) error
	PutWithExpireTime(ctx context.Context, key string, val []byte, expireSeconds int) error
	Delete(ctx context.Context, key string) error
	Keys(ctx context.Context, limit int) ([]string, error)
}
