package invokerlib

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
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

func BytesToUint64(bytes []byte) uint64 {
	num, _ := strconv.ParseUint(string(bytes), 10, 64)
	return num
}

func Uint64ToBytes(num uint64) []byte {
	s := strconv.FormatUint(num, 10)
	return []byte(s)
}
