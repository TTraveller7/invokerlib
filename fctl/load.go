package main

import (
	"os"
	"strings"

	"github.com/TTraveller7/invokerlib"
	"github.com/spf13/pflag"
)

func Load() {
	logs.Printf("producing workload file to Kafka")

	pathPtr := pflag.StringP("workload", "f", "", "workload path")
	topicPtr := pflag.StringP("topic", "t", "", "topic")
	pflag.Parse()
	if pathPtr == nil || len(*pathPtr) == 0 {
		logs.Printf("workload path is not provided. Use -f <workload path> to provide workload path. ")
		return
	}
	if topicPtr == nil || len(*topicPtr) == 0 {
		logs.Printf("topic is not provided. Use -t <topic> to provide target topic. ")
		return
	}

	file, err := os.OpenFile(*pathPtr, os.O_RDONLY, 0644)
	if err != nil {
		logs.Printf("open workload file failed: %v", err)
		return
	}
	levels := strings.Split(*pathPtr, "/")
	fileName := levels[len(levels)-1]

	// upload file to monitor
	logs.Println("uploading to monitor")
	uploadCli := NewMonitorUploadClient()
	resp, err := uploadCli.Upload(fileName, file)
	if err != nil {
		logs.Printf("upload failed: %v", err)
		return
	} else if resp.Code != invokerlib.ResponseCodes.Success {
		logs.Printf("upload failed with resp: %+v", resp)
		return
	}
	logs.Printf("upload finished with resp: %+v", resp)

	// send load command
	logs.Println("sending load command to monitor")
	loadParam := &invokerlib.LoadParams{
		FileName: fileName,
		Topics:   []string{*topicPtr},
	}
	cmdCli := NewMonitorClient()
	resp, err = cmdCli.Load(loadParam)
	if err != nil {
		logs.Printf("load failed: %v", err)
		return
	} else if resp.Code != invokerlib.ResponseCodes.Success {
		logs.Printf("load failed with resp: %+v", resp)
		return
	}
	logs.Printf("load finished with resp: %+v", resp)
}
