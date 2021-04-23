package dispatcherconsumer

import (
	"config-manager/infrastructure/kafka"
	"context"

	"github.com/google/uuid"

	"github.com/spf13/viper"
)

func Start(
	ctx context.Context,
	cfg *viper.Viper,
	errors chan<- error,
) {
	consumer := kafka.NewConsumer(cfg, cfg.GetString("Kafka_Dispatcher_Topic"))
	producer := kafka.NewProducer(cfg, cfg.GetString("Kafka_System_Profile_Topic"))

	handler := &handler{producer: producer, uuidGenerator: uuid.New}

	start := kafka.NewConsumerEventLoop(ctx, consumer, handler.onMessage, errors)

	go func() {
		defer consumer.Close()
		start()
	}()
}
