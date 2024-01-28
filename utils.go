package invokerlib

import (
	"bytes"
	"encoding/json"
	"io"
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
	marshalledS, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return ""
	}
	return string(marshalledS)
}
