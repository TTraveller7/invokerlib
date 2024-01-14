package main

import (
	"os"
	"time"

	"github.com/TTraveller7/invokerlib"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

func Create() {
	pathPtr := pflag.StringP("config", "c", "", "yaml config path")

	pflag.Parse()
	if pathPtr == nil || len(*pathPtr) == 0 {
		logs.Printf("config path is not provided. Use -c <config path> to provide config path. ")
		return
	}

	// TODO: check fission status

	// parse config yaml
	content, err := os.ReadFile(*pathPtr)
	if err != nil {
		logs.Printf("read config file failed: %v", err)
		return
	}

	invokerConfig := &invokerlib.RootConfig{}
	if err := yaml.Unmarshal(content, invokerConfig); err != nil {
		logs.Printf("unmarshal config file failed: %v", err)
		return
	}

	// validate config
	if err := invokerConfig.Validate(); err != nil {
		logs.Printf("validate config failed: %v", err)
		return
	}

	fissionStartSuccess := false

	// create fission env
	err = Run("fission", "env", "create",
		"--name", FissionEnv,
		"--image", "ttraveller7/go-env-1.19",
		"--builder", "ttraveller7/go-builder-1.19",
		"--poolsize", "1",
		"--version", "3")
	if err != nil {
		logs.Printf("create fission env failed: %v", err)
		return
	}
	defer func() {
		if !fissionStartSuccess {
			Run("fission", "env", "delete",
				"--name", FissionEnv)
		}
	}()

	// create monitor function
	err = Run("fission", "fn", "create",
		"--name", "monitor",
		"--env", FissionEnv,
		"--entrypoint", "Handler",
		"--src", ConcatPath(conf.MonitorDirectoryPath, "go.mod"),
		"--src", ConcatPath(conf.MonitorDirectoryPath, "go.sum"),
		"--src", ConcatPath(conf.MonitorDirectoryPath, "handler.go"))
	if err != nil {
		logs.Printf("create monitor function failed: %v", err)
		return
	}
	defer func() {
		if !fissionStartSuccess {
			Run("fission", "fn", "delete",
				"--name", "monitor")
		}
	}()

	// create monitor http endpoint
	err = Run("fission", "httptrigger", "create",
		"--name", "monitor-load-root-config",
		"--url", "/monitor/loadRootConfig",
		"--method", "POST",
		"--function", "monitor")
	if err != nil {
		logs.Printf("create monitor httptrigger failed: %v", err)
		return
	}
	defer func() {
		if !fissionStartSuccess {
			Run("fission", "httptrigger", "delete",
				"--name", "monitor-load-root-config")
		}
	}()

	time.Sleep(3 * time.Second)

	// load monitor config
	cli := NewMonitorClient()
	resp, err := cli.LoadRootConfig(invokerConfig)
	if err != nil {
		logs.Printf("load root config failed: %v", err)
		return
	} else if resp.Code != invokerlib.ResponseCodes.Success {
		logs.Printf("load root config failed: %s", resp.Message)
		return
	}
	logs.Printf("loadRootConfig finished with message: %v", resp.Message)

	fissionStartSuccess = true
}
