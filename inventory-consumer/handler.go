package inventoryconsumer

import (
	"config-manager/application"
	"config-manager/domain/message"
	"context"
	"encoding/json"
	"fmt"

	kafka "github.com/segmentio/kafka-go"
)

type handler struct {
	ConfigManagerService *application.ConfigManagerService
}

func (this *handler) onMessage(ctx context.Context, msg kafka.Message) {
	value := &message.InventoryEvent{}

	if err := json.Unmarshal(msg.Value, &value); err != nil {
		fmt.Println("Couldn't unmarshal inventory event: ", err)
		return
	}

	fmt.Printf("Unmarshalled Message: %+v\n", value)

	if value.Host.SystemProfile.RHCID != "" {
		accState, err := this.ConfigManagerService.GetAccountState(value.Host.Account)
		if err != nil {
			fmt.Println("Error retrieving state for account: ", value.Host.Account)
		}

		clientSlice := []string{value.Host.SystemProfile.RHCID}

		switch value.Type {
		case "created":
			responses, err := this.ConfigManagerService.ApplyState(ctx, accState, clientSlice)
			if err != nil {
				fmt.Println("Error applying state: ", err)
			}
			fmt.Println("Message sent to the dispatcher. Results: ", responses)
		case "updated":
			// TODO: Config-manager needs to update rhc_config_state in inventory before this can be implemented
			// Check rhc_config_state: if not equal to accState.StateID then apply new state
			fmt.Println("Existing RHC client.. checking state")
		default:
			// type "deleted". Remove host reference from state record
			fmt.Println("RHC client removed from inventory")
		}
	}
}
