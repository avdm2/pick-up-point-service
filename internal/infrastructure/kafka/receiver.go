package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"homework-1/internal/infrastructure/messaging"
	"homework-1/internal/infrastructure/messaging/messages"
)

type KafkaReceiver struct {
	consumer *messaging.Consumer
	topic    string
	stop     chan struct{}
}

func NewKafkaReceiver(consumer *messaging.Consumer, topic string) *KafkaReceiver {
	return &KafkaReceiver{
		consumer: consumer,
		topic:    topic,
		stop:     make(chan struct{}),
	}
}

func (r *KafkaReceiver) Subscribe() error {
	partitionConsumer, err := r.consumer.SingleConsumer.ConsumePartition(r.topic, 0, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("receiver.Subscribe error: %w", err)
	}

	go func(pc sarama.PartitionConsumer) {
		defer pc.Close()

		for {
			select {
			case <-r.stop:
				fmt.Println("receiver.Subscribe: stopping Kafka receiver")
				return
			case msg, ok := <-pc.Messages():
				if !ok {
					fmt.Printf("receiver.Subscribe: channel closed")
					return
				}

				cliMessage := messages.CLIMessage{}
				err := json.Unmarshal(msg.Value, &cliMessage)
				if err != nil {
					fmt.Printf("receiver.Subscribe error: %s\n", err)
				}
				fmt.Printf("[* Kafka *] Полученное сообщение: %s\n", cliMessage)
			}
		}
	}(partitionConsumer)

	return nil
}

func (r *KafkaReceiver) Stop() {
	close(r.stop)
}
