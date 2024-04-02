package state

import (
	"context"
	"time"

	"github.com/TTraveller7/invokerlib/pkg/consts"
	"github.com/TTraveller7/invokerlib/pkg/utils"
)

type StateStoreWrapper struct {
	StateStore
	s             StateStore
	metricsClient *utils.MetricsClient
}

func NewStateStoreWrapper(s StateStore, metricsClient *utils.MetricsClient) *StateStoreWrapper {
	return &StateStoreWrapper{
		s:             s,
		metricsClient: metricsClient,
	}
}

func (w *StateStoreWrapper) Get(ctx context.Context, key string) ([]byte, error) {
	startTime := time.Now()

	res, err := w.s.Get(ctx, key)

	elapsedTime := time.Since(startTime)
	w.metricsClient.EmitHistogram("get_latency", "Latency of get operation in milliseconds",
		float64(elapsedTime.Milliseconds()))
	if err != nil {
		if err == consts.ErrStateStoreKeyNotExist {
			w.metricsClient.EmitCounter("get_cache_miss", "Number of cache misses", 1)
		} else {
			w.metricsClient.EmitCounter("get_failure", "Number of get failures", 1)
		}
	}

	return res, err
}

func (w *StateStoreWrapper) Put(ctx context.Context, key string, val []byte) error {
	return w.PutWithExpireTime(ctx, key, val, 0)
}

func (w *StateStoreWrapper) PutWithExpireTime(ctx context.Context, key string, val []byte, expireSeconds int) error {
	startTime := time.Now()

	err := w.s.PutWithExpireTime(ctx, key, val, expireSeconds)

	elapsedTime := time.Since(startTime)
	w.metricsClient.EmitHistogram("put_latency", "Latency of put operation in milliseconds",
		float64(elapsedTime.Milliseconds()))
	if err != nil {
		w.metricsClient.EmitCounter("put_failure", "Number of put failures", 1)
	}

	return err
}

func (w *StateStoreWrapper) Delete(ctx context.Context, key string) error {
	startTime := time.Now()

	err := w.s.Delete(ctx, key)

	elapsedTime := time.Since(startTime)
	w.metricsClient.EmitHistogram("delete_latency", "Latency of delete operation in milliseconds",
		float64(elapsedTime.Milliseconds()))
	if err != nil {
		w.metricsClient.EmitCounter("delete_failure", "Number of delete failures", 1)
	}

	return err
}

func (w *StateStoreWrapper) Keys(ctx context.Context, limit int) ([]string, error) {
	return w.s.Keys(ctx, limit)
}
