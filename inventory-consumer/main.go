package inventoryconsumer

import (
	"config-manager/infrastructure"
	"config-manager/infrastructure/kafka"
	"config-manager/internal/config"
	"context"
)

// Start creates a new Kafka consumer, sets up a message handler, and starts
// running the consumer on a goroutine, reading messages from the consumer. It
// the module entrypoint for the inventory Kafka consumer, conforming to the
// startModuleFn type definition in config-manager/cmd.
func Start(ctx context.Context, errors chan<- error) {
	consumer := kafka.NewConsumer(config.DefaultConfig.KafkaInventoryTopic)

	container := infrastructure.Container{}

	cmService := container.CMService()

	handler := &handler{ConfigManagerService: cmService, Cfg: cfg, DB: container.Database()}

	start := kafka.NewConsumerEventLoop(ctx, consumer, handler.onMessage, errors)

	go func() {
		defer consumer.Close()
		start()
	}()
}
