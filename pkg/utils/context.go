package utils

import (
	"context"

	"github.com/TTraveller7/invokerlib/pkg/consts"
)

func loadValue(ctx context.Context, key any) (string, bool) {
	val := ctx.Value(consts.CTX_KEY_INVOKER_LIB_WORKER_ID)
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
	return loadValue(ctx, consts.CTX_KEY_INVOKER_LIB_WORKER_ID)
}

func WorkerIndex(ctx context.Context) (string, bool) {
	return loadValue(ctx, consts.CTX_KEY_INVOKER_LIB_WORKER_INDEX)
}

func NewWorkerContext(processorCtx context.Context, workerIndex int, processorName string,
	topicName string) context.Context {
	if _, exists := WorkerId(processorCtx); exists {
		return processorCtx
	}
	workerId := NewWorkerId(workerIndex, processorName, topicName)
	ctx := context.WithValue(processorCtx, consts.CTX_KEY_INVOKER_LIB_WORKER_ID, workerId)
	ctx = context.WithValue(ctx, consts.CTX_KEY_INVOKER_LIB_WORKER_INDEX, workerIndex)
	ctx = context.WithValue(ctx, consts.CTX_KEY_INVOKER_LIB_WORKER_TOPIC, topicName)
	return ctx
}
