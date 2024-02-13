package invokerlib

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// What metrics should I have in the whole process?
// - Number of records processed
// Let's keep that for now.
// I should wrap prometheus API for both me and client
// at first let's keep this to internal use
//
// For now we only support processor level metrics

type MetricsClient struct {
	processerName string
	counters      sync.Map
}

func NewMetricsClient() *MetricsClient {
	return &MetricsClient{
		processerName: conf.Name,
		counters:      sync.Map{},
	}
}

func (m *MetricsClient) EmitCounter(name string, help string, val float64) error {
	if _, exists := m.counters.Load(name); !exists {
		c := prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: MetricsNamespace,
			Subsystem: m.processerName,
			Name:      name,
			Help:      help,
		})
		if err := prometheus.DefaultRegisterer.Register(c); err != nil {
			return err
		}
		m.counters.Store(name, c)
	}
	c, _ := m.counters.Load(name)
	c.(prometheus.Counter).Add(val)
	return nil
}

func metricsHandler() http.Handler {
	return promhttp.Handler()
}
