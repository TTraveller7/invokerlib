package core

import (
	"context"
	"time"

	"github.com/TTraveller7/invokerlib/pkg/models"
	"github.com/TTraveller7/invokerlib/pkg/state"
	"github.com/TTraveller7/invokerlib/pkg/utils"
	"github.com/bytedance/sonic"
)

type JoinWorker struct {
	w          *Watermark
	s          state.StateStore
	expireTime int
}

func NewJoinWorker(w *Watermark, s state.StateStore, expierTime int) *JoinWorker {
	return &JoinWorker{
		w:          w,
		s:          s,
		expireTime: expierTime,
	}
}

func (j *JoinWorker) JoinWorkerProcessCallback(ctx context.Context, record *models.Record) error {
	processingTimestamp := time.Now().Unix()
	watermark := j.w.Get()
	if processingTimestamp < watermark {
		// drop record
		return nil
	}
	batchId := utils.BatchId(ctx, watermark)
	keySet, err := j.s.Get(ctx, batchId)
	keys := make([]string, 0)
	if err == nil {
		sonic.Unmarshal(keySet, &keys)
	}
	keys = append(keys, record.Key())
	if err := j.s.PutWithExpireTime(ctx, record.Key(), record.Value(), j.expireTime); err != nil {
		return err
	}
	keySet, _ = sonic.Marshal(keys)
	if err := j.s.Put(ctx, batchId, keySet); err != nil {
		return err
	}
	return nil
}
