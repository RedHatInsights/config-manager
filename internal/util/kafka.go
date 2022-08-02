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
	if config.DefaultConfig.KafkaUsername != "" && config.DefaultConfig.KafkaPassword != "" {
		mechanism := plain.Mechanism{
			Username: config.DefaultConfig.KafkaUsername,
			Password: config.DefaultConfig.KafkaPassword,
		}
		dialer := &kafka.Dialer{
			Timeout:       10 * time.Second,
			DualStack:     true,
			SASLMechanism: mechanism,
		}
		consumer := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     config.DefaultConfig.KafkaBrokers.Values,
			Topic:       topic,
			GroupID:     config.DefaultConfig.KafkaGroupID,
			StartOffset: config.DefaultConfig.KafkaConsumerOffset,
			Dialer:      dialer,
		})

		return consumer
	} else {
		consumer := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     config.DefaultConfig.KafkaBrokers.Values,
			Topic:       topic,
			GroupID:     config.DefaultConfig.KafkaGroupID,
			StartOffset: config.DefaultConfig.KafkaConsumerOffset,
		})

		return consumer
	}
}

// NewWriter creates a configured kafka.Writer.
func (k kafkautil) NewWriter(topic string) *kafka.Writer {
	if config.DefaultConfig.KafkaUsername != "" && config.DefaultConfig.KafkaPassword != "" {
		mechanism := plain.Mechanism{
			Username: config.DefaultConfig.KafkaUsername,
			Password: config.DefaultConfig.KafkaPassword,
		}
		sharedTransport := &kafka.Transport{
			SASL: mechanism,
			TLS: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}
		producer := &kafka.Writer{
			Addr:      kafka.TCP(config.DefaultConfig.KafkaBrokers.Values[0]),
			Topic:     topic,
			Transport: sharedTransport,
		}

		return producer
	} else {
		producer := &kafka.Writer{
			Addr:  kafka.TCP(config.DefaultConfig.KafkaBrokers.Values[0]),
			Topic: topic,
		}

		return producer
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
