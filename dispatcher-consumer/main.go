package dispatcherconsumer

import (
	"config-manager/infrastructure/kafka"
	"config-manager/internal/config"
	"context"

	"github.com/google/uuid"
)

// Start creates a new Kafka consumer and producer, sets up a message handler
// and starts running the consumer on a goroutine, reading messages from the
// consumer. It is the module entrypoint for the dispatcher Kafka consumer,
// conforming to the startModuleFn type definition in config-manager/cmd.
func Start(ctx context.Context, errors chan<- error) {
	consumer := kafka.NewConsumer(config.DefaultConfig.KafkaDispatcherTopic)
	producer := kafka.NewProducer(config.DefaultConfig.KafkaSystemProfileTopic)

	handler := &handler{producer: producer, uuidGenerator: uuid.New}

	start := kafka.NewConsumerEventLoop(ctx, consumer, handler.onMessage, errors)

	go func() {
		defer consumer.Close()
		start()
	}()
}
