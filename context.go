package invokerlib

import (
	"context"
	"fmt"
)

func prependWorkerId(ctx context.Context, s string) string {
	return fmt.Sprintf("%v_%s", ctx.Value(CTX_KEY_INVOKER_LIB_WORKER_ID), s)
}

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

func NewWorkerContext(functionCtx context.Context, workerIndex int) context.Context {
	if _, exists := WorkerId(functionCtx); exists {
		return functionCtx
	}
	workerId := NewWorkerId(workerIndex)
	ctx := context.WithValue(functionCtx, CTX_KEY_INVOKER_LIB_WORKER_ID, workerId)
	ctx = context.WithValue(ctx, CTX_KEY_INVOKER_LIB_WORKER_INDEX, workerIndex)
	return ctx
}
