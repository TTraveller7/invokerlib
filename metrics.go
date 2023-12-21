package invokerlib

import "context"

type MetricsClient interface {
	EmitCounter(ctx context.Context, name string, val uint64) error
}
