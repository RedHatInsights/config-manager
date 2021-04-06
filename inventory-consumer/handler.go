package inventoryconsumer

import (
	"config-manager/domain"
	"config-manager/domain/message"
	kafkaUtils "config-manager/infrastructure/kafka"
	"context"
	"encoding/json"
	"fmt"

	kafka "github.com/segmentio/kafka-go"
)

type handler struct {
	ConfigManagerService domain.ConfigManagerInterface
}

func (this *handler) onMessage(ctx context.Context, msg kafka.Message) {
	eventType, err := kafkaUtils.GetHeader(msg, "event_type")
	if err != nil {
		fmt.Println("Error getting header: ", err)
		return
	}

	if eventType == "delete" {
		return
	}

	value := &message.InventoryEvent{}

	if err := json.Unmarshal(msg.Value, &value); err != nil {
		fmt.Println("Couldn't unmarshal inventory event: ", err)
		return
	}

	if value.Host.Reporter == "cloud-connector" {
		accState, err := this.ConfigManagerService.GetAccountState(value.Host.Account)
		if err != nil {
			fmt.Println("Error retrieving state for account: ", value.Host.Account)
		}

		clientSlice := []string{value.Host.SystemProfile.RHCID}

		// TODO: Switch on event type. Once config-manager is updating rhc_config_state in inventory
		// a check can be made on the rhc_config_state id to determine if work should be done.
		responses, err := this.ConfigManagerService.ApplyState(ctx, accState, clientSlice)
		if err != nil {
			fmt.Println("Error applying state: ", err)
		}
		fmt.Println("Message sent to the dispatcher. Results: ", responses)
	}
}
