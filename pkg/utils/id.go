package utils

import (
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
)

func NewWorkerId(workerIndex int, processorName string) string {
	id := uuid.New()
	fullName := fmt.Sprintf("w_%s_%d_%s", processorName, workerIndex, id.String())
	encodedName := base64.StdEncoding.EncodeToString([]byte(fullName))
	return encodedName
}
