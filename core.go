package invokerlib

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
)

type InitCallback func()
type ProcessCallback func(ctx context.Context, record *Record) error
type ExitCallback func()

type ProcessorCallbacks struct {
	OnInit  InitCallback
	Process ProcessCallback
	OnExit  ExitCallback
}

var (
	conf *InternalProcessorConfig

	processorCallbacks *ProcessorCallbacks
	logs               *log.Logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	workerNotifyChannels []chan<- string
	errCh                chan error
	wg                   *sync.WaitGroup
	processorCtx         context.Context
)

func Initialize(internalPc *InternalProcessorConfig, pc *ProcessorCallbacks) error {
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
	logs.Printf("internalProcessorConfig: %s", SafeJsonIndent(conf))

	// set logger
	logs = log.New(os.Stdout, fmt.Sprintf("[%s] ", conf.Name), log.LstdFlags|log.Lshortfile)

	processorCtx = context.Background()
	workerNotifyChannels = make([]chan<- string, 0)
	errCh = make(chan error, conf.NumOfWorker)
	wg = &sync.WaitGroup{}

	// load processFunc
	if pc == nil {
		err = fmt.Errorf("processor callbacks are not specified")
		logs.Printf("%v", err)
		return err
	}
	if pc.Process == nil {
		err = fmt.Errorf("process in processor callbacks is not specified")
		logs.Printf("%v", err)
		return err
	}
	processorCallbacks = pc

	// create consumer and producers
	if err := initConsumer(); err != nil {
		err = fmt.Errorf("init consumer failed: %v", err)
		logs.Printf("%v", err)
		return err
	}
	logs.Printf("consumer starts")
	if err := initProducers(); err != nil {
		err = fmt.Errorf("init producers failed: %v", err)
		logs.Printf("%v", err)
		return err
	}
	logs.Printf("producers start")

	// call OnInit if user has one
	if processorCallbacks.OnInit != nil {
		processorCallbacks.OnInit()
	}

	if err := transitToInitialized(); err != nil {
		err = fmt.Errorf("transit to initialized failed: %v", err)
		logs.Printf("%v", err)
		return err
	}
	return nil
}

func Run() error {
	resetFunc, transitionErr := startTransition(functionStates.Running)
	if transitionErr != nil {
		return fmt.Errorf("start transition failed: %v", transitionErr)
	}
	hasReset := false
	defer func() {
		if !hasReset {
			resetFunc()
		}
	}()

	for i := 0; i < conf.NumOfWorker; i++ {
		workerCtx := NewWorkerContext(processorCtx, i)
		// TODO: block or non-block?

		workerNotifyChannel := make(chan string, 10)
		workerNotifyChannels = append(workerNotifyChannels, workerNotifyChannel)

		workerReadyChannel := make(chan bool)

		go Work(workerCtx, i, processorCallbacks.Process, errCh, wg, workerNotifyChannel, workerReadyChannel)

		select {
		case <-workerReadyChannel:
			logs.Printf("worker #%v starts successfully", i)
		case workerErr := <-errCh:
			// this worker does not start successfully, try to exit

			// reset previous transition
			resetFunc()
			hasReset = true

			// stop workers, close producers and consumer group
			Exit()

			return fmt.Errorf("start worker #%v failed: %v", i, workerErr)
		}

		wg.Add(1)
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

	// stop consumer group
	closeConsumerGroup()

	// stop producers
	closeProducers()

	// call OnExit if user has one
	if processorCallbacks.OnExit != nil {
		processorCallbacks.OnExit()
	}

	if transitionErr := transitToExited(); transitionErr != nil {
		logs.Printf("transit to exited failed: %v", transitionErr)
	}
}
