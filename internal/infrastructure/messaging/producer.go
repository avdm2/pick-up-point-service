package messaging

import (
	"fmt"
	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
}

func NewKafkaProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("messaging.NewKafkaProducer error: %w", err)
	}

	return &Producer{
		producer: producer,
	}, nil
}

func (p *Producer) ProduceMessage(message *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	return p.producer.SendMessage(message)
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
