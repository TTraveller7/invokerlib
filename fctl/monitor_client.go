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
		url: ConcatPath(FissionRouter, "monitor"),
	}
}

func (m *MonitorClient) LoadRootConfig(conf *invokerlib.RootConfig) (*invokerlib.InvokerResponse, error) {
	params, err := invokerlib.MarshalToParams(conf)
	if err != nil {
		err := fmt.Errorf("monitor client marshal to params failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	req := invokerlib.InvokerRequest{
		Command: invokerlib.MonitorCommands.LoadRootConfig,
		Params:  params,
	}
	reader, err := MarshalToReader(req)
	if err != nil {
		err := fmt.Errorf("monitor client marshal to reader failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	resp, err := http.Post(ConcatPath(m.url, invokerlib.MonitorCommands.LoadRootConfig), MimeTypeJson, reader)
	if err != nil {
		err := fmt.Errorf("monitor client send request failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	} else if resp == nil || resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("monitor client send request failed: resp=%v, status=%s", resp, resp.Status)
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
