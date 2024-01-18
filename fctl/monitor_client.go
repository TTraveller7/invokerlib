package main

import (
	"fmt"
	"net/http"

	"github.com/TTraveller7/invokerlib"
)

type MonitorClient struct {
	invokerlib.InvokerClient
}

func NewMonitorClient() *MonitorClient {
	return &MonitorClient{
		invokerlib.InvokerClient{
			Cli: http.DefaultClient,
			Url: ConcatPath(conf.FissionRouter, "monitor"),
		},
	}
}

func (m *MonitorClient) LoadRootConfig(conf *invokerlib.RootConfig) (*invokerlib.InvokerResponse, error) {
	params, err := invokerlib.MarshalToParams(conf)
	if err != nil {
		err := fmt.Errorf("monitor client marshal to params failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	resp, err := m.SendCommand(params, invokerlib.MonitorCommands.LoadRootConfig)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MonitorClient) CreateTopics() (*invokerlib.InvokerResponse, error) {
	params := make(map[string]any, 0)
	resp, err := m.SendCommand(params, invokerlib.MonitorCommands.CreateTopics)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MonitorClient) LoadProcessorEndpoints(p *invokerlib.LoadProcessorEndpointsParams) (*invokerlib.InvokerResponse, error) {
	params, err := invokerlib.MarshalToParams(p)
	if err != nil {
		err := fmt.Errorf("monitor client marshal to params failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	resp, err := m.SendCommand(params, invokerlib.MonitorCommands.LoadProcessorEndpoints)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MonitorClient) InitializeProcessors() (*invokerlib.InvokerResponse, error) {
	params := make(map[string]any, 0)
	resp, err := m.SendCommand(params, invokerlib.MonitorCommands.InitializeProcessors)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
