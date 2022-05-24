package dispatcherconsumer

import (
	"config-manager/domain/message"
	kafkaUtils "config-manager/infrastructure/kafka"
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"

	"github.com/google/uuid"

	kafka "github.com/segmentio/kafka-go"
)

// handler is a kafka message handler, designed to handle messages read from a
// platform.playbook-dispatcher.runs topic and produce messages on a
// platform.inventory.system-profile topic.
type handler struct {
	producer      kafkaUtils.KafkaWriterInterface
	uuidGenerator func() uuid.UUID
}

// buildMessage creates a message.InventoryUpdate structure, populated from
// values in payload. The message is then marshaled into JSON and returned.
func buildMessage(payload message.DispatcherEventPayload, reqID uuid.UUID) ([]byte, error) {
	// This will change once this information can be consumed from the run_hosts topic.
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
		log.Error().Err(err).Msg("error marshalling inventory update message")
		return nil, err
	}

	return bMsg, err
}

// onMessage is the handler function that is called during the consumer event
// loop.
func (this *handler) onMessage(ctx context.Context, msg kafka.Message) {
	eventService, err := kafkaUtils.GetHeader(msg, "service")
	if err != nil {
		log.Error().Err(err).Msgf("Error getting header: ", err)
		return
	}

	if eventService == "config_manager" {
		value := &message.DispatcherEvent{}

		if err := json.Unmarshal(msg.Value, value); err != nil {
			log.Error().Err(err).Msgf("Couldn't unmarshal dispatcher event: ", err)
			return
		}

		switch status := value.Payload.Status; status {
		case "success":
			log.Info().Msgf("Received success event for host ", value.Payload.Recipient)
			log.Printf("Message payload: %+v", value.Payload)

			reqID := this.uuidGenerator()
			updateMsg, err := buildMessage(value.Payload, reqID)
			if err != nil {
				log.Error().Err(err).Msgf("Error building message for inventory update: ", err)
				break
			}

			err = this.producer.WriteMessages(ctx,
				kafka.Message{
					Key:   []byte("cm-" + value.Payload.Labels["id"]),
					Value: updateMsg,
				},
			)
			if err != nil {
				log.Info().Msgf("Error producing message to system profile topic. request_id: ", reqID.String())
			} else {
				log.Printf("Message sent to inventory with request_id: %s, host_id: %s, account: %s",
					reqID.String(), value.Payload.Labels["id"], value.Payload.Account)
			}
		case "running":
			log.Info().Msgf("Received running event for host ", value.Payload.Recipient)
			// TODO anything to do for running?
		default:
			log.Info().Msgf("Received a failure event for host ", value.Payload.Recipient)
			// TODO handle failure/timeout.. retry?
		}
	}
}
