package invokerlib

import (
	"fmt"
	"net/http"
)

type ProcessorClient struct {
	InvokerClient

	ProcessorName string
}

func NewProcessorClient(processorName, url string) *ProcessorClient {
	return &ProcessorClient{
		InvokerClient: InvokerClient{
			Cli: http.DefaultClient,
			Url: url,
		},
		ProcessorName: processorName,
	}
}

func (pc *ProcessorClient) Initialize() (*InvokerResponse, error) {
	// build internal processor config from root config
	ipc := &InternalProcessorConfig{
		Name:              pc.ProcessorName,
		GlobalKafkaConfig: rootConfig.GlobalKafkaConfig,
	}

	kafkaAddr := rootConfig.GlobalKafkaConfig.Address

	for _, processConfig := range rootConfig.ProcessorConfigs {
		if processConfig.Name != pc.ProcessorName {
			continue
		}

		ipc.NumOfWorker = processConfig.NumOfWorker

		if processConfig.InputKafkaConfig != nil {
			ipc.InputKafkaConfig = processConfig.InputKafkaConfig
		} else {
			ipc.InputKafkaConfig = &KafkaConfig{
				Address: kafkaAddr,
				Topic:   processConfig.InputProcessor,
			}
		}

		if processConfig.OutputConfig.DefaultTopicPartitions > 0 {
			ipc.DefaultOutputKafkaConfig = &KafkaConfig{
				Address: kafkaAddr,
				Topic:   processConfig.Name,
			}
		}

		outputMap := make(map[string]*KafkaConfig, 0)
		for _, outputProcessor := range processConfig.OutputConfig.OutputProcessors {
			outputMap[outputProcessor] = &KafkaConfig{
				Address: kafkaAddr,
				Topic:   outputProcessor,
			}
		}
		for _, outputKafkaConfig := range processConfig.OutputConfig.OutputKafkaConfigs {
			key := outputKafkaConfig.Name
			val := &KafkaConfig{
				Address: outputKafkaConfig.Address,
				Topic:   outputKafkaConfig.Topic,
			}
			outputMap[key] = val
		}
		ipc.OutputKafkaConfigs = outputMap
	}
	if ipc.InputKafkaConfig == nil {
		err := fmt.Errorf("processor config for processor %s not found", pc.ProcessorName)
		logs.Printf("%v", err)
		return nil, err
	}

	params, err := MarshalToParams(ipc)
	if err != nil {
		err = fmt.Errorf("marshal InternalProcessorConfig to params failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	resp, err := pc.SendCommand(params, ProcessorCommands.Initialize)
	if err != nil {
		return nil, err
	}
	logs.Printf("processor %s client got response for initialize:\n%s", pc.ProcessorName, SafeJsonIndent(resp))
	return resp, nil
}
