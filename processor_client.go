package invokerlib

import "net/http"

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
	params := make(map[string]any, 0)
	resp, err := pc.SendCommand(params, ProcessorCommands.Initialize)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
