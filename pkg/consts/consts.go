package consts

type contextKey string

const (
	CTX_KEY_INVOKER_LIB_PROCESSOR_NAME = contextKey("invoker_lib_processor_name")
	CTX_KEY_INVOKER_LIB_WORKER_ID      = contextKey("invoker_lib_worker_id")
	CTX_KEY_INVOKER_LIB_WORKER_INDEX   = contextKey("invoker_lib_worker_index")
	CTX_KEY_INVOKER_LIB_WORKER_TOPIC   = contextKey("invoker_lib_worker_topic")
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

const (
	ProcessorTypeProcess = "process"
	ProcessorTypeJoin    = "join"
)

const JoinKeyBufferMinCapacity = 8

const JoinMinWindowSize = 10
