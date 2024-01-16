package invokerlib

import (
	"context"
)

func loadValue(ctx context.Context, key any) (string, bool) {
	val := ctx.Value(CTX_KEY_INVOKER_LIB_WORKER_ID)
	if val == nil {
		return "", false
	}
	if valStr, ok := val.(string); !ok {
		return "", false
	} else {
		return valStr, true
	}
}

func WorkerId(ctx context.Context) (string, bool) {
	return loadValue(ctx, CTX_KEY_INVOKER_LIB_WORKER_ID)
}

func WorkerIndex(ctx context.Context) (string, bool) {
	return loadValue(ctx, CTX_KEY_INVOKER_LIB_WORKER_INDEX)
}

func NewWorkerContext(processorCtx context.Context, workerIndex int) context.Context {
	if _, exists := WorkerId(processorCtx); exists {
		return processorCtx
	}
	workerId := NewWorkerId(workerIndex)
	ctx := context.WithValue(processorCtx, CTX_KEY_INVOKER_LIB_WORKER_ID, workerId)
	ctx = context.WithValue(ctx, CTX_KEY_INVOKER_LIB_WORKER_INDEX, workerIndex)
	return ctx
}
