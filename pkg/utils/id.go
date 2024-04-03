package utils

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
)

func NewWorkerId(workerIndex int, processorName string, topicName string) string {
	id := uuid.New()
	fullName := fmt.Sprintf("w_%s_%d_%s_%s", processorName, workerIndex, topicName, id.String())
	encodedName := base64.StdEncoding.EncodeToString([]byte(fullName))
	return encodedName
}

func NewRecordId() fmt.Stringer {
	return uuid.New()
}

func BatchId(ctx context.Context, watermark int64) string {
	workerId, success := WorkerId(ctx)
	if !success {
		panic("context does not have worker id")
	}
	return BatchIdFromWorkerId(workerId, watermark)
}

func BatchIdFromWorkerId(workerId string, watermark int64) string {
	return fmt.Sprintf("%s-%v", workerId, watermark-1288834974)
}
