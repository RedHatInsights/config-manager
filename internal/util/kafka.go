package util

import (
	"config-manager/internal/config"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

var Kafka kafkautil

type kafkautil struct{}

// NewReader creates a configured kafka.Reader.
func (k kafkautil) NewReader(topic string) *kafka.Reader {
	var dialer *kafka.Dialer

	if config.DefaultConfig.KafkaUsername != "" && config.DefaultConfig.KafkaPassword != "" {
		dialer = &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
			SASLMechanism: plain.Mechanism{
				Username: config.DefaultConfig.KafkaUsername,
				Password: config.DefaultConfig.KafkaPassword,
			},
		}
	}

	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     config.DefaultConfig.KafkaBrokers.Values,
		Topic:       topic,
		GroupID:     config.DefaultConfig.KafkaGroupID,
		StartOffset: config.DefaultConfig.KafkaConsumerOffset,
		Dialer:      dialer,
	})
}

// NewWriter creates a configured kafka.Writer.
func (k kafkautil) NewWriter(topic string) *kafka.Writer {
	var transport *kafka.Transport

	if config.DefaultConfig.KafkaUsername != "" && config.DefaultConfig.KafkaPassword != "" {
		transport = &kafka.Transport{
			SASL: plain.Mechanism{
				Username: config.DefaultConfig.KafkaUsername,
				Password: config.DefaultConfig.KafkaPassword,
			},
			TLS: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}

	}

	return &kafka.Writer{
		Addr:      kafka.TCP(config.DefaultConfig.KafkaBrokers.Values[0]),
		Topic:     topic,
		Transport: transport,
	}
}

// GetHeader loops over the message headers, returning the value of key, if
// found.
func (k kafkautil) GetHeader(message kafka.Message, key string) (string, error) {
	for _, value := range message.Headers {
		if value.Key == key {
			return string(value.Value), nil
		}
	}

	return "", fmt.Errorf("cannot find header with value %v", key)
}
