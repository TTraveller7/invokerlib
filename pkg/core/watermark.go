package core

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/TTraveller7/invokerlib/pkg/logs"
	"github.com/TTraveller7/invokerlib/pkg/models"
	"github.com/TTraveller7/invokerlib/pkg/state"
	"github.com/TTraveller7/invokerlib/pkg/utils"
)

type Watermark struct {
	t          atomic.Int64
	windowSize int64
}

func NewWatermark(windowSize int64) *Watermark {
	w := &Watermark{
		windowSize: windowSize,
	}
	w.t.Store(time.Now().Unix())
	return w
}

func (w *Watermark) Get() int64 {
	return w.t.Load()
}

func (w *Watermark) Advance() {
	w.t.Add(w.windowSize)
}

type Cron struct {
	t            time.Ticker
	done         <-chan bool
	w            *Watermark
	windowSize   int64
	tickInterval time.Duration
	available    bool
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
			logs.Printf("cron exits by done")
			return
		case <-ctx.Done():
			logs.Printf("cron exits by context done")
			return
		case ts := <-c.t.C:
			watermark := c.w.Get()
			if watermark+c.windowSize < ts.Unix() {
				c.w.Advance()
				go asyncJoin(ctx, watermark, joinCallback, stateStore)
				logs.Printf("watermark advanced: %v to %v", watermark, watermark+c.windowSize)
			}
		}
	}
}

func asyncJoin(ctx context.Context, watermark int64, joinCallback models.JoinCallback, stateStore state.StateStore) {
	leftBatchIds := make([]string, 0)
	rightBatchIds := make([]string, 0)
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
			stateStore.Delete(ctx, batchId)
		}
		for _, batchId := range rightBatchIds {
			stateStore.Delete(ctx, batchId)
		}
	}()

	// fetch key sets
	leftKeySets := fetchKeySets(ctx, stateStore, leftBatchIds)
	rightKeySets := fetchKeySets(ctx, stateStore, rightBatchIds)

	// fetch records
	leftRecords := fetchRecords(ctx, stateStore, leftKeySets)
	rightRecords := fetchRecords(ctx, stateStore, rightKeySets)

	// join
	for _, leftRecord := range leftRecords {
		for _, rightRecord := range rightRecords {
			if err := joinCallback(ctx, leftRecord, rightRecord); err != nil {
				logs.Printf("async join: join callback failed: %v", err)
				return
			}
		}
	}
}

func fetchKeySets(ctx context.Context, stateStore state.StateStore, batchIds []string) []string {
	res := make([]string, 0)
	for _, batchId := range batchIds {
		keySet, err := stateStore.Get(ctx, batchId)
		if err == nil {
			keySetArr := make([]string, 0)
			json.Unmarshal(keySet, &keySetArr)
			res = append(res, keySetArr...)
		}
	}
	return res
}

func fetchRecords(ctx context.Context, stateStore state.StateStore, keySet []string) []*models.Record {
	res := make([]*models.Record, 0, len(keySet))
	for _, key := range keySet {
		val, err := stateStore.Get(ctx, key)
		if err == nil {
			res = append(res, models.NewRecord(key, val))
		}
	}
	return res
}
