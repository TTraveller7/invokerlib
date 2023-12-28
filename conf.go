package invokerlib

type KafkaConfig struct {
	Address string
	Topic   string
}

type UserConfig struct {
	FunctionConfigs    []*UserFunctionConfig
	GlobalStoreConfigs []*GlobalStoreConfig
}

type UserFunctionConfig struct {
	// Files are the paths of user-defined go files, which contains fission handler, process, and init functions.
	Files []string

	// FunctionName is the name of the function. The name of a function should be unique among all functions
	// in a config.
	FunctionName string

	// NumOfWorker defines the number of workers in function. Note that the maximum of number of parallel workers
	// in a function is capped by the number of partitions in that function's source Kafka topic.
	NumOfWorker int

	// SourceFunctionName and SourceKafkaConfig defines the source of a function's input. Either SourceFunctionName
	// or SourceKafkaConfig must be not empty. A function cannot have both non-empty SourceFunctionName
	// and non-empty SourceKafkaConfig.
	SourceFunctionName string
	SourceKakfaConfig  *KafkaConfig

	// DestFunctioNames are the destination functions of a function's output. A function can have zero or more
	// destination function.
	DestFunctionNames []string

	// DestKAfkaConfigs are the destination Kafka topics of a function's output. A function can have zero or more
	// destination Kafka topic.
	DestKafkaConfigs []*KafkaConfig
}

type GlobalStoreConfig struct {
	// GlobaLStoreName is the name of a global store. The name of a global store should be unique among all global
	// stores in a config.
	GlobalStoreName string

	// GlobalStoreType is the type of a global store. Refer to GlobalStateStoreTypes in state_store.go for
	// possible store types.
	GlobalStoreType string
}

type FollowerFunctionConfig struct {
	// TODO: this shld be in global config
	FunctionName            string
	NumOfWorker             int
	KafkaSrc                *KafkaConfig
	FunctionNameToKafkaDest map[string]*KafkaConfig
}
