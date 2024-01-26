package main

import (
	"fmt"
	"strings"
)

type FctlConfig struct {
	FctlHome             string `yaml:"fctlHome"`
	MonitorDirectoryPath string `yaml:"monitorDirectoryPath"`
	FissionRouter        string `yaml:"fissionRouter"`
	RootConfigPath       string `yaml:"rootConfigPath"`
}

func defaultFctlConfig() (*FctlConfig, error) {
	// get fission router port
	fissionPort, err := Exec("kubectl", "get", "svc", "router",
		"-n", "fission",
		"-o", "jsonpath='{...nodePort}'")
	if err != nil {
		return nil, fmt.Errorf("get fission port failed: %v", err)
	}
	fissionPort = strings.Trim(fissionPort, "'")

	return &FctlConfig{
		FctlHome:             FctlHome,
		MonitorDirectoryPath: ConcatPath(FctlHome, MonitorDirectoryName),
		FissionRouter:        fmt.Sprintf("http://127.0.0.1:%v", fissionPort),
	}, nil
}
