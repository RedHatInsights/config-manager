package kafka

import (
	"context"
	"errors"
	"fmt"

	kafka "github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
)

func NewConsumer(cfg *viper.Viper, topic string) (*kafka.Reader, error) {
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.GetStringSlice("Kafka_Brokers"),
		Topic:       topic,
		GroupID:     cfg.GetString("Kafka_Group_ID"),
		StartOffset: cfg.GetInt64("Kafka_Consumer_Offset"),
	})

	err := verifyTopic(cfg.GetStringSlice("Kafka_Brokers"), topic)

	return consumer, err
}

func NewConsumerEventLoop(
	ctx context.Context,
	consumer *kafka.Reader,
	handler func(context.Context, kafka.Message),
	errors chan<- error,
) (start func()) {
	return func() {
		for {
			m, err := consumer.ReadMessage(ctx)
			if err != nil {
				fmt.Println(err)
				errors <- err
			}
			handler(ctx, m)
		}
	}
}

func verifyTopic(brokers []string, topic string) error {
	for _, broker := range brokers {
		conn, err := kafka.Dial("tcp", broker)
		if err != nil {
			return err
		}
		defer conn.Close()

		partitions, err := conn.ReadPartitions()
		if err != nil {
			return err
		}

		m := map[string]struct{}{}
		for _, p := range partitions {
			m[p.Topic] = struct{}{}
		}

		if _, ok := m[topic]; ok {
			return nil
		}
	}

	return errors.New("Topic not found")
}
