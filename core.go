package invokerlib

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
)

// Config for the function. Each function has one function config.
var (
	conf        *Config
	processFunc ProcessFunc
	logs        *log.Logger
	funcErrCh   chan error

	initMut       sync.Mutex
	isInitialized bool = false
)

func Initialize(config *Config) error {
	canLock := initMut.TryLock()
	if !canLock {
		return fmt.Errorf("another process is holding the initialization lock")
	}
	defer initMut.Unlock()
	if isInitialized {
		return fmt.Errorf("invokerlib is already initialized")
	}

	if conf.FunctionName == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	if conf.NumOfWorker == 0 {
		return fmt.Errorf("number of worker must be bigger than zero")
	}
	conf = config
	logs = log.New(os.Stdout, fmt.Sprintf("[%s] ", conf.FunctionName), log.LstdFlags|log.Lshortfile)
	funcErrCh = make(chan error, conf.NumOfWorker)
	InitConsumer()

	isInitialized = true
	return nil
}

func SetProcessRecord(pf ProcessFunc) {
	processFunc = pf
}

func Run(ctx context.Context) {
	logs.Printf("starts")
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			err := fmt.Errorf("%v", recoverErr)
			logs.Printf("exits with panic: %v", err)
		}
	}()

	for i := 0; i < conf.NumOfWorker; i++ {
		go Work(ctx, i, processFunc, funcErrCh)
	}

	for i := 0; i < conf.NumOfWorker; i++ {
		workerErr := <-funcErrCh
		logs.Printf("worker exits with error: %v", workerErr)
	}

	logs.Printf("exits")
}
