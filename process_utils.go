package invokerlib

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
)

func PassToFunction(ctx context.Context, functionName string, record *Record) error {
	kafkaDest, exists := conf.FunctionNameToKafkaDest[functionName]
	if !exists {
		return fmt.Errorf("function destination with name %s does not exist", functionName)
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
