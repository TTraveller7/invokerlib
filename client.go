package invokerlib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type InvokerClient struct {
	Cli *http.Client
	Url string
}

func (c *InvokerClient) SendCommand(params map[string]any, command string) (*InvokerResponse, error) {
	req := InvokerRequest{
		Command: command,
		Params:  params,
	}
	reader, err := MarshalToReader(req)
	if err != nil {
		err := fmt.Errorf("marshal to reader failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	resp, err := http.Post(c.Url, MimeTypeJson, reader)
	if err != nil {
		err := fmt.Errorf("send request failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	} else if resp == nil || resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("send request failed: url=%s, resp=%v, status=%s", c.Url, resp, resp.Status)
		logs.Printf("%v", err)
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("read response body failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}
	invokerResp := &InvokerResponse{}
	if err := json.Unmarshal(body, invokerResp); err != nil {
		err := fmt.Errorf("unmarshal response body failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	return invokerResp, nil
}
