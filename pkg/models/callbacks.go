package models

import "context"

type InitCallback func() error
type ProcessCallback func(ctx context.Context, record *Record) error
type ExitCallback func()

type ProcessorCallbacks struct {
	OnInit  InitCallback
	Process ProcessCallback
	OnExit  ExitCallback
}
