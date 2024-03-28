package main

import (
	"fmt"
	"net/http"

	"github.com/TTraveller7/invokerlib/pkg/api"
	"github.com/TTraveller7/invokerlib/pkg/conf"
)

type MonitorClient struct {
	api.InvokerClient
}

func NewMonitorClient() *MonitorClient {
	return &MonitorClient{
		api.InvokerClient{
			Cli: http.DefaultClient,
			Url: ConcatPath(c.FissionRouter, "monitor"),
		},
	}
}

func (m *MonitorClient) LoadRootConfig(conf *conf.RootConfig) (*api.InvokerResponse, error) {
	params, err := api.MarshalToParams(conf)
	if err != nil {
		err := fmt.Errorf("monitor client marshal to params failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	resp, err := m.SendCommand(params, api.MonitorCommands.LoadRootConfig)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MonitorClient) CreateTopics() (*api.InvokerResponse, error) {
	params := api.NewInvokerRequestParams()
	resp, err := m.SendCommand(params, api.MonitorCommands.CreateTopics)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MonitorClient) LoadProcessorEndpoints(p *api.LoadProcessorEndpointsParams) (*api.InvokerResponse, error) {
	params, err := api.MarshalToParams(p)
	if err != nil {
		err := fmt.Errorf("monitor client marshal to params failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	resp, err := m.SendCommand(params, api.MonitorCommands.LoadProcessorEndpoints)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MonitorClient) InitializeProcessors() (*api.InvokerResponse, error) {
	params := api.NewInvokerRequestParams()
	resp, err := m.SendCommand(params, api.MonitorCommands.InitializeProcessors)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MonitorClient) RunProcessors() (*api.InvokerResponse, error) {
	params := api.NewInvokerRequestParams()
	resp, err := m.SendCommand(params, api.MonitorCommands.RunProcessors)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MonitorClient) Load(p *api.LoadParams) (*api.InvokerResponse, error) {
	params, err := api.MarshalToParams(p)
	if err != nil {
		err := fmt.Errorf("monitor client marshal to params failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	resp, err := m.SendCommand(params, api.MonitorCommands.Load)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MonitorClient) CatProcessor(p *api.CatProcessorParams) (*api.InvokerResponse, error) {
	params, err := api.MarshalToParams(p)
	if err != nil {
		err := fmt.Errorf("monitor client marshal to params failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	resp, err := m.SendCommand(params, api.MonitorCommands.CatProcessor)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
