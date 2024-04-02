package core

import (
	"sync/atomic"
	"time"
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
