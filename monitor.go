package invokerlib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/IBM/sarama"
)

type FollowerMetadata struct {
	FunctionName        string
	FollowerKafkaConfig *KafkaConfig
	Status              string
}

var monitorCommands = struct {
	loadRootConfig string
	createTopics   string
}{
	loadRootConfig: "load_root_config",
	createTopics:   "create_topics",
}

var functionStatus = struct {
	Pending string
	Up      string
	Down    string
}{
	Pending: "Pending",
	Up:      "Up",
	Down:    "Down",
}

var (
	rootConfig       *RootConfig
	adminClient      sarama.ClusterAdmin
	followerMetadata map[string]*FollowerMetadata
)

func MonitorHandle(w http.ResponseWriter, r *http.Request) {
	if err := monitorHandle(w, r); err != nil {
		respBytes := marshalledErrorResponse(err)
		w.Write(respBytes)
	}
}

func monitorHandle(w http.ResponseWriter, r *http.Request) error {
	content, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("read request body failed: %v", err)
		logs.Printf("%v", err)
		return err
	}

	contentMap := make(map[string]any, 0)
	if err := json.Unmarshal(content, &contentMap); err != nil {
		err = fmt.Errorf("unmarshal content map failed: %v", err)
		logs.Printf("%v", err)
		return err
	}

	cmd, exists := contentMap[REQUEST_KEY_COMMAND]
	if !exists {
		err := fmt.Errorf("no command found in request body, body=%s", string(content))
		logs.Printf("%v", err)
		return err
	}
	paramsAny, paramExists := contentMap[REQUEST_KEY_PARAMS]
	var params map[string]any
	if paramExists {
		params = paramsAny.(map[string]any)
	}
	switch cmd {
	case monitorCommands.loadRootConfig:
		if err := loadRootConfig(paramExists, content, params); err != nil {
			return err
		}
	case monitorCommands.createTopics:
		if err := createTopics(); err != nil {
			return err
		}
	default:
		err := fmt.Errorf("unrecognized command %v", cmd)
		logs.Printf("%v", err)
		return err
	}

	return nil
}

func loadRootConfig(paramExists bool, content []byte, params map[string]any) error {
	logs.Printf("monitor load root config starts")
	if !paramExists {
		err := fmt.Errorf("no params is found for command %v, body=%s", monitorCommands.loadRootConfig, string(content))
		logs.Printf("%v", err)
		return err
	}

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		err := fmt.Errorf("marshal params failed: %v", err)
		logs.Printf("%v", err)
		return err
	}
	if err := json.Unmarshal(paramsBytes, rootConfig); err != nil {
		err := fmt.Errorf("unmarshal loadGlobalConfig params failed: err=%v, params=%s", err, string(paramsBytes))
		logs.Printf("%v", err)
		return err
	}

	// Validate function configs
	// TODO: check for topological ordering
	if len(rootConfig.FunctionConfigs) == 0 {
		err := fmt.Errorf("validate config failed: no function is specified in config")
		logs.Printf("%v", err)
		return err
	}
	functionNameSet := make(map[string]bool, 0)
	for _, functionConfig := range rootConfig.FunctionConfigs {
		if functionNameSet[functionConfig.FunctionName] {
			err := fmt.Errorf("validate config failed: duplicate function name %s", functionConfig.FunctionName)
			logs.Printf("%v", err)
			return err
		}
		functionNameSet[functionConfig.FunctionName] = true
		if functionConfig.SourceFunctionName == "" && functionConfig.SourceKakfaConfig == nil {
			err := fmt.Errorf("validate config failed: no source is specified for function %s", functionConfig.FunctionName)
			logs.Printf("%v", err)
			return err
		}
	}
	for _, functionConfig := range rootConfig.FunctionConfigs {
		srcName := functionConfig.SourceFunctionName
		if len(srcName) > 0 && !functionNameSet[srcName] {
			err := fmt.Errorf("validate config failed: source %s for function %s does not exist", srcName, functionConfig.FunctionName)
			logs.Printf("%v", err)
			return err
		}
	}

	// validate global kafka config
	if rootConfig.GlobalKafkaConfig == nil {
		err := fmt.Errorf("validate config failed: global kafka config is missing")
		logs.Printf("%v", err)
		return err
	}

	logs.Printf("monitor load root config starts")
	return nil
}

func createTopics() error {
	logs.Printf("monitor create topics starts")
	if rootConfig == nil {
		err := fmt.Errorf("create topics failed: root config is not loaded")
		logs.Printf("%v", err)
		return err
	}

	adminConf := sarama.NewConfig()
	adminConf.Version = sarama.V3_5_0_0

	var err error
	adminClient, err = sarama.NewClusterAdmin([]string{rootConfig.GlobalKafkaConfig.Address}, adminConf)
	if err != nil {
		err = fmt.Errorf("create admin client failed: %v", err)
		logs.Printf("%v", err)
		return err
	}

	// create topics
	for _, functionConfig := range rootConfig.FunctionConfigs {
		functionName := functionConfig.FunctionName

		// use function name as topic name
		topicName := functionName
		err := adminClient.CreateTopic(topicName, &sarama.TopicDetail{
			NumPartitions:     int32(functionConfig.NumOfPartition),
			ReplicationFactor: 0,
		}, false)
		if err != nil {
			err = fmt.Errorf("create topic failed: %v", err)
			logs.Printf("%v", err)
			if err := removeTopics(); err != nil {
				logs.Printf("remove topics failed: %v", err)
				return err
			} else {
				return err
			}
		}

		meta := &FollowerMetadata{
			FunctionName: functionName,
			Status:       functionStatus.Pending,
		}
		followerMetadata[functionName] = meta
	}

	logs.Printf("monitor create topics succeeds")
	return nil
}

func removeTopics() error {
	for _, functionMetadata := range followerMetadata {
		if err := adminClient.DeleteTopic(functionMetadata.FollowerKafkaConfig.Topic); err != nil {
			return err
		}
	}
	return nil
}
