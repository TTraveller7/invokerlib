package invokerlib

import (
	"fmt"

	"github.com/IBM/sarama"
)

var consumerGroup sarama.ConsumerGroup

func initConsumer() error {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0
	config.Consumer.Return.Errors = true

	grp, err := sarama.NewConsumerGroup([]string{conf.InputKafkaConfig.Address}, conf.Name, config)
	if err != nil {
		return fmt.Errorf("initialize consumer failed: %v", err)
	}
	consumerGroup = grp

	return nil
}

type workerConsumerHandler struct {
	setup               func() error
	consume             func(record *Record) error
	workerNotifyChannel <-chan string
	workerReadyChannel  chan<- bool
}

func (h workerConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	if h.setup != nil {
		return h.setup()
	}
	return nil
}

func (h workerConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h workerConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg := <-claim.Messages():
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

func NewConsumerGroupHandler(setupFunc func() error, consumeFunc func(record *Record) error,
	workerNotifyChannel <-chan string) sarama.ConsumerGroupHandler {

	return &workerConsumerHandler{
		setup:               setupFunc,
		consume:             consumeFunc,
		workerNotifyChannel: workerNotifyChannel,
	}
}

func closeConsumerGroup() {
	consumerGroup.Close()
}
