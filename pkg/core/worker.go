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
)

func Work(ctx context.Context, workerIndex int, processFunc models.ProcessCallback, errCh chan<- error, wg *sync.WaitGroup,
	workerNotifyChannel <-chan string, workerReadyChannel chan<- bool) {

	logs := log.New(os.Stdout, fmt.Sprintf("[worker #%v] ", workerIndex), log.LstdFlags|log.Lshortfile)

	var workerErr error
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			workerErr = fmt.Errorf("%v", recoverErr)
			logs.Printf("recovered. error: %v", workerErr)
		}
		if workerErr != nil {
			errCh <- workerErr
		}
		wg.Done()
		logs.Printf("ends")
	}()
	logs.Printf("starts")

	c := conf.Config()
	consumerGroup, err := sarama.NewConsumerGroup([]string{c.InputKafkaConfig.Address}, c.Name, consumerConfig())
	if err != nil {
		workerErr = fmt.Errorf("initialize consumer group failed: %v", err)
		logs.Printf("%v", workerErr)
		return
	}

	isReady := false
	setupFunc := func() error {
		if !isReady {
			isReady = true
			// signals to the main routine that this worker joins the consumer group already
			close(workerReadyChannel)
		}
		return nil
	}
	consumeFunc := func(record *models.Record) (consumeFuncErr error) {
		defer func() {
			if consumeFuncRecoverErr := recover(); consumeFuncRecoverErr != nil {
				consumeFuncErr = fmt.Errorf("consumerFunc recovered from panic: %v", consumeFuncRecoverErr)
			}
		}()
		consumeFuncErr = processFunc(ctx, record)
		return
	}
	consumerGroupHandler := NewConsumerGroupHandler(logs, setupFunc, consumeFunc, workerNotifyChannel)

	count := 0
	for {
		count++
		logs.Printf("consume loop #%v starts", count)

		if err := consumerGroup.Consume(ctx, []string{c.InputKafkaConfig.Topic}, consumerGroupHandler); err == ErrConsumerNotify {
			return
		} else if err != nil {
			logs.Printf("consume returns error: %v", err)
			workerErr = err
			return
		}
	}
}
