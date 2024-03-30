package conf

import (
	"fmt"

	"github.com/TTraveller7/invokerlib/pkg/consts"
)

type InitialTopic struct {
	Topic      string `yaml:"topic"`
	Partitions int    `yaml:"partitions"`
}

type GlobalKafkaConfig struct {
	Address       string          `yaml:"address"`
	InitialTopics []*InitialTopic `yaml:"initialTopics"`
}

type KafkaConfig struct {
	Address string `yaml:"address"`
	Topic   string `yaml:"topic"`
}

type NamedKafkaConfig struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
	Topic   string `yaml:"topic"`
}

// Assumptions:
// 1. Each processor can read from one input topic.
// 2. Each processor can write to one or more output topic.
// 3. All interim topics are stored on one single Kafka cluster.

// Proposals:
// 1. Topological ordering?
// 2. Monitor healthchecks
// 3. Local state store health check: We should make sure that the local state store always has enough size
// so that keys are not evicted automatically
// 4. upload file
// 5. state store commit log and check pointing
// 6. automatic # worker and partition adjustment
// 7. cat needs a limit - works like head in this way

type ProcessorConfig struct {
	// ParentDirectory is the parent directory that stores user-defined go files.
	ParentDirectory string `yaml:"parentDirectory"`

	// Files are the paths of user-defined go files, which includes fission handler, process function, init function.
	// If ParentDirectory is not empty, the file paths will be treate as paths relative to parent directory.
	Files []string `yaml:"files"`

	// EntryPoint is the entry point function of the processor.
	EntryPoint string `yaml:"entryPoint"`

	// Name is the name of the processor, which should be unique among all processors in a config.
	Name string `yaml:"name"`

	Type string `yaml:"type"`

	// NumOfWorker defines the number of workers in processor. Note that the number of parallel workers
	// in a processor is capped by the number of partitions in that processor's source Kafka topic.
	NumOfWorker int `yaml:"numOfWorker"`

	// InputProcessors and InputKafkaConfigs define the source of a processor's input. Either InputProcessor
	// or InputKafkaConfig must be not empty. If both are specified, InputKafkaConfig will be used.
	InputProcessors   []string       `yaml:"inputProcessors"`
	InputKafkaConfigs []*KafkaConfig `yaml:"inputKafkaConfigs"`

	// OutputConfig defines the output of a processor's input. OutputConfig must be specified in a config.
	OutputConfig *OutputConfig `yaml:"outputConfig"`
}

type OutputConfig struct {
	// DefaultTopicPartitions defines the number of partitions of the default output topic of the processor.
	// If DefaultTopicPartitions is set to 0, the default topic of the processor will not be created.
	DefaultTopicPartitions int `yaml:"defaultTopicPartitions"`

	OutputProcessors []string `yaml:"outputProcessors"`

	// OutputKafkaConfigs defines non-processor destinations. The name of a self-defined OutputKafkaConfig
	// must not be the same as any of the processor names, and must be unique among all the OutputKafkaConfigs.
	OutputKafkaConfigs []*NamedKafkaConfig `yaml:"outputKafkaConfigs"`
}

type RedisConfig struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
}

type MemcachedConfig struct {
	Name      string   `yaml:"name"`
	Addresses []string `yaml:"addresses"`
}

type GlobalStoreConfig struct {
	RedisConfigs     []*RedisConfig     `yaml:"redisConfigs"`
	MemcachedConfigs []*MemcachedConfig `yaml:"memcachedConfigs"`
}

type RootConfig struct {
	// ProcessorConfigs defines all the processors in the stream processing pipeline. At least one processor config
	// must be provided.
	ProcessorConfigs []*ProcessorConfig `yaml:"processorConfigs"`

	// GlobalStoreConfig defines all global stores used in the stream processing pipeline.
	GlobalStoreConfig *GlobalStoreConfig `yaml:"globalStoreConfig"`

	// GlobalKafkaConfig specifies which Kafka cluster the interim topics should be created on.
	GlobalKafkaConfig *GlobalKafkaConfig `yaml:"globalKafkaConfig"`
}

