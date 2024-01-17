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
	setup               func(session sarama.ConsumerGroupSession) error
	cleanup             func(session sarama.ConsumerGroupSession) error
	consume             func(record *Record) error
	workerNotifyChannel <-chan string
}

func (h workerConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	if h.setup == nil {
		return nil
	}
	return h.setup(session)
}

func (h workerConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	if h.cleanup == nil {
		return nil
	}
	return h.cleanup(session)
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

func NewConsumerGroupHandler(consumeFunc func(record *Record) error, workerNotifyChannel <-chan string) sarama.ConsumerGroupHandler {
	return &workerConsumerHandler{
		consume:             consumeFunc,
		workerNotifyChannel: workerNotifyChannel,
	}
}

func closeConsumerGroup() {
	consumerGroup.Close()
}
