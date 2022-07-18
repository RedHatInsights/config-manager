package kafka

import (
	config "config-manager/internal/config"
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	kafka "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
)

type KafkaWriterInterface interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
}

type MockWriter struct {
	mock.Mock
}

func (w *MockWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := w.Called(ctx, msgs)
	return args.Error(0)
}

// NewConsumer creates a configured kafka.Reader.
func NewConsumer(topic string) *kafka.Reader {
	if config.DefaultConfig.KafkaUsername == nil 
		|| config.DefaultConfig.KafkaPassword == nil 
		|| config.DefaultConfig.KafkaSASLMech == nil 
		|| config.DefaultConfig.KafkaProtocol == nil 
	{
		consumer := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     config.DefaultConfig.KafkaBrokers.Values,
			Topic:       topic,
			GroupID:     config.DefaultConfig.KafkaGroupID,
			StartOffset: config.DefaultConfig.KafkaConsumerOffset,
		})
	} else {
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
			Dialer: dialer
		})
	}

	return consumer
}

// NewProducer creates a configured kafka.Writer.
func NewProducer(topic string) *kafka.Writer {
	if config.DefaultConfig.KafkaUsername == nil 
		|| config.DefaultConfig.KafkaPassword == nil 
		|| config.DefaultConfig.KafkaSASLMech == nil 
		|| config.DefaultConfig.KafkaProtocol == nil 
	{
		producer := &kafka.Writer{
			Addr:  kafka.TCP(config.DefaultConfig.KafkaBrokers.Values[0]),
			Topic: topic,
		}
	}else {
		mechanism := plain.Mechanism{
			Username: config.DefaultConfig.KafkaUsername,
			Password: config.DefaultConfig.KafkaPassword,
		}
		sharedTransport := &kafka.Transport{
			SASL: mechanism,
			TLS: &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		}
		producer := &kafka.Writer{
			Addr:  kafka.TCP(config.DefaultConfig.KafkaBrokers.Values[0]),
			Topic: topic,
			Transport: sharedTransport,
		}
	}

	return producer
}

// NewConsumerEventLoop returns a function handler (start) that can be called to
// return a function handler that can be called to start reading messages from
// consumer. For every message read, handler is called.
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
				log.Info().Err(err)
				errors <- err
			}
			go handler(ctx, m)
		}
	}
}

// GetHeader loops over the message headers, returning the value of key, if
// found.
func GetHeader(msg kafka.Message, key string) (string, error) {
	for _, value := range msg.Headers {
		if value.Key == key {
			return string(value.Value), nil
		}
	}

	return "", fmt.Errorf("Header not found: %s", key)
}
