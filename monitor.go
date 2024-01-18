package invokerlib

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/IBM/sarama"
)

type ProcessorMetadata struct {
	Name   string
	Status string
	Client *ProcessorClient
}

var (
	rootConfig        *RootConfig = &RootConfig{}
	adminClient       sarama.ClusterAdmin
	interimTopics     []*KafkaConfig                = make([]*KafkaConfig, 0)
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
	case MonitorCommands.LoadProcessorEndpoints:
		return loadProcessorEndpoints(req)
	case MonitorCommands.InitializeProcessors:
		return initializeProcessors()
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

	logs.Printf("monitor load root config finished")
	return successResponse(), nil
}

func createTopics() (*InvokerResponse, error) {
	logs.Printf("monitor create topics starts")
	if rootConfig == nil {
		err := fmt.Errorf("create topics failed: root config is not loaded")
		logs.Printf("%v", err)
		return nil, err
	}

	kafkaAddr := rootConfig.GlobalKafkaConfig.Address
	adminConf := sarama.NewConfig()
	adminConf.Version = sarama.V3_5_0_0

	var err error
	adminClient, err = sarama.NewClusterAdmin([]string{kafkaAddr}, adminConf)
	if err != nil {
		err = fmt.Errorf("create admin client failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	// create topics
	for _, pc := range rootConfig.ProcessorConfigs {
		numOfPartitions := pc.OutputConfig.DefaultTopicPartitions
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
			Name:   processorName,
			Status: processorStatus.Pending,
		}
		processorMetadata[processorName] = meta

		interimTopics = append(interimTopics, &KafkaConfig{
			Address: kafkaAddr,
			Topic:   topicName,
		})
	}

	logs.Printf("monitor create topics succeeds")
	return successResponse(), nil
}

func removeTopics() (*InvokerResponse, error) {
	for _, kc := range interimTopics {
		if err := adminClient.DeleteTopic(kc.Topic); err != nil {
			return nil, err
		}
	}
	return successResponse(), nil
}

func loadProcessorEndpoints(req *InvokerRequest) (*InvokerResponse, error) {
	logs.Printf("monitor load processor endpoints starts")
	if req.Params == nil {
		err := fmt.Errorf("no params is found for command %v", MonitorCommands.LoadProcessorEndpoints)
		logs.Printf("%v", err)
		return nil, err
	}

	p := &LoadProcessorEndpointsParams{}
	if err := UnmarshalParams(req.Params, p); err != nil {
		err := fmt.Errorf("unmarshal loadProcessorEndpoints params failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	for name, metadata := range processorMetadata {
		for _, endpoint := range p.Endpoints {
			if strings.Contains(endpoint, name) {
				metadata.Client = NewProcessorClient(name, endpoint)
				break
			}
		}
	}

	for name, metadata := range processorMetadata {
		if metadata.Client == nil {
			err := fmt.Errorf("processor %s endpoint not found in request", name)
			logs.Printf("%v", err)
			return nil, err
		}
	}

	logs.Printf("monitor load processor endpoints finished")
	return successResponse(), nil
}

func initializeProcessors() (*InvokerResponse, error) {
	logs.Printf("monitor initialize processors starts")
	for _, metadata := range processorMetadata {
		if metadata.Client == nil {
			err := fmt.Errorf("processor %s client is not initialized", metadata.Name)
			logs.Printf("%v", err)
			return nil, err
		}
		resp, err := metadata.Client.Initialize()
		if err != nil {
			err = fmt.Errorf("processor initialize failed: %v", err)
			logs.Printf("%v", err)
			return nil, err
		} else if resp.Code != ResponseCodes.Success {
			err = fmt.Errorf("processor initialize failed with resp: %+v", resp)
			logs.Printf("%v", err)
			return nil, err
		}
		logs.Printf("processor initialize finished with resp: %+v", resp)
	}

	logs.Printf("monitor initialize processors finished")
	return successResponse(), nil
}
