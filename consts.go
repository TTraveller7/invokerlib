package invokerlib

import "fmt"

type ContextKey string

const (
	CTX_KEY_INVOKER_LIB_PROCESSOR_NAME = ContextKey("invoker_lib_processor_name")
	CTX_KEY_INVOKER_LIB_WORKER_ID      = ContextKey("invoker_lib_worker_id")
	CTX_KEY_INVOKER_LIB_WORKER_INDEX   = ContextKey("invoker_lib_worker_index")
)

var errConsumerNotify error = fmt.Errorf("consumer exits with notify")

const (
	REQUEST_KEY_COMMAND = "invoker_command"
	REQUEST_KEY_PARAMS  = "invoker_params"
)
