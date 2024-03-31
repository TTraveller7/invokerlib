package core

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/TTraveller7/invokerlib/pkg/conf"
	"github.com/TTraveller7/invokerlib/pkg/consts"
	"github.com/TTraveller7/invokerlib/pkg/logs"
	"github.com/TTraveller7/invokerlib/pkg/models"
	"github.com/TTraveller7/invokerlib/pkg/state"
	"github.com/TTraveller7/invokerlib/pkg/utils"
)

var (
	processorCallbacks *models.ProcessorCallbacks

	// channels to send instructions to workers
	workerNotifyChannels []chan<- string

	// channels for workers to send errors out
	workerErrorChannels []<-chan error

	workerReadyChannels []<-chan struct{}

	wg *sync.WaitGroup

	processorCtx context.Context

	metricsClient *utils.MetricsClient

	workerMetas map[string]*WorkerMeta

	cronDone chan<- bool
)

func Initialize(internalPc *conf.InternalProcessorConfig, pc *models.ProcessorCallbacks) error {
	resetFunc, err := startTransition(functionStates.Initialized)
	if err != nil {
		err = fmt.Errorf("start transition failed: %v", err)
		logs.Printf("%v\n", err)
		return err
	}
	defer resetFunc()

	// validate and load internal processor config
	if err := internalPc.Validate(); err != nil {
		logs.Printf("%v\n", err)
		return err
	}
	if err := conf.LoadConfig(internalPc); err != nil {
		logs.Printf("%v\n", err)
		return err
	}
	c := conf.Config()
	logs.Printf("internalProcessorConfig: %s", utils.SafeJsonIndent(c))

	// set logger prefix
	logs.SetPrefix(fmt.Sprintf("[%s] ", c.Name))

	// set metrics
	metricsClient = utils.NewMetricsClient(c.Name)

	processorCtx = context.Background()
	workerMetas = make(map[string]*WorkerMeta, 0)
	workerNotifyChannels = make([]chan<- string, 0)
	workerErrorChannels = make([]<-chan error, 0)
	wg = &sync.WaitGroup{}

	// load processor callbacks
	if pc == nil {
		err = fmt.Errorf("processor callbacks are not specified")
		logs.Printf("%v", err)
		return err
	}
	switch c.Type {
	case consts.ProcessorTypeProcess:
		if pc.Process == nil {
			err = fmt.Errorf("process in processor callbacks is not specified")
			logs.Printf("%v", err)
			return err
		}
	case consts.ProcessorTypeJoin:
		if pc.Join == nil {
			err = fmt.Errorf("join in processor callbacks is not specified")
			logs.Printf("%v", err)
			return err
		}
	default:
		return consts.ErrProcessorTypeNotRecognized
	}
	processorCallbacks = pc

	// init consumer
	if err := initConsumer(); err != nil {
		err = fmt.Errorf("init consumer failed: %v", err)
		logs.Printf("%v", err)
		return err
	}
	logs.Printf("consumers start")

	// create producers
	if err := initProducers(); err != nil {
		err = fmt.Errorf("init producers failed: %v", err)
		logs.Printf("%v", err)
		return err
	}
	logs.Printf("producers start")

	// call OnInit if user has one
	if processorCallbacks.OnInit != nil {
		if err := doOnInit(processorCallbacks.OnInit); err != nil {
			err = fmt.Errorf("user callback OnInit failed: %v", err)
			logs.Printf("%v", err)
			return err
		}
	}

	if err := transitToInitialized(); err != nil {
		err = fmt.Errorf("transit to initialized failed: %v", err)
		logs.Printf("%v", err)
		return err
	}
	return nil
}

func doOnInit(OnInit models.InitCallback) (err error) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = fmt.Errorf("%v. %s", panicErr, string(debug.Stack()))
		}
	}()
	err = OnInit()
	return
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

	c := conf.Config()
	workerTotalCount := 0
	for _, consumerConfig := range c.ConsumerConfigs {
		workerTotalCount += consumerConfig.NumOfWorkers
	}
	workerReadyChannels = make([]<-chan struct{}, workerTotalCount)

	if c.Type == consts.ProcessorTypeProcess {
		for _, consumerConfig := range c.ConsumerConfigs {
			for i := 0; i < consumerConfig.NumOfWorkers; i++ {
				workerCtx := utils.NewWorkerContext(processorCtx, i, c.Name, consumerConfig.Topic)

				workerNotifyChannel := make(chan string, 10)
				workerNotifyChannels = append(workerNotifyChannels, workerNotifyChannel)

				workerErrorChannel := make(chan error, 1)
				workerErrorChannels = append(workerErrorChannels, workerErrorChannel)

				workerReadyChannel := make(chan struct{}, 1)
				workerReadyChannels = append(workerReadyChannels, workerReadyChannel)

				go Work(workerCtx, consumerConfig, i, processorCallbacks.Process, workerErrorChannel, wg, workerNotifyChannel, workerReadyChannel)

				metricsClient.EmitCounter("worker_num", "Number of workers", 1)
			}
		}
	} else {
		windowSize := int64(c.WindowSize)
		w := NewWatermark(windowSize)

		cd := make(chan bool)
		cronDone = cd
		cronCtx := context.WithValue(processorCtx, "hello", "cron")
		cron := NewCron(5*time.Second, windowSize, w, cd)
		stateStore, err := state.NewRedisStateStore("state-store")
		if err != nil {
			logs.Printf("create redis state store failed: %v", err)
			return err
		}
		cron.run(cronCtx, processorCallbacks.Join, stateStore)

		for _, consumerConfig := range c.ConsumerConfigs {
			for i := 0; i < consumerConfig.NumOfWorkers; i++ {
				workerCtx := utils.NewWorkerContext(processorCtx, i, c.Name, consumerConfig.Topic)

				workerNotifyChannel := make(chan string, 10)
				workerNotifyChannels = append(workerNotifyChannels, workerNotifyChannel)

				workerErrorChannel := make(chan error, 1)
				workerErrorChannels = append(workerErrorChannels, workerErrorChannel)

				workerReadyChannel := make(chan struct{}, 1)
				workerReadyChannels = append(workerReadyChannels, workerReadyChannel)

				joinWorker := NewJoinWorker(w, stateStore, int(3*windowSize))
				go Work(workerCtx, consumerConfig, i, joinWorker.JoinWorkerProcessCallback, workerErrorChannel, wg, workerNotifyChannel, workerReadyChannel)

				metricsClient.EmitCounter("worker_num", "Number of workers", 1)
			}
		}
	}

	for i := 0; i < workerTotalCount; i++ {
		select {
		case workerErr := <-workerErrorChannels[i]:
			// some worker does not start successfully, try to exit

			// reset previous transition
			resetFunc()
			hasReset = true

			// stop workers, close producers and consumer group
			Exit()

			return fmt.Errorf("start worker #%v failed: %v", i, workerErr)
		default:
		}
	}

	logs.Printf("all %v orkers start successfully", workerTotalCount)

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
	wg.Wait()
	for _, errChan := range workerErrorChannels {
		for err := range errChan {
			logs.Printf("worker exited with error: %v", err)
		}
	}

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
