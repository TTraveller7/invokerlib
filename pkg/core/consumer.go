package core

import (
	"log"

	"github.com/IBM/sarama"
	"github.com/TTraveller7/invokerlib/pkg/models"
)

var consumerGroups []sarama.ConsumerGroup

func initConsumer() error {
	consumerGroups = make([]sarama.ConsumerGroup, 0)

	return nil
}

func consumerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0

	// start consuming from the oldest offset
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	config.Consumer.Return.Errors = true
	return config
}

type workerConsumerHandler struct {
	sarama.ConsumerGroupHandler
	logs                *log.Logger
	setup               func() error
	consume             func(record *models.Record) error
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
			r := models.NewRecord(msg)
			if err := h.consume(r); err != nil {
				metricsClient.EmitCounter("consume_error", "Number of messages that are not successfully consumed", 1)
				return err
			}
			session.MarkMessage(msg, "")
			metricsClient.EmitCounter("consume_success", "Number of messages that are successfully consumed", 1)
		case notify := <-h.workerNotifyChannel:
			if notify == "exit" {
				return ErrConsumerNotify
			}
		}
	}
}

func NewConsumerGroupHandler(logs *log.Logger, setupFunc func() error, consumeFunc func(record *models.Record) error,
	workerNotifyChannel <-chan string) sarama.ConsumerGroupHandler {

	return &workerConsumerHandler{
		logs:                logs,
		setup:               setupFunc,
		consume:             consumeFunc,
		workerNotifyChannel: workerNotifyChannel,
	}
}

func closeConsumerGroup() {
	for _, grp := range consumerGroups {
		grp.Close()
	}
}
