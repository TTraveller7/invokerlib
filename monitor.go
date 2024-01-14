package invokerlib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"

	"github.com/IBM/sarama"
)

type FollowerMetadata struct {
	FunctionName        string
	FollowerKafkaConfig *KafkaConfig
	Status              string
}

var MonitorCommands = struct {
	LoadRootConfig string
	CreateTopics   string
}{
	LoadRootConfig: "loadRootConfig",
	CreateTopics:   "createTopics",
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
	var err error
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = fmt.Errorf("%v. %s", panicErr, string(debug.Stack()))
		}
		if err != nil {
			respBytes := marshalledErrorResponse(err)
			w.Write(respBytes)
		}
	}()
	err = monitorHandle(w, r)
}

func monitorHandle(w http.ResponseWriter, r *http.Request) error {
	content, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("read request body failed: %v", err)
		logs.Printf("%v", err)
		return err
	}

	req := &InvokerRequest{}
	if err := json.Unmarshal(content, req); err != nil {
		err = fmt.Errorf("unmarshal request failed: %v", err)
		logs.Printf("%v", err)
		return err
	}
	if req.Command == "" {
		err = fmt.Errorf("request command is missing")
		logs.Printf("%v", err)
		return err
	}

	switch req.Command {
	case MonitorCommands.LoadRootConfig:
		if err := loadRootConfig(req); err != nil {
			return err
		}
	case MonitorCommands.CreateTopics:
		if err := createTopics(); err != nil {
			return err
		}
	default:
		err := fmt.Errorf("unrecognized command %v", req.Command)
		logs.Printf("%v", err)
		return err
	}

	return nil
}

func loadRootConfig(req *InvokerRequest) error {
	logs.Printf("monitor load root config starts")
	if req.Params == nil {
		err := fmt.Errorf("no params is found for command %v", MonitorCommands.LoadRootConfig)
		logs.Printf("%v", err)
		return err
	}

	if err := UnmarshalParams(req.Params, rootConfig); err != nil {
		err := fmt.Errorf("unmarshal loadGlobalConfig params failed: %v", err)
		logs.Printf("%v", err)
		return err
	}

	if err := rootConfig.Validate(); err != nil {
		logs.Printf("validate root config failed: %v", err)
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
		if functionConfig.NumOfPartition == 0 {
			// skip output topic creation of this function
			continue
		}

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
