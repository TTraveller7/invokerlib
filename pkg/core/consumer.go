package core

import (
	"log"
	"sync"

	"github.com/IBM/sarama"
	"github.com/TTraveller7/invokerlib/pkg/models"
)

var consumerGroups []sarama.ConsumerGroup

func initConsumer() error {
	consumerGroups = make([]sarama.ConsumerGroup, 0)

	return nil
}

func defaultSaramaConsumerConfig() *sarama.Config {
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
	workerReadyChannel  chan<- struct{}
	once                sync.Once
}

func (h *workerConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	h.logs.Println("Consumer Setup invoked")

	// signals to the main routine that this worker joins the consumer group already
	h.once.Do(func() {
		h.logs.Println("closing worker ready channel")
		close(h.workerReadyChannel)
	})

	if h.setup != nil {
		return h.setup()
	}
	return nil
}

func (h *workerConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	h.logs.Println("Consumer Cleanup invoked")
	return nil
}

func (h *workerConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {

	h.logs.Println("Consumer ConsumeClaim invoked")

	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				h.logs.Println("Message channel closed, exiting ConsumeClaim")
				return nil
			}
			r := models.NewRecordWithConsumerMessage(msg)
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
	workerNotifyChannel <-chan string, workerReadyChannel chan<- struct{}) sarama.ConsumerGroupHandler {

	return &workerConsumerHandler{
		logs:                logs,
		setup:               setupFunc,
		consume:             consumeFunc,
		workerNotifyChannel: workerNotifyChannel,
		workerReadyChannel:  workerReadyChannel,
	}
}

func closeConsumerGroup() {
	for _, grp := range consumerGroups {
		grp.Close()
	}
}
