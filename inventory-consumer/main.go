package inventoryconsumer

import (
	"config-manager/infrastructure"
	"config-manager/infrastructure/kafka"
	"context"

	"github.com/spf13/viper"
)

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
