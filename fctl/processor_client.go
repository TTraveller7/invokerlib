package main

import (
	"net/http"

	"github.com/TTraveller7/invokerlib"
)

type ProcessorClient struct {
	invokerlib.InvokerClient
}

func NewProcessorClient(name string) *ProcessorClient {
	return &ProcessorClient{
		InvokerClient: invokerlib.InvokerClient{
			Cli: http.DefaultClient,
			Url: ConcatPath(conf.FissionRouter, name),
		},
	}
}

func (pc *ProcessorClient) Ping() (*invokerlib.InvokerResponse, error) {
	resp, err := pc.SendCommand(make(map[string]any, 0), invokerlib.ProcessorCommands.Ping)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
