package invokerlib

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
)

func Work(ctx context.Context, workerIndex int, processFunc ProcessFunc, errCh chan<- error, wg *sync.WaitGroup, workerNotifyChannel <-chan string) {
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

	consumerGroupHandler := NewConsumerGroupHandler(func(record *Record) error {
		if err := processFunc(ctx, record); err != nil {
			return err
		}
		return nil
	}, workerNotifyChannel)

	for {
		if err := consumerGroup.Consume(ctx, []string{kafkaSrc.Topic}, consumerGroupHandler); err == errConsumerNotify {
			return
		} else if err != nil {
			logs.Printf("consume returns error: %v", err)
			errCh <- err
			return
		}
	}
}
