package api

import (
	"fmt"
	"net/http"

	"github.com/TTraveller7/invokerlib/pkg/conf"
)

type ProcessorClient struct {
	InvokerClient

	ProcessorName string
}

func NewProcessorClient(processorName, url string) *ProcessorClient {
	return &ProcessorClient{
		InvokerClient: InvokerClient{
			Cli: http.DefaultClient,
			Url: url,
		},
		ProcessorName: processorName,
	}
}

func (pc *ProcessorClient) Initialize() (*InvokerResponse, error) {
	// build internal processor config from root config
	ipc := conf.NewInternalProcessorConfig(rootConfig, pc.ProcessorName)
	if err := ipc.Validate(); err != nil {
		err := fmt.Errorf("validate config failed: processorName=%s, error=%v", pc.ProcessorName, err)
		logs.Printf("%v", err)
		return nil, err
	}

	params, err := MarshalToParams(ipc)
	if err != nil {
		err = fmt.Errorf("marshal InternalProcessorConfig to params failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	resp, err := pc.SendCommand(params, ProcessorCommands.Initialize)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (pc *ProcessorClient) Run() (*InvokerResponse, error) {
	resp, err := pc.SendCommand(NewInvokerRequestParams(), ProcessorCommands.Run)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (pc *ProcessorClient) Cat() (*InvokerResponse, error) {
	resp, err := pc.SendCommand(NewInvokerRequestParams(), ProcessorCommands.Cat)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
