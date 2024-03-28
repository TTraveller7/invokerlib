package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/TTraveller7/invokerlib/pkg/consts"
	"github.com/TTraveller7/invokerlib/pkg/logs"
	"github.com/TTraveller7/invokerlib/pkg/utils"
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
	reader, err := utils.MarshalToReader(req)
	if err != nil {
		err := fmt.Errorf("marshal to reader failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	resp, err := http.Post(c.Url, consts.MimeTypeJson, reader)
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

func (c *InvokerClient) SendFile(fileName string, file io.Reader) (*InvokerResponse, error) {
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		err = fmt.Errorf("read file failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	b := &bytes.Buffer{}
	multipartWriter := multipart.NewWriter(b)
	fileWriter, err := multipartWriter.CreateFormFile("file", fileName)
	if err != nil {
		err = fmt.Errorf("multipart writer create form file failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}
	if _, err := fileWriter.Write(fileBytes); err != nil {
		err = fmt.Errorf("file writer write failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}
	if err := multipartWriter.Close(); err != nil {
		err = fmt.Errorf("close multipart writer failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	resp, err := http.Post(c.Url, consts.MimeTypeJson, b)
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
