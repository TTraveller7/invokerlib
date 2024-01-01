package fctl

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

func ConcatPath(paths ...string) string {
	sb := strings.Builder{}
	for i, p := range paths {
		trimmedPath := strings.TrimSuffix(p, "/")
		sb.WriteString(trimmedPath)
		if i < len(paths)-1 {
			sb.WriteRune('/')
		}
	}
	return sb.String()
}

func MarshalToReader(s any) (io.Reader, error) {
	sBytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(sBytes), nil
}
