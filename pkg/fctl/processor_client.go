package main

import (
	"net/http"

	"github.com/TTraveller7/invokerlib/pkg/api"
)

type ProcessorClient struct {
	api.InvokerClient
}

func NewProcessorClient(name string) *ProcessorClient {
	return &ProcessorClient{
		InvokerClient: api.InvokerClient{
			Cli: http.DefaultClient,
			Url: ConcatPath(c.FissionRouter, name),
		},
	}
}

func (pc *ProcessorClient) Ping() (*api.InvokerResponse, error) {
	resp, err := pc.SendCommand(make(map[string]any, 0), api.ProcessorCommands.Ping)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
