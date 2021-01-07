package kafka

import (
	kafka "github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
)

func NewResultsConsumer(cfg *viper.Viper) *kafka.Reader {
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.GetStringSlice("KafkaBrokers"),
		Topic:       cfg.GetString("KafkaResultsTopic"),
		GroupID:     cfg.GetString("KafkaGroupID"),
		StartOffset: cfg.GetInt64("KafkaConsumerOffset"),
	})

	return consumer
}

func NewConnectionsConsumer(cfg *viper.Viper) *kafka.Reader {
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.GetStringSlice("KafkaBrokers"),
		Topic:       cfg.GetString("KafkaConnectionsTopic"),
		GroupID:     cfg.GetString("KafkaGroupID"),
		StartOffset: cfg.GetInt64("KafkaConsumerOffset"),
	})

	return consumer
}
