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
	logs.Printf("processor %s client got response for initialize:\n%s", pc.ProcessorName, SafeJsonIndent(resp))
	return resp, nil
}
