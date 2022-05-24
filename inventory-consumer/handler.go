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
	eventType, err := kafkaUtils.GetHeader(msg, "event_type")
	if err != nil {
		log.Error().Err(err).Msgf("Error getting event_type: ", err)
		return
	}

	if eventType == "created" || eventType == "updated" {
		value := &message.InventoryEvent{}

		if err := json.Unmarshal(msg.Value, &value); err != nil {
			log.Error().Err(err).Msgf("Couldn't unmarshal inventory event: ", err)
			return
		}

		if value.Host.Reporter == "cloud-connector" {
			if eventType == "created" {
				log.Printf("New host detected; setting up for playbook execution")
				messageID, err := this.ConfigManagerService.SetupHost(ctx, value.Host)
				if err != nil {
					log.Printf("Error setting up host: %v: %v", value.Host, err)
					return
				}
				log.Printf("Cloud-connector setup host message id: %v", messageID)
			}

			accState, err := this.ConfigManagerService.GetAccountState(value.Host.Account)
			if err != nil {
				log.Printf("Error retrieving state for account: %v: %v", value.Host.Account, err)
				return
			}

			reqID, err := kafkaUtils.GetHeader(msg, "request_id")
			if err != nil {
				log.Error().Err(err).Msgf("Error getting request_id: ", err)
				k := requestIDkey("request_id")
				reqID = uuid.New().String()
				log.Info().Msgf("Creating new request_id and adding to context: ", reqID)
				ctx = context.WithValue(ctx, k, reqID)
			}

			log.Printf("Cloud-connector inventory event request_id: %s, data: %+v", reqID, value)

			if value.Host.SystemProfile.RHCState != accState.StateID.String() {
				log.Printf("rhc_state_id %s for client %s does not match current state id %s for account %s. Updating.",
					value.Host.SystemProfile.RHCState, value.Host.SystemProfile.RHCID, accState.StateID.String(), accState.AccountID)
				client := []domain.Host{value.Host}
				responses, err := this.ConfigManagerService.ApplyState(ctx, accState, client)
				if err != nil {
					log.Error().Err(err).Msgf("Error applying state: ", err)
				}
				log.Info().Msgf("Message sent to the dispatcher. Results: ", responses)
			} else {
				log.Printf("rhc_state_id %s for client %s is up to date for account %s. Not updating.",
					value.Host.SystemProfile.RHCState, value.Host.SystemProfile.RHCID, accState.AccountID)
			}
		}
	}
}
