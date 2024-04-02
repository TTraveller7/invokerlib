package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/TTraveller7/invokerlib/pkg/consts"
	"github.com/TTraveller7/invokerlib/pkg/models"
	"github.com/TTraveller7/invokerlib/pkg/state"
	"github.com/TTraveller7/invokerlib/pkg/utils"
	"github.com/bytedance/sonic"
)

type Cron struct {
	t            time.Ticker
	done         <-chan bool
	w            *Watermark
	windowSize   int64
	tickInterval time.Duration
	available    bool
	logs         *log.Logger
}

func NewCron(tickInterval time.Duration, windowSize int64, w *Watermark, done <-chan bool) *Cron {
	ticker := time.NewTicker(tickInterval)
	return &Cron{
		t:            *ticker,
		done:         done,
		w:            w,
		windowSize:   windowSize,
		tickInterval: tickInterval,
		available:    true,
		logs:         log.New(os.Stdout, "cron", log.LstdFlags|log.Lshortfile),
	}
}

func (c *Cron) run(ctx context.Context, joinCallback models.JoinCallback, stateStore state.StateStore) {
	defer func() {
		c.t.Stop()
		c.available = false
	}()
	if !c.available {
		panic("cron can only be run once")
	}
	for c.w.Get()+c.windowSize < time.Now().Unix() {
		c.w.Advance()
	}
	for {
		select {
		case <-c.done:
			c.logs.Printf("cron exits by done")
			return
		case <-ctx.Done():
			c.logs.Printf("cron exits by context done")
			return
		case ts := <-c.t.C:
			watermark := c.w.Get()
			if watermark+c.windowSize < ts.Unix() {
				c.w.Advance()
				go asyncJoin(ctx, watermark, joinCallback, stateStore)
				c.logs.Printf("watermark advanced: %v to %v", watermark, watermark+c.windowSize)
			}
		}
	}
}

func asyncJoin(ctx context.Context, watermark int64, joinCallback models.JoinCallback, stateStore state.StateStore) {
	asyncJoinPrefix := fmt.Sprintf("[async join at %v] ", watermark)
	logs := log.New(os.Stdout, asyncJoinPrefix, log.LstdFlags|log.Lshortfile)

	leftBatchIds := make([]string, 0)
	rightBatchIds := make([]string, 0)
	workerMetaMu.RLock()
	defer workerMetaMu.RUnlock()
	for _, workerMeta := range workerMetas {
		batchId := utils.BatchIdFromWorkerId(workerMeta.WorkerId, watermark)
		if workerMeta.TopicIndex == 0 {
			leftBatchIds = append(leftBatchIds, batchId)
		} else {
			rightBatchIds = append(rightBatchIds, batchId)
		}
	}
	defer func() {
		for _, batchId := range leftBatchIds {
			if err := stateStore.Delete(ctx, batchId); err != nil {
				logs.Printf("async join delete keySet record failed: %v", err)
			}
		}
		for _, batchId := range rightBatchIds {
			if err := stateStore.Delete(ctx, batchId); err != nil {
				logs.Printf("async join delete keySet record failed: %v", err)
			}
		}
	}()

	// fetch key sets
	leftKeySets, err := fetchKeySets(ctx, stateStore, leftBatchIds)
	if err != nil {
		logs.Printf("async join fetch key sets failed: %v", err)
		return
	}
	defer func() {
		for _, key := range leftKeySets {
			if err := stateStore.Delete(ctx, key); err != nil {
				logs.Printf("async join delete left record failed: %v", err)
			}
		}
	}()
	rightKeySets, err := fetchKeySets(ctx, stateStore, rightBatchIds)
	if err != nil {
		logs.Printf("async join fetch key sets failed: %v", err)
		return
	}
	defer func() {
		for _, key := range rightKeySets {
			if err := stateStore.Delete(ctx, key); err != nil {
				logs.Printf("async join delete right record failed: %v", err)
			}
		}
	}()

	// fetch records
	leftRecords, err := fetchRecords(ctx, stateStore, leftKeySets)
	if err != nil {
		logs.Printf("async join fetch records failed: %v", err)
	}

	// join
	for _, rightKey := range rightKeySets {
		val, err := stateStore.Get(ctx, rightKey)
		if err == consts.ErrStateStoreKeyNotExist {
			logs.Printf("cache miss, key=%s", rightKey)
			continue
		} else if err != nil {
			logs.Printf("async join fetch record failed: %v", err)
			continue
		}

		rightRecord := models.NewRecord(rightKey, val)
		for _, leftRecord := range leftRecords {
			if err := joinCallback(ctx, leftRecord, rightRecord); err != nil {
				logs.Printf("async join: join callback failed: %v", err)
				return
			}
		}
	}
}

func fetchKeySets(ctx context.Context, stateStore state.StateStore, batchIds []string) ([]string, error) {
	res := make([]string, 0)
	for _, batchId := range batchIds {
		keySet, err := stateStore.Get(ctx, batchId)
		if err == consts.ErrStateStoreKeyNotExist {
			logs.Printf("cache miss, batchId=%s", batchId)
			continue
		} else if err != nil {
			return nil, err
		}

		keySetArr := make([]string, 0)
		sonic.Unmarshal(keySet, &keySetArr)
		res = append(res, keySetArr...)
	}
	return res, nil
}

func fetchRecords(ctx context.Context, stateStore state.StateStore, keySet []string) ([]*models.Record, error) {
	res := make([]*models.Record, 0, len(keySet))
	for _, key := range keySet {
		val, err := stateStore.Get(ctx, key)
		if err == consts.ErrStateStoreKeyNotExist {
			logs.Printf("cache miss, key=%s", key)
			continue
		} else if err != nil {
			return nil, err
		}

		res = append(res, models.NewRecord(key, val))
	}
	return res, nil
}
