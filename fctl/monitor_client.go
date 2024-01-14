package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/TTraveller7/invokerlib"
)

const MimeTypeJson = "application/json"

type MonitorClient struct {
	cli *http.Client
	url string
}

func NewMonitorClient() *MonitorClient {
	return &MonitorClient{
		cli: http.DefaultClient,
		url: ConcatPath(conf.FissionRouter, "monitor"),
	}
}

func (m *MonitorClient) sendCommand(params map[string]any, command string) (*invokerlib.InvokerResponse, error) {
	req := invokerlib.InvokerRequest{
		Command: command,
		Params:  params,
	}
	reader, err := MarshalToReader(req)
	if err != nil {
		err := fmt.Errorf("monitor client marshal to reader failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	url := ConcatPath(m.url, "loadRootConfig")
	resp, err := http.Post(url, MimeTypeJson, reader)
	if err != nil {
		err := fmt.Errorf("monitor client send request failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	} else if resp == nil || resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("monitor client send request failed: url=%s, resp=%v, status=%s", url, resp, resp.Status)
		logs.Printf("%v", err)
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("monitor client read response body failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}
	invokerResp := &invokerlib.InvokerResponse{}
	if err := json.Unmarshal(body, invokerResp); err != nil {
		err := fmt.Errorf("monitor client unmarshal response body failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	return invokerResp, nil
}

func (m *MonitorClient) LoadRootConfig(conf *invokerlib.RootConfig) (*invokerlib.InvokerResponse, error) {
	params, err := invokerlib.MarshalToParams(conf)
	if err != nil {
		err := fmt.Errorf("monitor client marshal to params failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	resp, err := m.sendCommand(params, invokerlib.MonitorCommands.LoadRootConfig)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *MonitorClient) CreateTopics() (*invokerlib.InvokerResponse, error) {
	params := make(map[string]any, 0)
	resp, err := m.sendCommand(params, invokerlib.MonitorCommands.CreateTopics)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
