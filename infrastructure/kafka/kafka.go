package kafka

import (
	kafka "github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
)

func NewResultsConsumer(cfg *viper.Viper) *kafka.Reader {
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.GetStringSlice("Kafka_Brokers"),
		Topic:       cfg.GetString("Kafka_Results_Topic"),
		GroupID:     cfg.GetString("Kafka_Group_ID"),
		StartOffset: cfg.GetInt64("Kafka_Consumer_Offset"),
	})

	return consumer
}

func NewConnectionsConsumer(cfg *viper.Viper) *kafka.Reader {
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.GetStringSlice("Kafka_Brokers"),
		Topic:       cfg.GetString("Kafka_Connections_Topic"),
		GroupID:     cfg.GetString("Kafka_Group_ID"),
		StartOffset: cfg.GetInt64("Kafka_Consumer_Offset"),
	})

	return consumer
}
