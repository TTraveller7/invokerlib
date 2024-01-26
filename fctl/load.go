package main

import "github.com/TTraveller7/invokerlib"

func Load() {
	cli := NewMonitorClient()
	logs.Printf("sending command load to monitor")
	resp, err := cli.Load()
	if err != nil {
		logs.Printf("load failed: %v", err)
		return
	} else if resp.Code != invokerlib.ResponseCodes.Success {
		logs.Printf("load failed with resp: %+v", resp)
		return
	}
	logs.Printf("load finished with resp: %+v", resp)
}
