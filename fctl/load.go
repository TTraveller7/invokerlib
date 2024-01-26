package main

import (
	"bufio"
	"os"

	"github.com/IBM/sarama"
	"github.com/spf13/pflag"
)

func Load() {
	workloadPathPtr := pflag.StringP("workload", "w", "", "workload path")
	brokerPtr := pflag.StringP("broker", "b", "", "Kafka broker address")
	topicPtr := pflag.StringP("topic", "t", "", "topic to load")
	pflag.Parse()
	if workloadPathPtr == nil || len(*workloadPathPtr) == 0 {
		logs.Printf("workload path is not provided. Use -w <path> to provide config path. ")
		return
	}
	if brokerPtr == nil || len(*brokerPtr) == 0 {
		logs.Printf("Kafka broker adddress is not provided. Use -b <address> to provide Kafka broker adddress. ")
		return
	}
	if topicPtr == nil || len(*topicPtr) == 0 {
		logs.Printf("topic is not provided. Use -t <topic> to provide topic to load. ")
		return
	}
	logs.Printf("broker address: %s", *brokerPtr)

	workloadFile, err := os.OpenFile(*workloadPathPtr, os.O_RDONLY, 0644)
	if err != nil {
		logs.Printf("open file failed: err=%v, filePath=%s", err, *workloadPathPtr)
		return
	}
	defer workloadFile.Close()

	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{*brokerPtr}, producerConfig)
	if err != nil {
		logs.Printf("create producer failed: %v", err)
		return
	}
	defer producer.Close()

	scanner := bufio.NewScanner(workloadFile)
	var offset int64
	for scanner.Scan() {
		msg := &sarama.ProducerMessage{
			Topic: *topicPtr,
			Value: sarama.StringEncoder(scanner.Text()),
		}
		_, offset, err = producer.SendMessage(msg)
		if err != nil {
			logs.Printf("produce message failed: %v", err)
			return
		}
	}
	logs.Printf("load completed. final offset: %v", offset)
}
