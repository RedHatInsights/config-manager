package inventoryconsumer

import (
	"context"
	"fmt"

	kafka "github.com/segmentio/kafka-go"
)

type handler struct {
}

func (this *handler) onMessage(ctx context.Context, msg kafka.Message) {
	fmt.Println("Message: ", msg.Value)
}
