package invokerlib

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"

	"github.com/IBM/sarama"
)

func ProcessorHandle(w http.ResponseWriter, r *http.Request, pf ProcessFunc, initF InitFunc) {
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

	content, readAllErr := io.ReadAll(r.Body)
	if readAllErr != nil {
		err = fmt.Errorf("read request body failed: %v", readAllErr)
		logs.Printf("%v", err)
		return
	}

	req := &InvokerRequest{}
	if unmarshalErr := json.Unmarshal(content, req); unmarshalErr != nil {
		err = fmt.Errorf("unmarshal request failed: %v", unmarshalErr)
		logs.Printf("%v", err)
		return
	}
	if req.Command == "" {
		err = fmt.Errorf("request command is missing")
		logs.Printf("%v", err)
		return
	}
	logs.Printf("received request:\n%s", string(SafeJsonIndent(req)))

	var handleErr error
	switch req.Command {
	case ProcessorCommands.Initialize:
		resp, handleErr = handleInitialize(req, pf, initF)
	case ProcessorCommands.Ping:
		resp = successResponse()
		resp.Message = "pong"
	default:
		err = fmt.Errorf("unrecognized command %s", req.Command)
		logs.Printf("%v", err)
		return
	}
	if handleErr != nil {
		err = fmt.Errorf("handle processor command failed: %v", handleErr)
		return
	}
}

func handleInitialize(req *InvokerRequest, pf ProcessFunc, initF InitFunc) (*InvokerResponse, error) {
	logs.Printf("handle initialize starts")
	ipc := &InternalProcessorConfig{}
	if err := UnmarshalParams(req.Params, ipc); err != nil {
		err = fmt.Errorf("unmarshal params failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}

	if err := Initialize(ipc, pf, initF); err != nil {
		err = fmt.Errorf("Initialize failed: %v", err)
		logs.Printf("%v", err)
		return nil, err
	}
	logs.Printf("handle initialize finished")
	return successResponse(), nil
}

func PassToDefaultOutputTopic(ctx context.Context, record *Record) error {
	producer, err := defaultProducer()
	if err != nil {
		return err
	}
	producer.produce(&sarama.ProducerMessage{
		Key:   sarama.StringEncoder(record.Key),
		Value: sarama.ByteEncoder(record.Value),
	})
	return nil
}

func PassToOutputTopic(ctx context.Context, name string, record *Record) error {
	kafkaDest, exists := conf.OutputKafkaConfigs[name]
	if !exists {
		return fmt.Errorf("output topic with name %s does not exist", name)
	}
	producer, err := getProducer(kafkaDest.Topic)
	if err != nil {
		return err
	}
	producer.produce(&sarama.ProducerMessage{
		Key:   sarama.StringEncoder(record.Key),
		Value: sarama.ByteEncoder(record.Value),
	})
	return nil
}
