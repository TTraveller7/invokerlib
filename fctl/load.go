package main

import (
	"github.com/TTraveller7/invokerlib"
	"github.com/spf13/pflag"
)

func Load() {
	logs.Printf("producing workload file to Kafka")

	urlPtr := pflag.StringP("url", "u", "", "workload url")
	namePtr := pflag.StringP("name", "n", "", "workload name")
	topicPtr := pflag.StringP("topic", "t", "", "topic")
	pflag.Parse()
	if urlPtr == nil || len(*urlPtr) == 0 {
		logs.Printf("url is not provided. Use -u <url> to provide workload url. ")
		return
	}
	if namePtr == nil || len(*namePtr) == 0 {
		logs.Printf("name is not provided. Use -n <name> to provide workload name. ")
		return
	}
	if topicPtr == nil || len(*topicPtr) == 0 {
		logs.Printf("topic is not provided. Use -t <topic> to provide target topic. ")
		return
	}

	// send load command
	logs.Println("sending load command to monitor")
	loadParam := &invokerlib.LoadParams{
		Url:    *urlPtr,
		Name:   *namePtr,
		Topics: []string{*topicPtr},
	}
	cmdCli := NewMonitorClient()
	resp, err := cmdCli.Load(loadParam)
	if err != nil {
		logs.Printf("load failed: %v", err)
		return
	} else if resp.Code != invokerlib.ResponseCodes.Success {
		logs.Printf("load failed with resp: %+v", resp)
		return
	}
	logs.Printf("load finished with resp: %+v", resp)
}