func (rc *RootConfig) Validate() error {
	if len(rc.ProcessorConfigs) == 0 {
		return fmt.Errorf("no processor is specified in config")
	}
	processorNameSet := make(map[string]bool, 0)
	for _, pc := range rc.ProcessorConfigs {
		// check name
		if pc.Name == "" {
			return fmt.Errorf("processor name cannot be empty")
		}
		name := pc.Name
		if processorNameSet[name] {
			return fmt.Errorf("duplicate processor name %s", name)
		}
		processorNameSet[name] = true

		// check entrypoint
		if pc.EntryPoint == "" {
			return fmt.Errorf("processor %s entrypoint cannot be empty", name)
		}

		// check type and input config based on type
		inputCount := len(pc.InputProcessors) + len(pc.InputKafkaConfigs)
		switch pc.Type {
		case consts.ProcessorTypeProcess:
			if inputCount != 1 {
				return fmt.Errorf("processor with type=process must have one and only one input source")
			}
		case consts.ProcessorTypeJoin:
			if inputCount <= 1 {
				return fmt.Errorf("processor with type=join must have more than one input sources")
			}
		default:
			return consts.ErrProcessorTypeNotRecognized
		}

		if pc.NumOfWorker <= 0 {
			return fmt.Errorf("NumOfWorker must be greater than 0 for processor %s", name)
		}
		if pc.OutputConfig == nil {
			return fmt.Errorf("output config is not specified for processor %s", name)
		}
		if pc.OutputConfig.DefaultTopicPartitions < 0 {
			return fmt.Errorf("DefaultTopicPartitions must be greater than or equal to 0 for processor %s", name)
		}
	}

	// check input processor name
	for _, pc := range rc.ProcessorConfigs {
		for _, n := range pc.InputProcessors {
			if len(n) == 0 {
				return fmt.Errorf("input processor name cannot be empty")
			}
			if !processorNameSet[n] {
				return fmt.Errorf("input processor %s does not exist", n)
			}
		}
	}

	// check input kafka config
	for _, pc := range rc.ProcessorConfigs {
		for _, kc := range pc.InputKafkaConfigs {
			if kc.Address == "" {
				return consts.ErrKakfaAddressEmpty("inputKafkaConfigs")
			}
			if kc.Topic == "" {
				return consts.ErrKafkaTopicEmpty("inputKafkaConfigs")
			}
		}
	}

	// check output processor config
	for _, pc := range rc.ProcessorConfigs {
		okcNameSet := make(map[string]bool, 0)
		for _, okc := range pc.OutputConfig.OutputKafkaConfigs {
			if processorNameSet[okc.Name] {
				return fmt.Errorf("OutputKafkaConfig name %s in processor %s is duplicated with processor names",
					okc.Name, pc.Name)
			}
			if okcNameSet[okc.Name] {
				return fmt.Errorf("OutputKafkaConfig name %s in processor %s is duplicated with another OutputKafkaConfig name",
					okc.Name, pc.Name)
			}
			okcNameSet[okc.Name] = true
		}
	}

	if rc.GlobalKafkaConfig == nil {
		return fmt.Errorf("global kafka config is missing")
	}
	for _, it := range rc.GlobalKafkaConfig.InitialTopics {
		if it.Topic == "" {
			return fmt.Errorf("initial topic name must not be empty")
		}
		if it.Partitions == 0 {
			return fmt.Errorf("the partition field of initial topic %s must be greater than 0", it.Topic)
		}
	}

	if rc.GlobalStoreConfig != nil {
		redisNames := make(map[string]bool, 0)
		for _, redisConfig := range rc.GlobalStoreConfig.RedisConfigs {
			if redisConfig.Name == "" {
				return fmt.Errorf("redis config name cannot be empty")
			}
			if redisNames[redisConfig.Name] {
				return fmt.Errorf("redis config name should not duplicate")
			}
			redisNames[redisConfig.Name] = true
			if redisConfig.Address == "" {
				return fmt.Errorf("redis config address cannot be empty")
			}
		}
		memcachedNames := make(map[string]bool, 0)
		for _, mc := range rc.GlobalStoreConfig.MemcachedConfigs {
			if mc.Name == "" {
				return fmt.Errorf("memcached config name cannot be empty")
			}
			if memcachedNames[mc.Name] {
				return fmt.Errorf("memcached config name should not duplicate")
			}
			memcachedNames[mc.Name] = true
			if len(mc.Addresses) == 0 {
				return fmt.Errorf("memcached config address cannot be empty")
			}
		}
	}

	return nil
}

type ConsumerConfig struct {
	Address      string `json:"address"`
	Topic        string `json:"topic"`
	NumOfWorkers int    `json:"num_of_workers"`
}

type InternalProcessorConfig struct {
	Name                     string                  `json:"name"`
	Type                     string                  `json:"type"`
	GlobalKafkaConfig        *GlobalKafkaConfig      `json:"global_kafka_config"`
	ConsumerConfigs          []*ConsumerConfig       `json:"consumer_configs"`
	DefaultOutputKafkaConfig *KafkaConfig            `json:"default_output_kafka_config"`
	OutputKafkaConfigs       map[string]*KafkaConfig `json:"output_kafka_configs"`
	GlobalStoreConfig        *GlobalStoreConfig      `json:"global_store_config"`
}

