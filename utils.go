package invokerlib

import (
	"bytes"
	"encoding/json"
	"io"
)

func MarshalToReader(s any) (io.Reader, error) {
	sBytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(sBytes), nil
}
