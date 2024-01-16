package invokerlib

import "fmt"

type GlobalKafkaConfig struct {
	Address string `yaml:"address"`
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

type ProcessorConfig struct {
	// ParentDirectory is the parent directory that stores user-defined go files.
	ParentDirectory string `yaml:"parentDirectory"`

	// Files are the paths of user-defined go files, which includes fission handler, process function, init function.
	// If ParentDirectory is not empty, the file paths will be treate as paths relative to parent directory.
	Files []string `yaml:"files"`

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
	// DefaultOutputTopicPartitions defines the number of partitions of the default output topic of the processor.
	// If DefaultOutputTopicPartitions is set to 0, the default output topic of the processor will not be created.
	DefaultOutputTopicPartitions int      `yaml:"defaultOutputTopicPartitions"`
	OutputProcessors             []string `yaml:"outputProcessors"`

	// OutputKafkaConfigs defines non-processor destinations. The name of a self-defined OutputKafkaConfig
	// must not be the same as any of the processor names, and must be unique among all the OutputKafkaConfigs.
	OutputKafkaConfigs []*NamedKafkaConfig `yaml:"outputKafkaConfigs"`
}

type GlobalStoreConfig struct {
	// GlobaLStoreName is the name of a global store. The name of a global store should be unique among all global
	// stores in a config.
	GlobalStoreName string `yaml:"globalStoreName"`

	// GlobalStoreType is the type of a global store. Refer to GlobalStateStoreTypes in state_store.go for
	// possible store types.
	GlobalStoreType string `yaml:"globalStoreType"`

	// GlobalStoreSpec specifies further configuration of this global store.
	GlobalStoreSpec map[string]string `yaml:"globalStoreSpec"`
}

type RootConfig struct {
	// ProcessorConfigs defines all the processors in the stream processing pipeline. At least one processor config
	// must be provided.
	ProcessorConfigs []*ProcessorConfig `yaml:"processorConfigs"`

	// GlobalStoreConfigs defines all global stores used in the stream processing pipeline.
	GlobalStoreConfigs []*GlobalStoreConfig `yaml:"globalStoreConfigs"`

	// GlobalKafkaConfig specifies which Kafka cluster the interim topics should be created on.
	GlobalKafkaConfig *GlobalKafkaConfig `yaml:"globalKafkaConfig"`
}

func (rc *RootConfig) Validate() error {
	if len(rc.ProcessorConfigs) == 0 {
		return fmt.Errorf("no processor is specified in config")
	}
	processorNameSet := make(map[string]bool, 0)
	for _, pc := range rc.ProcessorConfigs {
		name := pc.Name
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
		if pc.OutputConfig.DefaultOutputTopicPartitions < 0 {
			return fmt.Errorf("DefaultOutputTopicPartitions must be greater than or equal to 0 for processor %s", name)
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
				return fmt.Errorf("OutputKafkaConfig name %s in processor %s is duplicated with processor names.",
					okc.Name, pc.Name)
			}
			if okcNameSet[okc.Name] {
				return fmt.Errorf("OutputKafkaConfig name %s in processor %s is duplicated with another OutputKafkaConfig name.",
					okc.Name, pc.Name)
			}
			okcNameSet[okc.Name] = true
		}
	}

	if rc.GlobalKafkaConfig == nil {
		return fmt.Errorf("global kafka config is missing")
	}

	return nil
}
