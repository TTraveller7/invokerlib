package fctl

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

	if os.Args[1] == "init" {
		initFctl()
		return
	}

	if err := check(); err != nil {
		return
	}
	MonitorDirectoryPath = ConcatPath(FctlHome, MonitorDirectoryName)

	switch os.Args[1] {
	case "create":
		create()
	}
}

func check() error {
	homePath := os.Getenv("HOME")
	FctlHome = ConcatPath(homePath, FctlDirectoryName)
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
