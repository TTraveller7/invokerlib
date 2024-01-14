package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	logs *log.Logger = log.New(os.Stdout, "", 0)
	conf *FctlConfig
)

func main() {
	if len(os.Args) == 1 {
		help()
		return
	}

	// set global variable
	homePath := os.Getenv("HOME")
	FctlHome = ConcatPath(homePath, FctlHomeDirectoryName)

	if os.Args[1] == "init" {
		InitFctl()
		return
	}

	// check fctl home and load config
	if err := checkFctlHome(); err != nil {
		return
	}
	if err := loadFctlConfig(); err != nil {
		return
	}

	switch os.Args[1] {
	case "create":
		Create()
	}
}

func checkFctlHome() error {
	if _, err := os.Stat(FctlHome); err != nil {
		logs.Printf("load FctlHome failed: %v. Try `fctl init`", err)
		return err
	}
	return nil
}

func loadFctlConfig() error {
	configFilePath := ConcatPath(FctlHome, ConfigFileName)
	fileContent, err := os.ReadFile(configFilePath)
	if err != nil {
		logs.Printf("read config file failed: path=%s, err=%v", configFilePath, err)
		return err
	}
	logs.Printf("%s", string(fileContent))

	fctlConfig := FctlConfig{}
	if err := yaml.Unmarshal(fileContent, &fctlConfig); err != nil {
		logs.Printf("unmarshal fctl config failed: %v", err)
		return err
	}
	conf = &fctlConfig
	return nil
}

// stop
// metric
// log
// status