func NewInternalProcessorConfig(rootConfig *RootConfig, processorName string) *InternalProcessorConfig {
	ipc := &InternalProcessorConfig{
		Name:              processorName,
		GlobalKafkaConfig: rootConfig.GlobalKafkaConfig,
		GlobalStoreConfig: rootConfig.GlobalStoreConfig,
	}

	kafkaAddr := rootConfig.GlobalKafkaConfig.Address
	for _, processorConfig := range rootConfig.ProcessorConfigs {
		if processorConfig.Name != processorName {
			continue
		}

		consumerConfigs := make([]*ConsumerConfig, 0)
		for _, inputProcessor := range processorConfig.InputProcessors {
			cc := &ConsumerConfig{
				Address:      kafkaAddr,
				Topic:        inputProcessor,
				NumOfWorkers: processorConfig.NumOfWorker,
			}
			consumerConfigs = append(consumerConfigs, cc)
		}
		for _, inputKakfaConfig := range processorConfig.InputKafkaConfigs {
			cc := &ConsumerConfig{
				Address:      inputKakfaConfig.Address,
				Topic:        inputKakfaConfig.Topic,
				NumOfWorkers: processorConfig.NumOfWorker,
			}
			consumerConfigs = append(consumerConfigs, cc)
		}
		ipc.ConsumerConfigs = consumerConfigs

		if processorConfig.OutputConfig.DefaultTopicPartitions > 0 {
			ipc.DefaultOutputKafkaConfig = &KafkaConfig{
				Address: kafkaAddr,
				Topic:   processorConfig.Name,
			}
		}

		outputMap := make(map[string]*KafkaConfig, 0)
		for _, outputProcessor := range processorConfig.OutputConfig.OutputProcessors {
			outputMap[outputProcessor] = &KafkaConfig{
				Address: kafkaAddr,
				Topic:   outputProcessor,
			}
		}
		for _, outputKafkaConfig := range processorConfig.OutputConfig.OutputKafkaConfigs {
			key := outputKafkaConfig.Name
			val := &KafkaConfig{
				Address: outputKafkaConfig.Address,
				Topic:   outputKafkaConfig.Topic,
			}
			outputMap[key] = val
		}
		ipc.OutputKafkaConfigs = outputMap
	}
	return ipc
}

func (ipc *InternalProcessorConfig) Validate() error {
	if ipc.Name == "" {
		return fmt.Errorf("processor name should not be empty")
	}
	if ipc.Type == "" {
		return fmt.Errorf("processor type should not be empty")
	}
	if ipc.GlobalKafkaConfig == nil {
		return fmt.Errorf("global kafka config should be nil")
	}
	if len(ipc.ConsumerConfigs) == 0 {
		return fmt.Errorf("consumer configs should not be empty")
	}
	for _, cc := range ipc.ConsumerConfigs {
		if cc.Address == "" {
			return fmt.Errorf("consumer config address should not be empty")
		}
		if cc.Topic == "" {
			return fmt.Errorf("consumer config topic should not be empty")
		}
		if cc.NumOfWorkers <= 0 {
			return fmt.Errorf("consumer numOfWorker must be positive")
		}
	}
	if ipc.OutputKafkaConfigs == nil {
		ipc.OutputKafkaConfigs = make(map[string]*KafkaConfig, 0)
	}
	return nil
}

type InternalKafkaConfig struct {
	Address    string
	Topic      string
	Partitions int
}

var (
	c                *InternalProcessorConfig
	redisConfigs     map[string]*RedisConfig     = make(map[string]*RedisConfig, 0)
	memcachedConfigs map[string]*MemcachedConfig = make(map[string]*MemcachedConfig, 0)
)

func Config() *InternalProcessorConfig {
	if c == nil {
		panic(consts.ErrConfigNotInitialized)
	}
	return c
}

func GetRedisConfigByName(name string) *RedisConfig {
	return redisConfigs[name]
}

func GetMemcachedConfigByName(name string) *MemcachedConfig {
	return memcachedConfigs[name]
}

func LoadConfig(config *InternalProcessorConfig) error {
	if err := config.Validate(); err != nil {
		err = fmt.Errorf("validate internal processor config failed: %v", err)
		return err
	}
	c = config
	if c.GlobalStoreConfig != nil {
		for _, rc := range c.GlobalStoreConfig.RedisConfigs {
			redisConfigs[rc.Name] = rc
		}
		for _, mc := range c.GlobalStoreConfig.MemcachedConfigs {
			memcachedConfigs[mc.Name] = mc
		}
	}
	return nil
}
