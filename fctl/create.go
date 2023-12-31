package fctl

import (
	"os"

	"github.com/TTraveller7/invokerlib"
	"gopkg.in/yaml.v2"
)

func create() {
	// TODO: check fission status

	// parse config yaml
	configPath := ""
	content, err := os.ReadFile(configPath)
	if err != nil {
		logs.Printf("read config file failed: %v", err)
		return
	}

	conf := &invokerlib.RootConfig{}
	if err := yaml.Unmarshal(content, conf); err != nil {
		logs.Printf("unmarshal config file failed: %v", err)
		return
	}

	// create monitor function
	err = Run("fission", "fn", "create",
		"--name", "monitor",
		"--env", "fctl",
		"--entrypoint", "Handler",
		"--src", ConcatPath(MonitorDirectoryPath, "go.mod"),
		"--src", ConcatPath(MonitorDirectoryPath, "go.sum"),
		"--src", ConcatPath(MonitorDirectoryPath, "handler.go"))
	if err != nil {
		logs.Printf("create monitor function failed: %v", err)
		return
	}

	// load monitor config

}
