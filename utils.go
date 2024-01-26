package invokerlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// MarshalToReader transforms a struct to io.Reader.
func MarshalToReader(s any) (io.Reader, error) {
	sBytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(sBytes), nil
}

func SafeJsonIndent(s any) string {
	if s == nil {
		return ""
	}
	marshalledS, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		return ""
	}
	return string(marshalledS)
}

func WorkloadFileName(workloadFilePath string) string {
	levels := strings.Split(workloadFilePath, "/")
	fileName := levels[len(levels)-1]
	realFileName := fmt.Sprintf("invoker_workload_%s", fileName)
	return realFileName
}
