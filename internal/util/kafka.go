package util

import (
	"config-manager/internal/config"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

var Kafka kafkautil

type kafkautil struct{}

// NewReader creates a configured kafka.Reader.
func (k kafkautil) NewReader(topic string) *kafka.Reader {
	var dialer *kafka.Dialer = kafka.DefaultDialer

	if config.DefaultConfig.KafkaSecurityProtocol == "SASL_SSL" {

		saslMechanism, tlsConfig := getSaslAndTLSConfig()
		dialer = &kafka.Dialer{
			Timeout:       10 * time.Second,
			DualStack:     true,
			TLS:           tlsConfig,
			SASLMechanism: saslMechanism,
		}
	}

	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     config.DefaultConfig.KafkaBrokers.Values,
		Topic:       topic,
		GroupID:     config.DefaultConfig.KafkaGroupID,
		StartOffset: kafka.LastOffset,
		Dialer:      dialer,
	})
}

// NewWriter creates a configured kafka.Writer.
func (k kafkautil) NewWriter(topic string) *kafka.Writer {
	var transport *kafka.Transport = kafka.DefaultTransport.(*kafka.Transport)

	if config.DefaultConfig.KafkaSecurityProtocol == "SASL_SSL" {
		saslMechanism, tlsConfig := getSaslAndTLSConfig()
		transport = &kafka.Transport{
			TLS:  tlsConfig,
			SASL: saslMechanism,
		}
	}

	return &kafka.Writer{
		Addr:      kafka.TCP(config.DefaultConfig.KafkaBrokers.Values...),
		Topic:     topic,
		Transport: transport,
	}
}

func getSaslAndTLSConfig() (sasl.Mechanism, *tls.Config) {
	username := config.DefaultConfig.KafkaUsername
	password := config.DefaultConfig.KafkaPassword
	saslmechanismName := config.DefaultConfig.KafkaSaslMechanism

	tlsConfig, err := createTLSConfig(config.DefaultConfig.KafkaCAPath)
	if err != nil {
		log.Error().Err(err).Msg("Error creating TLS configuration for Kafka:")
		tlsConfig = &tls.Config{} // Providing default empty TLS configuration
	}

	saslMechanism, err := createSaslMechanism(
		saslmechanismName,
		username,
		password,
	)
	if err != nil {
		log.Error().Err(err).Msg("Error creating SASL Mechanism for Kafka, using plain mechanism")
	}

	return saslMechanism, tlsConfig
}

func createTLSConfig(pathToCert string) (*tls.Config, error) {

	tlsConfig := tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	}

	if pathToCert == "" {
		return &tlsConfig, nil
	}

	caCert, err := os.ReadFile(pathToCert)
	if err != nil {
		return nil, fmt.Errorf("unable to open cert file (%s): %w", pathToCert, err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig.RootCAs = caCertPool

	return &tlsConfig, nil
}

func createSaslMechanism(saslMechanism string, username string, password string) (sasl.Mechanism, error) {

	switch strings.ToLower(saslMechanism) {
	case "plain":
		return createPlainMechanism(username, password), nil

	case "scram-sha-512":
		return createScramMechanism(scram.SHA512, username, password)

	case "scram-sha-256":
		return createScramMechanism(scram.SHA256, username, password)

	default:
		// create plain mechanism as default
		log.Error().Msgf("unable to configure sasl mechanism (%s)", saslMechanism)
		return createPlainMechanism(username, password), fmt.Errorf("unable to configure sasl mechanism (%s)", saslMechanism)
	}
}

func createPlainMechanism(username, password string) sasl.Mechanism {
	return plain.Mechanism{
		Username: username,
		Password: password,
	}
}

func createScramMechanism(hash scram.Algorithm, username, password string) (sasl.Mechanism, error) {
	mechanism, err := scram.Mechanism(hash, username, password)
	if err != nil {
		log.Error().Err(err).Msgf("unable to create scram mechanism (%s)", hash)
	}
	return mechanism, err
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
