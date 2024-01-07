package invokerlib

import "fmt"

type KafkaConfig struct {
	Address string `yaml:"address"`
	Topic   string `yaml:"topic"`
}

type UserFunctionConfig struct {
	// ParentDirectory is the parent directory where user-defined go files are located.
	ParentDirectory string `yaml:"parentDirectory"`

	// Files are the paths of user-defined go files, which contains fission handler, process, and init functions.
	// If ParentDirectory is not empty, the file paths will be interpreted as relative paths to parent directory.
	Files []string `yaml:"files"`

	// FunctionName is the name of the function. The name of a function should be unique among all functions
	// in a config.
	FunctionName string `yaml:"functionName"`

	// NumOfWorker defines the number of workers in function. Note that the maximum of number of parallel workers
	// in a function is capped by the number of partitions in that function's source Kafka topic.
	NumOfWorker int `yaml:"numOfWorker"`

	// NumOfPartition defines the number of partitions of the output topic for the function.
	// If NumOfPartition equals to 0, no output topic will be created for this function.
	NumOfPartition int `yaml:"numOfPartition"`

	// SourceFunction and SourceKafkaConfig defines the source of a function's input. Either SourceFunction
	// or SourceKafkaConfig must be not empty. If both are specified, SourceKafkaConfig will be used.
	SourceFunction    string       `yaml:"sourceFunction"`
	SourceKakfaConfig *KafkaConfig `yaml:"sourceKafkaConfig"`
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
	// FunctionConfigs defines all the functions in the stream processing pipeline. At least one function config
	// must present in root config.
	FunctionConfigs []*UserFunctionConfig `yaml:"functionConfigs"`

	// GlobalStoreConfigs defines all global stores used in the stream processing pipeline.
	GlobalStoreConfigs []*GlobalStoreConfig `yaml:"globalStoreConfigs"`

	// GlobalKafkaConfig specifies which cluster the interim topics should be created on.
	// Currently we assume all interim topics are stored on one single cluster
	GlobalKafkaConfig *KafkaConfig `yaml:"globalKafkaConfig"`
}

type FollowerConfig struct {
	FunctionName            string
	NumOfWorker             int
	KafkaSrc                *KafkaConfig
	FunctionNameToKafkaDest map[string]*KafkaConfig
}

// TODO: check for topological ordering
func (rc *RootConfig) Validate() error {
	if len(rc.FunctionConfigs) == 0 {
		return fmt.Errorf("no function is specified in config")
	}
	functionNameSet := make(map[string]bool, 0)
	for _, functionConfig := range rc.FunctionConfigs {
		if functionNameSet[functionConfig.FunctionName] {
			return fmt.Errorf("duplicate function name %s", functionConfig.FunctionName)
		}
		functionNameSet[functionConfig.FunctionName] = true
		if functionConfig.SourceFunction == "" && functionConfig.SourceKakfaConfig == nil {
			return fmt.Errorf("no source is specified for function %s", functionConfig.FunctionName)
		}
		if functionConfig.NumOfWorker <= 0 {
			return fmt.Errorf("NumOfWorker must be greater than 0 for function %s", functionConfig.FunctionName)
		}
		if functionConfig.NumOfPartition < 0 {
			return fmt.Errorf("NumOfPartition must be greater than or equal to 0 for function %s",
				functionConfig.FunctionName)
		}
	}
	for _, functionConfig := range rc.FunctionConfigs {
		srcName := functionConfig.SourceFunction
		if len(srcName) > 0 && !functionNameSet[srcName] {
			return fmt.Errorf("source %s for function %s does not exist", srcName, functionConfig.FunctionName)
		}
	}

	if rc.GlobalKafkaConfig == nil {
		return fmt.Errorf("global kafka config is missing")
	}

	return nil
}
