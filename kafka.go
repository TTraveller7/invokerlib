package invokerlib

import (
	"github.com/IBM/sarama"
)

var consumerGroup sarama.ConsumerGroup

func InitConsumer() {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0
	config.Consumer.Return.Errors = true

	grp, err := sarama.NewConsumerGroup([]string{conf.KafkaSrc.Address}, conf.FunctionName, config)
	if err != nil {
		logs.Panic("initialize consumer failed: %v", err)
	}
	consumerGroup = grp
}

type workerConsumerHandler struct {
	setup   func(session sarama.ConsumerGroupSession) error
	cleanup func(session sarama.ConsumerGroupSession) error
	consume func(record *Record) error
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
	for msg := range claim.Messages() {
		r := &Record{
			Key:   string(msg.Key),
			Value: msg.Value,
		}
		if err := h.consume(r); err != nil {
			return err
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

func NewConsumerGroupHandler(consumeFunc func(record *Record) error) sarama.ConsumerGroupHandler {
	return &workerConsumerHandler{
		consume: consumeFunc,
	}
}
