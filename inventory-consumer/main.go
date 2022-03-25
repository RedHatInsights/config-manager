package inventoryconsumer

import (
	"config-manager/infrastructure"
	"config-manager/infrastructure/kafka"
	"context"

	"github.com/spf13/viper"
)

// Start creates a new Kafka consumer, sets up a message handler, and starts
// running the consumer on a goroutine, reading messages from the consumer. It
// the module entrypoint for the inventory Kafka consumer, conforming to the
// startModuleFn type definition in config-manager/cmd.
func Start(
	ctx context.Context,
	cfg *viper.Viper,
	errors chan<- error,
) {
	consumer := kafka.NewConsumer(cfg, cfg.GetString("Kafka_Inventory_Topic"))

	container := infrastructure.Container{Config: cfg}

	cmService := container.CMService()

	handler := &handler{ConfigManagerService: cmService}

	start := kafka.NewConsumerEventLoop(ctx, consumer, handler.onMessage, errors)

	go func() {
		defer consumer.Close()
		start()
	}()
}
