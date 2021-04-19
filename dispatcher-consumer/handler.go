package dispatcherconsumer

import (
	"config-manager/domain"
	"config-manager/domain/message"
	kafkaUtils "config-manager/infrastructure/kafka"
	"context"
	"encoding/json"
	"fmt"
	"log"

	kafka "github.com/segmentio/kafka-go"
)

type handler struct {
	producer             *kafka.Writer
	ConfigManagerService domain.ConfigManagerInterface
}

func buildMessage(stateID string, payload message.DispatcherEventPayload) ([]byte, error) {
	msg := message.InventoryUpdate{
		Operation: "add_host",
		Data: message.HostUpdateData{
			ID:      payload.Labels["id"],
			Account: payload.Account,
			SystemProfile: message.HostUpdateSystemProfile{
				RHCState: stateID,
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

		if err := json.Unmarshal(msg.Value, &value); err != nil {
			log.Println("Couldn't unmarshal dispatcher event: ", err)
			return
		}

		switch status := value.Payload.Status; status {
		case "success":
			log.Println("Received success event for host ", value.Payload.Recipient)
			log.Println(fmt.Sprintf("Message payload: %+v", value.Payload))

			state, err := this.ConfigManagerService.GetSingleStateChange(value.Payload.Labels["state_id"])
			if err != nil {
				log.Println("Error retrieving the state archive for this run: ", err)
				break
			}

			updateMsg, err := buildMessage(state.StateID.String(), value.Payload)
			if err != nil {
				log.Println("Error building message for inventory update: ", err)
				break
			}

			err = this.producer.WriteMessages(ctx,
				kafka.Message{
					Value: updateMsg,
				},
			)
			if err != nil {
				log.Println("Error producing message to system profile topic")
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
