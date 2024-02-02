package invokerlib

import (
	"os"
)

var ResponseCodes = struct {
	Success int
	Failed  int
}{
	Success: 0,
	Failed:  1,
}

type InvokerResponse struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	HostName string `json:"host_name"`
}

type ProcessorCatResult map[string][]StateStoreEntry

type StateStoreEntry struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

func successResponse() *InvokerResponse {
	return &InvokerResponse{
		Code:     ResponseCodes.Success,
		HostName: os.Getenv("HOSTNAME"),
	}
}

func failureResponse(err error) *InvokerResponse {
	return &InvokerResponse{
		Code:     ResponseCodes.Failed,
		HostName: os.Getenv("HOSTNAME"),
		Message:  err.Error(),
	}
}
