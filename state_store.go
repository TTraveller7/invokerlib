package invokerlib

import "context"

type StateStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, val []byte) error
	Delete(ctx context.Context, key string) error
}
