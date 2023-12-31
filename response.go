package invokerlib

import "encoding/json"

type ErrorResponse struct {
	Err string
}

// marshalledErrorResponse returns a marshalled json struct containing error message.
func marshalledErrorResponse(err error) []byte {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	resp := ErrorResponse{
		Err: errMsg,
	}
	respBytes, _ := json.Marshal(resp)
	return respBytes
}
