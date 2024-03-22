package invokerlib

import "fmt"

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

	// NumOfWorker defines the number of workers in processor. Note that the number of parallel workers
	// in a processor is capped by the number of partitions in that processor's source Kafka topic.
	NumOfWorker int `yaml:"numOfWorker"`

	// InputProcessor and InputKafkaConfig defines the source of a processor's input. Either InputProcessor
	// or InputKafkaConfig must be not empty. If both are specified, InputKafkaConfig will be used.
	InputProcessor   string       `yaml:"inputProcessor"`
	InputKafkaConfig *KafkaConfig `yaml:"inputKafkaConfig"`

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

type GlobalStoreConfig struct {
	RedisConfigs []*RedisConfig `yaml:"redisConfigs"`
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
		if pc.Name == "" {
			return fmt.Errorf("processor name cannot be empty")
		}
		name := pc.Name
		if pc.EntryPoint == "" {
			return fmt.Errorf("processor %s entrypoint cannot be empty", name)
		}
		if processorNameSet[name] {
			return fmt.Errorf("duplicate processor name %s", name)
		}
		processorNameSet[name] = true
		if pc.InputProcessor == "" && pc.InputKafkaConfig == nil {
			return fmt.Errorf("no source is specified for processor %s", name)
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
	for _, pc := range rc.ProcessorConfigs {
		srcName := pc.InputProcessor
		if len(srcName) > 0 && !processorNameSet[srcName] {
			return fmt.Errorf("InputFunction %s for processor %s does not exist", srcName, pc.Name)
		}
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
		for _, rc := range rc.GlobalStoreConfig.RedisConfigs {
			if rc.Name == "" {
				return fmt.Errorf("redis config name cannot be empty")
			}
			if redisNames[rc.Name] {
				return fmt.Errorf("redis config name should not duplicate")
			}
			redisNames[rc.Name] = true
			if rc.Address == "" {
				return fmt.Errorf("redis config address cannot be empty")
			}
		}
	}

	return nil
}

type InternalProcessorConfig struct {
	Name                     string                  `json:"name"`
	GlobalKafkaConfig        *GlobalKafkaConfig      `json:"global_kafka_config"`
	NumOfWorker              int                     `json:"num_of_worker"`
	InputKafkaConfig         *KafkaConfig            `json:"input_kafka_config"`
	DefaultOutputKafkaConfig *KafkaConfig            `json:"default_output_kafka_config"`
	OutputKafkaConfigs       map[string]*KafkaConfig `json:"output_kafka_configs"`
	GlobalStoreConfig        *GlobalStoreConfig      `json:"global_store_config"`
}

func (ipc *InternalProcessorConfig) Validate() error {
	if ipc.Name == "" {
		return fmt.Errorf("processor name should not be empty")
	}
	if ipc.GlobalKafkaConfig == nil {
		return fmt.Errorf("global kafka config should be nil")
	}
	if ipc.NumOfWorker == 0 {
		return fmt.Errorf("number of worker should not be 0")
	}
	if ipc.InputKafkaConfig == nil {
		return fmt.Errorf("input kafka config should not be nil")
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
