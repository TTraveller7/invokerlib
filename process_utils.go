package invokerlib

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
)

func PassToOutputTopic(ctx context.Context, name string, record *Record) error {
	kafkaDest, exists := destNameToKafkaConfig[name]
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
