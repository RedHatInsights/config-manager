package inventoryconsumer

import (
	"config-manager/application"
	"config-manager/domain"
	"config-manager/domain/message"
	kafkaUtils "config-manager/infrastructure/kafka"
	"context"
	"encoding/json"
	"log"

	kafka "github.com/segmentio/kafka-go"
)

type handler struct {
	ConfigManagerService application.ConfigManagerInterface
}

func (this *handler) onMessage(ctx context.Context, msg kafka.Message) {
	eventType, err := kafkaUtils.GetHeader(msg, "event_type")
	if err != nil {
		log.Println("Error getting header: ", err)
		return
	}

	if eventType == "created" || eventType == "updated" {
		value := &message.InventoryEvent{}

		if err := json.Unmarshal(msg.Value, &value); err != nil {
			log.Println("Couldn't unmarshal inventory event: ", err)
			return
		}

		if value.Host.Reporter == "cloud-connector" {
			accState, err := this.ConfigManagerService.GetAccountState(value.Host.Account)
			if err != nil {
				log.Println("Error retrieving state for account: ", value.Host.Account)
			}

			log.Printf("Cloud-connector inventory message received: %+v", value)

			if value.Host.SystemProfile.RHCState != accState.StateID.String() {
				log.Printf("rhc_state_id %s for client %s does not match current state id %s for account %s. Updating.",
					value.Host.SystemProfile.RHCState, value.Host.SystemProfile.RHCID, accState.StateID.String(), accState.AccountID)
				client := []domain.Host{value.Host}
				responses, err := this.ConfigManagerService.ApplyState(ctx, accState, client)
				if err != nil {
					log.Println("Error applying state: ", err)
				}
				log.Println("Message sent to the dispatcher. Results: ", responses)
			} else {
				log.Printf("rhc_state_id %s for client %s is up to date for account %s. Not updating.",
					value.Host.SystemProfile.RHCState, value.Host.SystemProfile.RHCID, accState.AccountID)
			}
		}
	}
}
