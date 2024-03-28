package consts

type ContextKey string

const (
	CTX_KEY_INVOKER_LIB_PROCESSOR_NAME = ContextKey("invoker_lib_processor_name")
	CTX_KEY_INVOKER_LIB_WORKER_ID      = ContextKey("invoker_lib_worker_id")
	CTX_KEY_INVOKER_LIB_WORKER_INDEX   = ContextKey("invoker_lib_worker_index")
)

const (
	REQUEST_KEY_COMMAND = "invoker_command"
	REQUEST_KEY_PARAMS  = "invoker_params"
)

const (
	MimeTypeJson              = "application/json"
	MimeTypeMultipartFormData = "multipart/form-data"
)

const CacheSize = 100 * 1024 * 1024

const MetricsNamespace = "invoker"

const DefaultCatLimit = 1000

const RedisPingRetryTimes = 3
