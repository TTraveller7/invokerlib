package invokerlib

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

var consumerGroup sarama.ConsumerGroup

func initConsumer() error {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0

	// start consuming from the oldest offset
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	config.Consumer.Return.Errors = true

	grp, err := sarama.NewConsumerGroup([]string{conf.InputKafkaConfig.Address}, conf.Name, config)
	if err != nil {
		return fmt.Errorf("initialize consumer failed: %v", err)
	}
	consumerGroup = grp

	return nil
}

type workerConsumerHandler struct {
	logs                *log.Logger
	setup               func() error
	consume             func(record *Record) error
	workerNotifyChannel <-chan string
}

func (h workerConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	h.logs.Println("Consumer Setup invoked")
	if h.setup != nil {
		return h.setup()
	}
	return nil
}

func (h workerConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	h.logs.Println("Consumer Cleanup invoked")
	return nil
}

func (h workerConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {

	h.logs.Println("Consumer ConsumeClaim invoked")

	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				h.logs.Println("Message channel closed, exiting ConsumeClaim")
				return nil
			}
			r := &Record{
				Key:   string(msg.Key),
				Value: msg.Value,
			}
			if err := h.consume(r); err != nil {
				return err
			}
			session.MarkMessage(msg, "")
		case notify := <-h.workerNotifyChannel:
			if notify == "exit" {
				return errConsumerNotify
			}
		}
	}
}

func NewConsumerGroupHandler(logs *log.Logger, setupFunc func() error, consumeFunc func(record *Record) error,
	workerNotifyChannel <-chan string) sarama.ConsumerGroupHandler {

	return &workerConsumerHandler{
		logs:                logs,
		setup:               setupFunc,
		consume:             consumeFunc,
		workerNotifyChannel: workerNotifyChannel,
	}
}

func closeConsumerGroup() {
	consumerGroup.Close()
}
