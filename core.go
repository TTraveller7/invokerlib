package invokerlib

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
)

type ProcessFunc func(ctx context.Context, record *Record) error

// Config for the function. Each function has one function config.
var (
	initMut sync.Mutex

	conf        *FollowerFunctionConfig
	processFunc ProcessFunc
	logs        *log.Logger

	workerNotifyChannels []chan<- string
	funcErrCh            chan error
	wg                   *sync.WaitGroup
)

func Initialize(pf ProcessFunc) error {
	canLock := initMut.TryLock()
	if !canLock {
		return fmt.Errorf("another process is holding the initialization lock")
	}
	defer initMut.Unlock()
	if isInitialized() {
		return fmt.Errorf("invokerlib is already initialized")
	}

	// prevent another routine from entering transition
	resetFunc, err := startTransition()
	if err != nil {
		return fmt.Errorf("start transition failed: %v", err)
	}
	defer resetFunc()

	// TODO: query conf from leader function
	if conf.FunctionName == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	if conf.NumOfWorker == 0 {
		return fmt.Errorf("number of worker must be bigger than zero")
	}
	processFunc = pf
	logs = log.New(os.Stdout, fmt.Sprintf("[%s] ", conf.FunctionName), log.LstdFlags|log.Lshortfile)
	if err := initConsumer(); err != nil {
		return err
	}
	if err := initProducers(); err != nil {
		return err
	}

	workerNotifyChannels = make([]chan<- string, 0)
	funcErrCh = make(chan error, conf.NumOfWorker)
	wg = &sync.WaitGroup{}

	if err := transitToInitialized(); err != nil {
		return err
	}
	return nil
}

func Run(ctx context.Context) (err error) {
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			err = fmt.Errorf("%v", recoverErr)
			logs.Printf("exits with panic: %v", err)
		}
	}()

	resetFunc, transitionErr := startTransition()
	if transitionErr != nil {
		return fmt.Errorf("start transition failed: %v", transitionErr)
	}
	defer resetFunc()

	for i := 0; i < conf.NumOfWorker; i++ {
		wg.Add(1)
		workerCtx := NewWorkerContext(ctx, i)
		// TODO: block or non-block?
		workerNotifyChannel := make(chan string, 10)
		workerNotifyChannels = append(workerNotifyChannels, workerNotifyChannel)
		go Work(workerCtx, i, processFunc, funcErrCh, wg, workerNotifyChannel)
	}

	if transitionErr := transitToRunning(); transitionErr != nil {
		return transitionErr
	}

	return
}

func Exit() {
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			logs.Printf("exits with panic: %v", recoverErr)
		}
	}()

	resetFunc, transitionErr := startTransition()
	if transitionErr != nil {
		logs.Printf("start transition failed: %v", transitionErr)
		return
	}
	defer resetFunc()

	// stop workers
	for _, nc := range workerNotifyChannels {
		nc <- "exit"
	}
	wg.Done()

	// stop consumers
	closeConsumerGroup()

	// stop producers
	closeProducers()

	if transitionErr := transitToExited(); transitionErr != nil {
		logs.Printf("transit to exited failed: %v", transitionErr)
	}
}
