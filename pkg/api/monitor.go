package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/TTraveller7/invokerlib/pkg/conf"
	"github.com/TTraveller7/invokerlib/pkg/logs"
	"github.com/TTraveller7/invokerlib/pkg/utils"
)

var MonitorCommands = struct {
	LoadRootConfig         string
	CreateTopics           string
	LoadProcessorEndpoints string
	InitializeProcessors   string
	RunProcessors          string
	Load                   string
	CatProcessor           string
}{
	LoadRootConfig:         "loadRootConfig",
	CreateTopics:           "createTopics",
	LoadProcessorEndpoints: "loadProcessorEndpoints",
	InitializeProcessors:   "initializeProcessors",
	RunProcessors:          "runProcessors",
	Load:                   "load",
	CatProcessor:           "catProcessor",
}

type ProcessorMetadata struct {
	Name   string
	Status string
	Client *ProcessorClient
}

var (
	monitorMut sync.Mutex = sync.Mutex{}

	rootConfig        *conf.RootConfig = &conf.RootConfig{}
	adminClient       sarama.ClusterAdmin
	initialTopics     []*conf.InternalKafkaConfig   = make([]*conf.InternalKafkaConfig, 0)
	interimTopics     []*conf.InternalKafkaConfig   = make([]*conf.InternalKafkaConfig, 0)
	processorMetadata map[string]*ProcessorMetadata = make(map[string]*ProcessorMetadata, 0)
)

func MonitorHandle(w http.ResponseWriter, r *http.Request) {
	logs.SetPrefix("[monitor] ")

	switch r.URL.Path {
	case "/metrics":
		utils.MetricsHandler().ServeHTTP(w, r)
		return
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

	var content []byte
	content, err = io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("read request body failed: %v", err)
		logs.Printf("%v", err)
		return
	}

	req := &InvokerRequest{}
	if err = json.Unmarshal(content, req); err != nil {
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
	case MonitorCommands.RunProcessors:
		return runProcessors()
	case MonitorCommands.Load:
		return load(req)
	case MonitorCommands.CatProcessor:
		return catProcessor(req)
	default:
		err := fmt.Errorf("unrecognized command %v", req.Command)
		logs.Printf("%v", err)
		return nil, err
	}
}

func loadRootConfig(req *InvokerRequest) (*InvokerResponse, error) {
	logs.Printf("monitor load root config starts")
	if lockSuccess := monitorMut.TryLock(); !lockSuccess {
		err := fmt.Errorf("fail to lock monitor: another client holds the lock")
		logs.Printf("%v", err)
		return nil, err
	}
	defer monitorMut.Unlock()

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
	if lockSuccess := monitorMut.TryLock(); !lockSuccess {
		err := fmt.Errorf("fail to lock monitor: another client holds the lock")
		logs.Printf("%v", err)
		return nil, err
	}
	defer monitorMut.Unlock()

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

	// create global initial topics
	logs.Printf("start to create global initial topics")
	for _, kc := range rootConfig.GlobalKafkaConfig.InitialTopics {
		err := tryCreateTopic(kc.Topic, kc.Partitions)
		if err != nil {
			err = fmt.Errorf("try create topic failed: %v", err)
			logs.Printf("%v", err)
			if _, err := removeTopics(); err != nil {
				logs.Printf("remove topics failed: %v", err)
				return nil, err
			} else {
				return nil, err
			}
		}

		initialTopics = append(initialTopics, &conf.InternalKafkaConfig{
			Address:    kafkaAddr,
			Topic:      kc.Topic,
			Partitions: kc.Partitions,
		})
	}
	logs.Printf("finish creating global initial topics")

	// create interim topics for processors
	logs.Printf("start to create interim topics for processors")
	for _, pc := range rootConfig.ProcessorConfigs {
		numOfPartitions := pc.OutputConfig.DefaultTopicPartitions
		if numOfPartitions == 0 {
			// skip output topic creation of this processor
			continue
		}

		processorName := pc.Name

		// use processor name as topic name
		topicName := processorName
		err := tryCreateTopic(topicName, numOfPartitions)
		if err != nil {
			err = fmt.Errorf("try create topic failed: %v", err)
			logs.Printf("%v", err)
			if _, err := removeTopics(); err != nil {
				logs.Printf("remove topics failed: %v", err)
				return nil, err
			} else {
				return nil, err
			}
		}

		interimTopics = append(interimTopics, &conf.InternalKafkaConfig{
			Address:    kafkaAddr,
			Topic:      topicName,
			Partitions: numOfPartitions,
		})
	}
	logs.Printf("finish creating interim topics for processors")

	for _, pc := range rootConfig.ProcessorConfigs {
		meta := &ProcessorMetadata{
			Name: pc.Name,
		}
		processorMetadata[pc.Name] = meta
	}
	logs.Printf("processorMetadata: %s", utils.SafeJsonIndent(processorMetadata))

	logs.Printf("monitor create topics succeeds")
	return successResponse(), nil
}

