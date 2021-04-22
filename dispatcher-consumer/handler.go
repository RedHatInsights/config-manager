package dispatcherconsumer

import (
	"config-manager/domain/message"
	kafkaUtils "config-manager/infrastructure/kafka"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"

	kafka "github.com/segmentio/kafka-go"
)

type handler struct {
	producer      kafkaUtils.KafkaWriterInterface
	uuidGenerator func() uuid.UUID
}

// This message to inventory is constructed using data from provided labels. This will change
// once this information can be consumed from the run_hosts topic.
func buildMessage(payload message.DispatcherEventPayload, reqID uuid.UUID) ([]byte, error) {
	msg := message.InventoryUpdate{
		Operation: "add_host",
		Metadata:  message.PlatformMetadata{RequestID: reqID.String()},
		Data: message.HostUpdateData{
			ID:      payload.Labels["id"],
			Account: payload.Account,
			SystemProfile: message.HostUpdateSystemProfile{
				RHCState: payload.Labels["state_id"],
			},
		},
	}

	bMsg, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshalling inventory update message")
		return nil, err
	}

	return bMsg, err
}

func (this *handler) onMessage(ctx context.Context, msg kafka.Message) {
	eventService, err := kafkaUtils.GetHeader(msg, "service")
	if err != nil {
		log.Println("Error getting header: ", err)
		return
	}

	if eventService == "config_manager" {
		value := &message.DispatcherEvent{}

		if err := json.Unmarshal(msg.Value, value); err != nil {
			log.Println("Couldn't unmarshal dispatcher event: ", err)
			return
		}

		switch status := value.Payload.Status; status {
		case "success":
			log.Println("Received success event for host ", value.Payload.Recipient)
			log.Println(fmt.Sprintf("Message payload: %+v", value.Payload))

			reqID := this.uuidGenerator()
			updateMsg, err := buildMessage(value.Payload, reqID)
			if err != nil {
				log.Println("Error building message for inventory update: ", err)
				break
			}

			err = this.producer.WriteMessages(ctx,
				kafka.Message{
					Key:   []byte("cm-" + value.Payload.Labels["id"]),
					Value: updateMsg,
				},
			)
			if err != nil {
				log.Println("Error producing message to system profile topic. request_id: ", reqID.String())
			} else {
				log.Println(fmt.Sprintf("Message sent to inventory with request_id: %s, host_id: %s, account: %s",
					reqID.String(), value.Payload.Labels["id"], value.Payload.Account))
			}
		case "running":
			log.Println("Received running event for host ", value.Payload.Recipient)
			// TODO anything to do for running?
		default:
			log.Println("Received a failure event for host ", value.Payload.Recipient)
			// TODO handle failure/timeout.. retry?
		}
	}
}
