package invokerlib

import (
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
)

func NewWorkerId(workerIndex int) string {
	id := uuid.New()
	fullName := fmt.Sprintf("w_%s_%d_%s", conf.Name, workerIndex, id.String())
	encodedName := base64.StdEncoding.EncodeToString([]byte(fullName))
	return encodedName
}
