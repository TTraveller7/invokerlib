package invokerlib

import "encoding/json"

var ResponseCodes = struct {
	Success int
	Failed  int
}{
	Success: 0,
	Failed:  1,
}

type InvokerResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// marshalledErrorResponse returns a marshalled json struct containing error message.
func marshalledErrorResponse(err error) []byte {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	resp := InvokerResponse{
		Code:    ResponseCodes.Failed,
		Message: errMsg,
	}
	respBytes, _ := json.Marshal(resp)
	return respBytes
}
