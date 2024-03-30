package utils

import (
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
