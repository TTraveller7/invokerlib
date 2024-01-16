package invokerlib

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/IBM/sarama"
)

type ProcessorMetadata struct {
	ProcessorName              string
	ProcessorOutputKafkaConfig *KafkaConfig
	Status                     string
}

var MonitorCommands = struct {
	LoadRootConfig string
	CreateTopics   string
}{
	LoadRootConfig: "loadRootConfig",
	CreateTopics:   "createTopics",
}

var processorStatus = struct {
	Pending string
	Up      string
	Down    string
}{
	Pending: "Pending",
	Up:      "Up",
	Down:    "Down",
}

var (
	rootConfig        *RootConfig = &RootConfig{}
	adminClient       sarama.ClusterAdmin
	processorMetadata map[string]*ProcessorMetadata = make(map[string]*ProcessorMetadata, 0)
)

func MonitorHandle(w http.ResponseWriter, r *http.Request) {
	if logs == nil {
		logs = log.New(os.Stdout, "[monitor] ", log.LstdFlags|log.Lshortfile)
	}

	resp := &InvokerResponse{}
	var err error
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = fmt.Errorf("%v. %s", panicErr, string(debug.Stack()))
		}
		if err != nil {
			resp = failureResponse(err)
		}
		respBytes, _ := json.Marshal(resp)
		w.Write(respBytes)
	}()

	content, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("read request body failed: %v", err)
		logs.Printf("%v", err)
		return
	}

	req := &InvokerRequest{}
	if err := json.Unmarshal(content, req); err != nil {
		err = fmt.Errorf("unmarshal request failed: %v", err)
		logs.Printf("%v", err)
		return
	}
	if req.Command == "" {
		err = fmt.Errorf("request command is missing")
		logs.Printf("%v", err)
		return
	}

	resp, err = monitorHandle(req)
	if err != nil {
		err = fmt.Errorf("handle monitor command failed: %v", err)
		logs.Printf("%v", err)
	}
}

func monitorHandle(req *InvokerRequest) (*InvokerResponse, error) {
	switch req.Command {
	case MonitorCommands.LoadRootConfig:
		return loadRootConfig(req)
	case MonitorCommands.CreateTopics:
		return createTopics()
	default:
		err := fmt.Errorf("unrecognized command %v", req.Command)
		logs.Printf("%v", err)
		return nil, err
	}
}

func loadRootConfig(req *InvokerRequest) (*InvokerResponse, error) {
	logs.Printf("monitor load root config starts")
	if req.Params == nil {
		err := fmt.Errorf("no params is found for command %v", MonitorCommands.LoadRootConfig)
		logs.Printf("%v", err)
		return nil, err
	}

	if err := UnmarshalParams(req.Params, rootConfig); err != nil {
		err := fmt.Errorf("unmarshal loadGlobalConfig params failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	if err := rootConfig.Validate(); err != nil {
		logs.Printf("validate root config failed: %v", err)
		return nil, err
	}

	logs.Printf("monitor load root config finshed")
	return successResponse(), nil
}

func createTopics() (*InvokerResponse, error) {
	logs.Printf("monitor create topics starts")
	if rootConfig == nil {
		err := fmt.Errorf("create topics failed: root config is not loaded")
		logs.Printf("%v", err)
		return nil, err
	}

	adminConf := sarama.NewConfig()
	adminConf.Version = sarama.V3_5_0_0

	var err error
	adminClient, err = sarama.NewClusterAdmin([]string{rootConfig.GlobalKafkaConfig.Address}, adminConf)
	if err != nil {
		err = fmt.Errorf("create admin client failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	// create topics
	for _, pc := range rootConfig.ProcessorConfigs {
		numOfPartitions := pc.OutputConfig.DefaultOutputTopicPartitions
		if numOfPartitions == 0 {
			// skip output topic creation of this processor
			continue
		}

		processorName := pc.Name

		// use processor name as topic name
		topicName := processorName
		err := adminClient.CreateTopic(topicName, &sarama.TopicDetail{
			NumPartitions:     int32(numOfPartitions),
			ReplicationFactor: 1,
		}, false)
		if err != nil {
			err = fmt.Errorf("create topic failed: %v", err)
			logs.Printf("%v", err)
			if _, err := removeTopics(); err != nil {
				logs.Printf("remove topics failed: %v", err)
				return nil, err
			} else {
				return nil, err
			}
		}

		meta := &ProcessorMetadata{
			ProcessorName: processorName,
			Status:        processorStatus.Pending,
		}
		processorMetadata[processorName] = meta
	}

	logs.Printf("monitor create topics succeeds")
	return successResponse(), nil
}

func removeTopics() (*InvokerResponse, error) {
	for _, pmt := range processorMetadata {
		if err := adminClient.DeleteTopic(pmt.ProcessorOutputKafkaConfig.Topic); err != nil {
			return nil, err
		}
	}
	return successResponse(), nil
}
