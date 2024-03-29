package core

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/TTraveller7/invokerlib/pkg/conf"
	"github.com/TTraveller7/invokerlib/pkg/models"
)

func PassToDefaultOutputTopic(ctx context.Context, record *models.Record) error {
	producer, err := defaultProducer()
	if err != nil {
		return err
	}
	producer.produce(&sarama.ProducerMessage{
		Key:   sarama.ByteEncoder(record.Key()),
		Value: sarama.ByteEncoder(record.Value()),
	})
	return nil
}

func PassToOutputTopic(ctx context.Context, name string, record *models.Record) error {
	c := conf.Config()
	kafkaDest, exists := c.OutputKafkaConfigs[name]
	if !exists {
		return fmt.Errorf("output topic with name %s does not exist", name)
	}
	producer, err := getProducer(kafkaDest.Topic)
	if err != nil {
		return err
	}
	producer.produce(&sarama.ProducerMessage{
		Key:   sarama.ByteEncoder(record.Key()),
		Value: sarama.ByteEncoder(record.Value()),
	})
	return nil
}
