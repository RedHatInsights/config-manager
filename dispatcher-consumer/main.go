package dispatcherconsumer

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
	consumer := kafka.NewConsumer(cfg, cfg.GetString("Kafka_Dispatcher_Topic"))
	producer := kafka.NewProducer(cfg, cfg.GetString("Kafka_System_Profile_Topic"))

	container := infrastructure.Container{Config: cfg}

	cmService := container.CMService()

	handler := &handler{producer: producer, ConfigManagerService: cmService}

	start := kafka.NewConsumerEventLoop(ctx, consumer, handler.onMessage, errors)

	go func() {
		defer consumer.Close()
		start()
	}()
}
