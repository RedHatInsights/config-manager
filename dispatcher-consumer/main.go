package dispatcherconsumer

import (
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

	handler := &handler{}

	start := kafka.NewConsumerEventLoop(ctx, consumer, handler.onMessage, errors)

	go func() {
		defer consumer.Close()
		start()
	}()
}
