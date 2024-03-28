package main

import (
	"github.com/TTraveller7/invokerlib/pkg/api"
	"github.com/TTraveller7/invokerlib/pkg/utils"
	"github.com/spf13/pflag"
)

func Cat() {
	processorPtr := pflag.StringP("processor", "p", "", "processor name")
	pflag.Parse()
	if processorPtr == nil || len(*processorPtr) == 0 {
		logs.Printf("processor is not provided. Use -p <processor> to provide processor name. ")
		return
	}

	// send cat processor command
	catProcessorParam := &api.CatProcessorParams{
		ProcessorName: *processorPtr,
	}
	cmdCli := NewMonitorClient()
	resp, err := cmdCli.CatProcessor(catProcessorParam)
	if err != nil {
		logs.Printf("cat failed: %v", err)
		return
	} else if resp.Code != api.ResponseCodes.Success {
		logs.Printf("cat failed with resp: %+v", resp)
		return
	}
	logs.Printf("%+v", utils.SafeJsonIndent(resp))
}
