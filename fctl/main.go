package main

import (
	"log"
	"os"
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
	case "run":
		RunProcessors()
	case "load":
		Load()
	}
}

func checkFctlHome() error {
	if _, err := os.Stat(FctlHome); err != nil {
		logs.Printf("load FctlHome failed: %v. Try `fctl init`", err)
		return err
	}
	return nil
}

// stop
// metric
// log
// status
