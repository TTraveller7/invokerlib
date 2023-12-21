package invokerlib

type Config struct {
	// TOOD: this shld be in global config
	KafkaSrc   *KafkaConfig
	KafkaDests map[string]*KafkaConfig

	NumOfWorker int

	FunctionName string

	RedisConf *RedisStateStoreConfig
}

type RedisStateStoreConfig struct {
	IsGlobal bool
	Address  string
}

type KafkaConfig struct {
	Address string
	Topic   string
}
