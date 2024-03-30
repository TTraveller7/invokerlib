package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/TTraveller7/invokerlib/pkg/models"
)

func asyncJoin(ctx context.Context, workerIndex int, processFunc models.JoinCallback, errCh chan<- error,
	wg *sync.WaitGroup, workerNotifyChannel <-chan string, workerReadyChannel chan<- bool) {

	logs := log.New(os.Stdout, fmt.Sprintf("[async join worker #%v] ", workerIndex), log.LstdFlags|log.Lshortfile)

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
}
