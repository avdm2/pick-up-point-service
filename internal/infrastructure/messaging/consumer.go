package messaging

import (
	"fmt"
	"github.com/IBM/sarama"
	"time"
)

type Consumer struct {
	SingleConsumer sarama.Consumer
}

func NewKafkaConsumer(brokers []string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = false
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 5 * time.Second
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("messaging.NewKafkaConsumer error: %w", err)
	}

	return &Consumer{
		SingleConsumer: consumer,
	}, nil
}
