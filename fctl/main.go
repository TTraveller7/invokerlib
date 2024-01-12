package main

import (
	"log"
	"os"
)

var logs *log.Logger = log.New(os.Stdout, "", 0)

func main() {
	if len(os.Args) == 1 {
		help()
		return
	}

	homePath := os.Getenv("HOME")
	FctlHome = ConcatPath(homePath, FctlDirectoryName)
	MonitorDirectoryPath = ConcatPath(FctlHome, MonitorDirectoryName)
	FissionRouter = "http://127.0.0.1"
	if fr := os.Getenv("FISSION_ROUTER"); fr != "" {
		FissionRouter = fr
	}

	if os.Args[1] == "init" {
		initFctl()
		return
	}

	if err := check(); err != nil {
		return
	}

	switch os.Args[1] {
	case "create":
		create()
	}
}

func check() error {
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
