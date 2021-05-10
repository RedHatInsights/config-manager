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

			client := []domain.Host{value.Host}

			// TODO: Switch on event type. Once config-manager is updating rhc_config_state in inventory
			// a check can be made on the rhc_config_state id to determine if work should be done.
			responses, err := this.ConfigManagerService.ApplyState(ctx, accState, client)
			if err != nil {
				log.Println("Error applying state: ", err)
			}
			log.Println("Message sent to the dispatcher. Results: ", responses)
		}
	}
}
