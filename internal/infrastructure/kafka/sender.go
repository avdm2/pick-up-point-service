package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"homework-1/internal/infrastructure/messaging"
	"homework-1/internal/infrastructure/messaging/messages"
)

type KafkaSender struct {
	producer *messaging.Producer
	topic    string
}

func NewKafkaSender(producer *messaging.Producer, topic string) *KafkaSender {
	return &KafkaSender{
		producer,
		topic,
	}
}

func (s *KafkaSender) SendMessage(message *messages.CLIMessage) error {
	kafkaMsg, err := s.buildMessage(*message)
	if err != nil {
		return fmt.Errorf("sender.SendMessage error: %w", err)
	}

	_, _, err = s.producer.ProduceMessage(kafkaMsg)
	if err != nil {
		return fmt.Errorf("sender.SendMessage error: %w", err)
	}

	return nil
}

func (s *KafkaSender) buildMessage(message messages.CLIMessage) (*sarama.ProducerMessage, error) {
	msg, err := json.Marshal(message)

	if err != nil {
		return nil, fmt.Errorf("sender.buildMessage error: %w", err)
	}

	return &sarama.ProducerMessage{
		Topic: s.topic,
		Value: sarama.ByteEncoder(msg),
	}, nil
}
