package invokerlib

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
)

type InitFunc func()
type ProcessFunc func(ctx context.Context, record *Record) error

var (
	conf *InternalProcessorConfig

	processFunc ProcessFunc
	logs        *log.Logger

	workerNotifyChannels []chan<- string
	errCh                chan error
	wg                   *sync.WaitGroup
)

func Initialize(internalPc *InternalProcessorConfig, pf ProcessFunc, initF InitFunc) error {
	resetFunc, err := startTransition(functionStates.Initialized)
	if err != nil {
		err = fmt.Errorf("start transition failed: %v", err)
		logs.Printf("%v\n", err)
		return err
	}
	defer resetFunc()

	// validate and load internal processor config
	if err := internalPc.Validate(); err != nil {
		err = fmt.Errorf("validate internal processor config failed: %v", err)
		logs.Printf("%v\n", err)
		return err
	}
	conf = internalPc

	// set logger
	logs = log.New(os.Stdout, fmt.Sprintf("[%s] ", conf.Name), log.LstdFlags|log.Lshortfile)

	// load processFunc
	if pf == nil {
		err = fmt.Errorf("processorFunc is not specified")
		logs.Printf("%v", err)
		return err
	}
	processFunc = pf

	// call init if user has one
	if initF != nil {
		initF()
	}

	// create consumer and producers
	if err := initConsumer(); err != nil {
		err = fmt.Errorf("init consumer failed: %v", err)
		logs.Printf("%v", err)
		return err
	}
	if err := initProducers(); err != nil {
		err = fmt.Errorf("init producers failed: %v", err)
		logs.Printf("%v", err)
		return err
	}

	workerNotifyChannels = make([]chan<- string, 0)
	errCh = make(chan error, conf.NumOfWorker)
	wg = &sync.WaitGroup{}

	if err := transitToInitialized(); err != nil {
		err = fmt.Errorf("transit to initialized failed: %v", err)
		logs.Printf("%v", err)
		return err
	}
	return nil
}

func Run(ctx context.Context) error {
	resetFunc, transitionErr := startTransition(functionStates.Running)
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
		go Work(workerCtx, i, processFunc, errCh, wg, workerNotifyChannel)
	}

	if transitionErr := transitToRunning(); transitionErr != nil {
		err := fmt.Errorf("transit to running failed: %v", transitionErr)
		logs.Printf("%v", err)
		return err
	}

	return nil
}

func Exit() {
	resetFunc, transitionErr := startTransition(functionStates.Exited)
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
