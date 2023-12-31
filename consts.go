package invokerlib

import "fmt"

const (
	CTX_KEY_INVOKER_LIB_FUNCTION_NAME = "invoker_lib_function_name"
	CTX_KEY_INVOKER_LIB_WORKER_ID     = "invoker_lib_worker_id"
	CTX_KEY_INVOKER_LIB_WORKER_INDEX  = "invoker_lib_worker_index"
)

var consumerNotify = fmt.Errorf("consumer exits with notify")

const (
	REQUEST_KEY_COMMAND = "invoker_command"
	REQUEST_KEY_PARAMS  = "invoker_params"
)
