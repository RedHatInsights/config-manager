package inventoryconsumer

import (
	"config-manager/application"
	"config-manager/domain"
	"config-manager/domain/message"
	kafkaUtils "config-manager/infrastructure/kafka"
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
	kafka "github.com/segmentio/kafka-go"
)

// handler is a kafka message handler, designed to handle messages read from a
// platform.inventory.events topic.
type handler struct {
	ConfigManagerService application.ConfigManagerInterface
}

type requestIDkey string

// onMessage is the handler function that is called during the consumer event
// loop. It unmarshals the received message and evaluates whether state should
// be applied for the identified host, applying as needed.
func (this *handler) onMessage(ctx context.Context, msg kafka.Message) {
	log.Trace().Str("module", "inventory-consumer").Msg("received message")
	eventType, err := kafkaUtils.GetHeader(msg, "event_type")
	if err != nil {
		log.Error().Err(err).Msg("error getting event_type")
		return
	}
	log.Trace().Str("module", "inventory-consumer").Msgf("event_type = %v", eventType)

	if eventType == "created" || eventType == "updated" {
		value := &message.InventoryEvent{}

		if err := json.Unmarshal(msg.Value, &value); err != nil {
			log.Error().Str("module", "inventory-consumer").Err(err).Msg("couldn't unmarshal inventory event")
			return
		}

		if value.Host.Reporter == "cloud-connector" {
			if eventType == "created" {
				log.Info().Str("module", "inventory-consumer").Msg("new host detected; setting up for playbook execution")
				messageID, err := this.ConfigManagerService.SetupHost(ctx, value.Host)
				if err != nil {
					log.Error().Str("module", "inventory-consumer").Err(err).Msgf("error setting up host: %v", value.Host)
					return
				}
				log.Info().Str("module", "inventory-consumer").Msgf("Cloud-connector setup host message id: %v", messageID)
			}

			accState, err := this.ConfigManagerService.GetAccountState(value.Host.Account)
			if err != nil {
				log.Error().Str("module", "inventory-consumer").Err(err).Msgf("Error retrieving state for account: %v", value.Host.Account)
				return
			}

			reqID, err := kafkaUtils.GetHeader(msg, "request_id")
			if err != nil {
				log.Error().Str("module", "inventory-consumer").Err(err).Msg("Error getting request_id header")
				k := requestIDkey("request_id")
				reqID = uuid.New().String()
				log.Info().Str("module", "inventory-consumer").Msgf("Creating new request_id and adding to context: %v", reqID)
				ctx = context.WithValue(ctx, k, reqID)
			}

			log.Info().Str("module", "inventory-consumer").Msgf("Cloud-connector inventory event request_id: %s, data: %+v", reqID, value)

			if value.Host.SystemProfile.RHCState != accState.StateID.String() {
				log.Info().Str("module", "inventory-consumer").Msgf("rhc_state_id %s for client %s does not match current state id %s for account %s. Updating.",
					value.Host.SystemProfile.RHCState, value.Host.SystemProfile.RHCID, accState.StateID.String(), accState.AccountID)
				client := []domain.Host{value.Host}
				responses, err := this.ConfigManagerService.ApplyState(ctx, accState, client)
				if err != nil {
					log.Error().Str("module", "inventory-consumer").Err(err).Msg("error applying state")
				}
				log.Info().Str("module", "inventory-consumer").Msgf("Message sent to the dispatcher. Results: %v", responses)
			} else {
				log.Info().Str("module", "inventory-consumer").Msgf("rhc_state_id %s for client %s is up to date for account %s. Not updating.",
					value.Host.SystemProfile.RHCState, value.Host.SystemProfile.RHCID, accState.AccountID)
			}
		}
	}
}
