package invokerlib

import (
	"context"
	"fmt"
	"sync"
)

type InMemoryMetricsClient struct {
	Counters   sync.Map
	Histograms sync.Map
}

func NewInMemoryMetricsClient() MetricsClient {
	return &InMemoryMetricsClient{
		Counters:   sync.Map{},
		Histograms: sync.Map{},
	}
}

func (m *InMemoryMetricsClient) EmitCounter(ctx context.Context, name string, delta uint64) error {
	key := PrependWorkerId(ctx, name)
	var newValue uint64 = 0
	val, exists := m.Counters.Load(key)
	if exists {
		newValue = val.(uint64) + delta
	}
	m.Counters.Store(key, newValue)
	return nil
}

func (m *InMemoryMetricsClient) GetCounter(ctx context.Context, name string) (uint64, error) {
	key := PrependWorkerId(ctx, name)
	if val, exists := m.Counters.Load(key); !exists {
		return 0, fmt.Errorf("counter with name %s does not exist", name)
	} else {
		return val.(uint64), nil
	}
}

func (m *InMemoryMetricsClient) EmitHistogram(ctx context.Context, name string, val float64) error {
	key := PrependWorkerId(ctx, name)
	if histogram, exists := m.Histograms.Load(key); !exists {
		m.Histograms.Store(key, NewHistogram(name, val))
	} else {
		histogram.(*Histogram).addValue(val)
	}
	return nil
}
