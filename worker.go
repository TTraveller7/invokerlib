package invokerlib

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
)

func Work(ctx context.Context, workerIndex int, processFunc ProcessCallback, errCh chan<- error, wg *sync.WaitGroup,
	workerNotifyChannel <-chan string, workerReadyChannel chan<- bool) {
	logs := log.New(os.Stdout, fmt.Sprintf("[worker #%v] ", workerIndex), log.LstdFlags|log.Lshortfile)
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			err := fmt.Errorf("%v", recoverErr)
			logs.Printf("recovered. error: %v", err)
			errCh <- err
		}
		wg.Done()
		logs.Printf("ends")
	}()
	logs.Printf("starts")

	isReady := false
	setupFunc := func() error {
		if !isReady {
			isReady = true
			// signals to the main routine that this worker joins the consumer group already
			close(workerReadyChannel)
		}
		return nil
	}
	consumeFunc := func(record *Record) error {
		if err := processFunc(ctx, record); err != nil {
			return err
		}
		return nil
	}
	consumerGroupHandler := NewConsumerGroupHandler(setupFunc, consumeFunc, workerNotifyChannel)

	count := 0
	for {
		count++
		logs.Printf("consume loop #%v starts", count)

		if err := consumerGroup.Consume(ctx, []string{conf.InputKafkaConfig.Topic}, consumerGroupHandler); err == errConsumerNotify {
			return
		} else if err != nil {
			logs.Printf("consume returns error: %v", err)
			errCh <- err
			return
		}
	}
}
