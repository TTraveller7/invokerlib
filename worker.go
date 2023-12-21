package invokerlib

import (
	"context"
	"fmt"
	"log"
	"os"
)

func Work(funcCtx context.Context, workerIndex int, processFunc ProcessFunc, errCh chan<- error) {
	logs := log.New(os.Stdout, fmt.Sprintf("[worker #%v] ", workerIndex), log.LstdFlags|log.Lshortfile)
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			err := fmt.Errorf("%v", recoverErr)
			logs.Printf("recovered. error: %v", err)
			errCh <- err
		}
	}()

	ctx := NewWorkerContext(funcCtx, workerIndex)

	consume := func(record *Record) error {
		// process
		processFunc(ctx, record)

		// produce
		return nil
	}
	consumerGroupHandler := NewConsumerGroupHandler(consume)

	for {
		if err := consumerGroup.Consume(ctx, []string{conf.KafkaSrc.Topic}, consumerGroupHandler); err != nil {
			panic(err)
		}
	}
}
