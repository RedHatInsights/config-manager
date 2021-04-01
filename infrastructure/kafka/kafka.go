package kafka

import (
	"context"
	"fmt"

	kafka "github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
)

func NewConsumer(cfg *viper.Viper, topic string) *kafka.Reader {
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.GetStringSlice("Kafka_Brokers"),
		Topic:       topic,
		GroupID:     cfg.GetString("Kafka_Group_ID"),
		StartOffset: cfg.GetInt64("Kafka_Consumer_Offset"),
	})

	return consumer
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
