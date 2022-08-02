package dispatcherconsumer

import (
	"config-manager/internal/config"
	"config-manager/internal/util"
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

var Command ffcli.Command = ffcli.Command{
	Name:      "dispatcher-consumer",
	ShortHelp: "Run the dispatcher kafka consumer",
	LongHelp:  "Consumes message from the 'kafka-dispatcher-topic' and produces messages on the 'kafka-system-profile' topic.",
	Exec: func(ctx context.Context, args []string) error {
		log.Info().Str("command", "dispatcher-consumer").Msg("starting command")

		reader := util.Kafka.NewReader(config.DefaultConfig.KafkaDispatcherTopic)
		writer := util.Kafka.NewWriter(config.DefaultConfig.KafkaSystemProfileTopic)

		for {
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Error().Err(err).Msg("unable to read message")
				continue
			}
			go handler(ctx, writer, m)
		}
	},
}

// DispatcherEvent represents a message read off the playbook-dispatcher.runs
// topic.
type DispatcherEvent struct {
	Type    string `json:"event_type"`
	Payload struct {
		ID            string            `json:"id"`
		OrgID         string            `json:"org_id"`
		Recipient     string            `json:"recipient"`
		CorrelationID string            `json:"correlation_id"`
		Service       string            `json:"service"`
		URL           string            `json:"url"`
		Labels        map[string]string `json:"labels"`
		Status        string            `json:"status"`
	} `json:"payload"`
}

// InventoryUpdate represents a message written to the inventory.system-profile
// topic.
type InventoryUpdate struct {
	Operation        string `json:"operation"`
	PlatformMetadata struct {
		RequestID string `json:"request_id"`
	} `json:"platform_metadata"`
	Data struct {
		ID            string `json:"id"`
		OrgID         string `json:"org_id"`
		SystemProfile struct {
			RHCConfigState string `json:"rhc_config_state"`
		} `json:"system_profile"`
	} `json:"data"`
}

func handler(ctx context.Context, writer *kafka.Writer, msg kafka.Message) {
	eventService, err := util.Kafka.GetHeader(msg, "service")
	if err != nil {
		log.Error().Err(err).Msg("error getting header")
		return
	}

	if eventService == "config_manager" {
		value := &DispatcherEvent{}

		if err := json.Unmarshal(msg.Value, value); err != nil {
			log.Error().Err(err).Msg("couldn't unmarshal dispatcher event")
			return
		}

		switch status := value.Payload.Status; status {
		case "success":
			log.Info().Msgf("Received success event for host %v", value.Payload.Recipient)
			log.Info().Msgf("Message payload: %+v", value.Payload)

			reqID := uuid.New()
			newMessage := InventoryUpdate{
				Operation: "",
				PlatformMetadata: struct {
					RequestID string `json:"request_id"`
				}{
					RequestID: reqID.String(),
				},
				Data: struct {
					ID            string `json:"id"`
					OrgID         string `json:"org_id"`
					SystemProfile struct {
						RHCConfigState string `json:"rhc_config_state"`
					} `json:"system_profile"`
				}{
					ID:    value.Payload.Labels["id"],
					OrgID: value.Payload.OrgID,
					SystemProfile: struct {
						RHCConfigState string `json:"rhc_config_state"`
					}{
						RHCConfigState: value.Payload.Labels["state_id"],
					},
				},
			}
			data, err := json.Marshal(newMessage)
			if err != nil {
				log.Error().Err(err).Msg("cannot marshal json")
				return
			}
			err = writer.WriteMessages(ctx,
				kafka.Message{
					Key:   []byte("cm-" + value.Payload.Labels["id"]),
					Value: data,
				},
			)
			if err != nil {
				log.Info().Msgf("Error producing message to system profile topic. request_id: %v", reqID.String())
			} else {
				log.Info().Msgf("Message sent to inventory with request_id: %s, host_id: %s, org_id: %s",
					reqID.String(), value.Payload.Labels["id"], value.Payload.OrgID)
			}
		case "running":
			log.Info().Msgf("Received running event for host %v", value.Payload.Recipient)
			// TODO anything to do for running?
		default:
			log.Info().Msgf("Received a failure event for host %v", value.Payload.Recipient)
			// TODO handle failure/timeout.. retry?
		}
	}
}
