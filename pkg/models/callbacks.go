package models

import "context"

type InitCallback func() error
type ProcessCallback func(ctx context.Context, record *Record) error
type JoinCallback func(ctx context.Context, leftRecord *Record, rightRecord *Record) error
type ExitCallback func()

type ProcessorCallbacks struct {
	OnInit  InitCallback
	Process ProcessCallback
	Join    JoinCallback
	OnExit  ExitCallback
}
