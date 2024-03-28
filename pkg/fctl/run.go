package main

import "github.com/TTraveller7/invokerlib/pkg/api"

func RunProcessors() {
	cli := NewMonitorClient()
	logs.Printf("sending command runProcessors to monitor")
	resp, err := cli.RunProcessors()
	if err != nil {
		logs.Printf("runProcessors failed: %v", err)
		return
	} else if resp.Code != api.ResponseCodes.Success {
		logs.Printf("runProcessors failed with resp: %+v", resp)
		return
	}
	logs.Printf("runProcessors finished with resp: %+v", resp)
}
