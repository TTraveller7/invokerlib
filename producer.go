package invokerlib

import (
	"fmt"
	"sync"

	"github.com/IBM/sarama"
)

type Producer struct {
	saramaProducer sarama.SyncProducer
	topic          string
}

func (p *Producer) produce(msg *sarama.ProducerMessage) error {
	msg.Topic = p.topic
	_, _, err := p.saramaProducer.SendMessage(msg)
	return err
}

var producers sync.Map

func initProducers() error {
	producers = sync.Map{}

	// one sarama producer per address
	addrToSaramaProducer := make(map[string]sarama.SyncProducer, 0)

	for _, producerConf := range destNameToKafkaConfig {
		if _, exists := addrToSaramaProducer[producerConf.Address]; !exists {
			// create sarama producer
			saramaProducer, err := sarama.NewSyncProducer([]string{producerConf.Address}, sarama.NewConfig())
			if err != nil {
				return fmt.Errorf("initialize producer failed: %v", err)
			}
			addrToSaramaProducer[producerConf.Address] = saramaProducer
		}

		producer := &Producer{
			saramaProducer: addrToSaramaProducer[producerConf.Address],
			topic:          producerConf.Topic,
		}
		producers.Store(producerConf.Topic, producer)
	}
	return nil
}

func getProducer(topic string) (*Producer, error) {
	producer, exists := producers.Load(topic)
	if !exists {
		return nil, fmt.Errorf("producer with topic %s does not exist", topic)
	}
	return producer.(*Producer), nil
}

func closeProducers() {
	producers.Range(func(topic any, producer any) bool {
		p := producer.(*Producer)
		p.saramaProducer.Close()
		return true
	})
}
