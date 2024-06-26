package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/IBM/sarama"
	"github.com/TTraveller7/invokerlib/pkg/conf"
	"github.com/TTraveller7/invokerlib/pkg/models"
	"github.com/TTraveller7/invokerlib/pkg/utils"
)

type WorkerMeta struct {
	WorkerId   string
	TopicIndex int
	Alive      bool
}

func Work(ctx context.Context, consumerConfig *conf.ConsumerConfig, workerIndex int, processFunc models.ProcessCallback,
	errCh chan<- error, wg *sync.WaitGroup, workerNotifyChannel <-chan string, workerReadyChannel chan<- struct{}) {
	// set up worker logger
	logPrefix := fmt.Sprintf("[worker-%s-%v] ", consumerConfig.Topic, workerIndex)
	logs := log.New(os.Stdout, logPrefix, log.LstdFlags|log.Lshortfile)

	// add worker meta
	workerId, _ := utils.WorkerId(ctx)
	workerMeta := &WorkerMeta{
		WorkerId:   workerId,
		TopicIndex: consumerConfig.TopicIndex,
		Alive:      true,
	}
	addWorkerMeta(workerMeta)

	var workerErr error
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			workerErr = fmt.Errorf("%v", recoverErr)
			logs.Printf("recovered. error: %v", workerErr)
		}
		if workerErr != nil {
			errCh <- workerErr
		}
		close(errCh)
		wg.Done()
		logs.Printf("ends")
	}()
	wg.Add(1)

	logs.Printf("starts")

	// init consumer group
	consumerGroup, err := sarama.NewConsumerGroup([]string{consumerConfig.Address}, consumerConfig.Topic,
		defaultSaramaConsumerConfig())
	if err != nil {
		workerErr = fmt.Errorf("initialize consumer group failed: %v", err)
		logs.Printf("%v", workerErr)
		return
	}
	consumerGroups = append(consumerGroups, consumerGroup)

	setupFunc := func() error {
		return nil
	}
	consumeFunc := func(record *models.Record) (consumeFuncErr error) {
		defer func() {
			if consumeFuncRecoverErr := recover(); consumeFuncRecoverErr != nil {
				consumeFuncErr = fmt.Errorf("consumerFunc recovered from panic: %v", consumeFuncRecoverErr)
			}
			if consumeFuncErr != nil {
				logs.Printf("consumeFunc failed: %v", consumeFuncErr)
			}
		}()
		consumeFuncErr = processFunc(ctx, record)
		return
	}
	consumerGroupHandler := NewConsumerGroupHandler(logs, setupFunc, consumeFunc, workerNotifyChannel, workerReadyChannel)

	count := 0
	for {
		count++
		logs.Printf("consume loop #%v starts", count)

		if err := consumerGroup.Consume(ctx, []string{consumerConfig.Topic}, consumerGroupHandler); err == ErrConsumerNotify {
			return
		} else if err != nil {
			logs.Printf("consume returns error: %v", err)
			workerErr = err
			return
		}
	}
}