func tryCreateTopic(topic string, partitions int) error {
	if adminClient == nil {
		err := fmt.Errorf("admin client must be initialized before calling tryCreateTopic")
		logs.Printf("%v", err)
		return err
	}

	topics, err := adminClient.ListTopics()
	if err != nil {
		err = fmt.Errorf("list topics failed: %v", err)
		logs.Printf("%v", err)
		return err
	}
	for existingTopic := range topics {
		if existingTopic == topic {
			return nil
		}
	}

	err = adminClient.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     int32(partitions),
		ReplicationFactor: 1,
	}, false)
	if err != nil {
		err = fmt.Errorf("create topic failed: %v", err)
		logs.Printf("%v", err)
		return err
	}
	return nil
}

func removeTopics() (*InvokerResponse, error) {
	for _, kc := range initialTopics {
		if err := adminClient.DeleteTopic(kc.Topic); err != nil {
			return nil, err
		}
	}
	for _, kc := range interimTopics {
		if err := adminClient.DeleteTopic(kc.Topic); err != nil {
			return nil, err
		}
	}
	return successResponse(), nil
}

func loadProcessorEndpoints(req *InvokerRequest) (*InvokerResponse, error) {
	logs.Printf("monitor load processor endpoints starts")
	if lockSuccess := monitorMut.TryLock(); !lockSuccess {
		err := fmt.Errorf("fail to lock monitor: another client holds the lock")
		logs.Printf("%v", err)
		return nil, err
	}
	defer monitorMut.Unlock()

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
	if lockSuccess := monitorMut.TryLock(); !lockSuccess {
		err := fmt.Errorf("fail to lock monitor: another client holds the lock")
		logs.Printf("%v", err)
		return nil, err
	}
	defer monitorMut.Unlock()

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

func runProcessors() (*InvokerResponse, error) {
	logs.Printf("monitor run processors starts")
	if lockSuccess := monitorMut.TryLock(); !lockSuccess {
		err := fmt.Errorf("fail to lock monitor: another client holds the lock")
		logs.Printf("%v", err)
		return nil, err
	}
	defer monitorMut.Unlock()

	for _, metadata := range processorMetadata {
		if metadata.Client == nil {
			err := fmt.Errorf("processor %s client is not initialized", metadata.Name)
			logs.Printf("%v", err)
			return nil, err
		}
		resp, err := metadata.Client.Run()
		if err != nil {
			err = fmt.Errorf("processor run failed: %v", err)
			logs.Printf("%v", err)
			return nil, err
		} else if resp.Code != ResponseCodes.Success {
			err = fmt.Errorf("processor run failed with resp: %+v", resp)
			logs.Printf("%v", err)
			return nil, err
		}
		logs.Printf("processor run finished with resp: %+v", resp)
	}

	logs.Printf("monitor run processors finished")
	return successResponse(), nil
}

func load(req *InvokerRequest) (*InvokerResponse, error) {
	logs.Printf("monitor load starts")
	if lockSuccess := monitorMut.TryLock(); !lockSuccess {
		err := fmt.Errorf("fail to lock monitor: another client holds the lock")
		logs.Printf("%v", err)
		return nil, err
	}
	defer monitorMut.Unlock()

	loadParams := &LoadParams{}
	if err := UnmarshalParams(req.Params, loadParams); err != nil {
		err = fmt.Errorf("unmarshal params failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	// retrieve workload with url
	workloadResp, err := http.Get(loadParams.Url)
	if err != nil {
		err = fmt.Errorf("retrieve workload failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}
	defer workloadResp.Body.Close()
	if workloadResp.StatusCode != http.StatusOK {
		err = fmt.Errorf("retrieve workload failed: http status is %v", workloadResp.Status)
		logs.Printf("%v", err)
		return nil, err
	}

	file, err := os.OpenFile(loadParams.Name, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		err = fmt.Errorf("open file failed: err=%v, fileName=%s", err, loadParams.Name)
		logs.Printf("%v", err)
		return nil, err
	}
	defer file.Close()

	written, err := io.Copy(file, workloadResp.Body)
	if err != nil {
		err = fmt.Errorf("copy response to file failed: err=%v, fileName=%s", err, loadParams.Name)
		logs.Printf("%v", err)
		return nil, err
	}
	logs.Printf("copy body to file finished, written=%v", written)
	file.Close()

	file, err = os.OpenFile(loadParams.Name, os.O_RDONLY, 0644)
	if err != nil {
		err = fmt.Errorf("open file failed: err=%v, fileName=%s", err, loadParams.Name)
		logs.Printf("%v", err)
		return nil, err
	}

	producerConfig := sarama.NewConfig()
	producerConfig.Producer.RequiredAcks = sarama.NoResponse
	// TODO: change this to true
	producerConfig.Producer.Return.Errors = false
	producer, err := sarama.NewAsyncProducer([]string{rootConfig.GlobalKafkaConfig.Address}, producerConfig)
	if err != nil {
		err = fmt.Errorf("create producer failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}
	defer func() {
		logs.Printf("shutting down producer")
		producer.Close()
		logs.Printf("produce shut down")
	}()

	count := 0
	startTime := time.Now()
	for _, topic := range loadParams.Topics {
		s := bufio.NewScanner(file)
		var offset int64
		for s.Scan() {
			msg := &sarama.ProducerMessage{
				Topic: topic,
				Value: sarama.ByteEncoder(s.Bytes()),
			}
			producer.Input() <- msg
			count++
			if count%10000 == 0 {
				logs.Printf("%v messages submitted. Elapsed time: %v", count, time.Since(startTime))
			}
		}
		logs.Printf("produce finished: fileName=%s, topic=%s, final offset=%v", loadParams.Name, topic, offset)
	}

	return successResponse(), nil
}

func catProcessor(req *InvokerRequest) (*InvokerResponse, error) {
	p := &CatProcessorParams{}
	if err := UnmarshalParams(req.Params, p); err != nil {
		err = fmt.Errorf("unmarshal params failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	m, exists := processorMetadata[p.ProcessorName]
	if !exists {
		err := fmt.Errorf("processor with name %s does not exist", p.ProcessorName)
		logs.Printf("%v", err)
		return nil, err
	}

	catResp, err := m.Client.Cat()
	if err != nil {
		err = fmt.Errorf("processor cat failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	} else if catResp.Code != ResponseCodes.Success {
		err = fmt.Errorf("processor cat failed with resp: %+v", catResp)
		logs.Printf("%v", err)
		return nil, err
	}
	logs.Printf("processor cat finished with resp: %+v", catResp)

	resp := successResponse()
	resp.Message = catResp.Message
	logs.Printf("monitor cat processor finished")
	return resp, nil
}
