package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/TTraveller7/invokerlib/pkg/conf"
	"gopkg.in/yaml.v3"
)

func ConcatPath(paths ...string) string {
	sb := strings.Builder{}
	for i, p := range paths {
		trimmedPath := strings.TrimSuffix(p, "/")
		sb.WriteString(trimmedPath)
		if i < len(paths)-1 {
			sb.WriteRune('/')
		}
	}
	return sb.String()
}

func saveFctlConfig(conf *FctlConfig) error {
	configFilePath := ConcatPath(FctlHome, ConfigFileName)
	return saveYaml(conf, configFilePath)
}

func saveYaml(s any, path string) error {
	marshalledStruct, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal default config failed: %v", err)
	}
	if err := os.WriteFile(path, marshalledStruct, 0755); err != nil {
		return fmt.Errorf("write default config failed: %v", err)
	}
	return nil
}

func loadFctlConfig() error {
	configFilePath := ConcatPath(FctlHome, ConfigFileName)
	fctlConfig := &FctlConfig{}
	if err := loadYaml(fctlConfig, configFilePath); err != nil {
		return fmt.Errorf("load yaml failed: %v", err)
	}
	c = fctlConfig
	return nil
}

func getRootConfig() (*conf.RootConfig, error) {
	if c == nil {
		return nil, fmt.Errorf("fctl config is empty. Try fctl init")
	}
	if c.RootConfigPath == "" {
		return nil, fmt.Errorf("root config path is empty. Try fctl create -c <config path>")
	}

	rootConfig := &conf.RootConfig{}
	if err := loadYaml(rootConfig, c.RootConfigPath); err != nil {
		return nil, fmt.Errorf("load yaml failed: %v", err)
	}
	return rootConfig, nil
}

func loadYaml(sPtr any, path string) error {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file failed: path=%s, err=%v", path, err)
	}
	logs.Printf("%s", string(fileContent))

	if err := yaml.Unmarshal(fileContent, sPtr); err != nil {
		return fmt.Errorf("unmarshal yaml failed: %v", err)
	}
	return nil
}

func getProcessorEndpointName(processorName string) string {
	return strings.ToLower(processorName)
}
