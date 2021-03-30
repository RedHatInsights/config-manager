package inventoryconsumer

import (
	"config-manager/infrastructure/kafka"
	"context"
	"fmt"

	"github.com/spf13/viper"
)

func Start(
	ctx context.Context,
	cfg *viper.Viper,
	errors chan<- error,
) {
	consumer, err := kafka.NewConsumer(cfg, cfg.GetString("Kafka_Inventory_Topic"))
	if err != nil {
		fmt.Println("Error during inventory consumer setup")
		errors <- err
	}

	handler := &handler{}

	start := kafka.NewConsumerEventLoop(ctx, consumer, handler.onMessage, errors)

	go func() {
		defer consumer.Close()
		start()
	}()
}
