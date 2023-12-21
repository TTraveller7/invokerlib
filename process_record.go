package invokerlib

import "context"

type ProcessFunc func(ctx context.Context, record *Record)
