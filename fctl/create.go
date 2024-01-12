package main

import (
	"os"

	"github.com/TTraveller7/invokerlib"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

func create() {
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

	conf := &invokerlib.RootConfig{}
	if err := yaml.Unmarshal(content, conf); err != nil {
		logs.Printf("unmarshal config file failed: %v", err)
		return
	}

	// validate config
	if err := conf.Validate(); err != nil {
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
		"--src", ConcatPath(MonitorDirectoryPath, "go.mod"),
		"--src", ConcatPath(MonitorDirectoryPath, "go.sum"),
		"--src", ConcatPath(MonitorDirectoryPath, "handler.go"))
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

	// load monitor config
	cli := NewMonitorClient()
	if _, err = cli.LoadRootConfig(conf); err != nil {
		logs.Printf("load root config failed: %v", err)
		return
	}

	// create topics
}
