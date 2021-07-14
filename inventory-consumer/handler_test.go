package inventoryconsumer

import (
	"config-manager/application"
	"config-manager/domain"
	"config-manager/domain/message"
	"config-manager/infrastructure/persistence/dispatcher"
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func newKafkaMessage(t *testing.T, headers []kafka.Header, data []byte) kafka.Message {
	return kafka.Message{
		Headers: headers,
		Value:   data,
	}
}

func newKafkaHeaders(eventType string) []kafka.Header {
	var headers []kafka.Header

	headers = append(headers, kafka.Header{
		Key:   "event_type",
		Value: []byte(eventType),
	})

	headers = append(headers, kafka.Header{
		Key:   "request_id",
		Value: []byte("01"),
	})

	return headers
}

var tests = []struct {
	name         string
	eventType    string
	data         []byte
	account      string
	currentState string
	validEvent   bool
}{
	{
		"cloud-connector event: created",
		"created",
		[]byte(`{
			"type": "created",
			"host": {
				"id": "1234",
				"account": "0000001",
				"reporter": "cloud-connector",
				"system_profile": {
					"rhc_client_id": "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7"
				}
			}
		}`),
		"0000001",
		"74368f32-4e6d-4ea2-9b8f-22dac89f9ae4",
		true,
	},
	{
		"cloud-connector event: updated",
		"updated",
		[]byte(`{
			"type": "updated",
			"host": {
				"id": "1234",
				"account": "0000002",
				"reporter": "cloud-connector",
				"system_profile": {
					"rhc_client_id": "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7",
					"rhc_config_state": "74368f32-4e6d-4ea2-9b8f-22dac89f9ae4"
				}
			}
		}`),
		"0000002",
		"74368f32-4e6d-4ea2-9b8f-22dac89f9ae4",
		true,
	},
	{
		"cloud-connector event: delete",
		"delete",
		[]byte(`{
			"type": "delete",
			"id": "1234",
			"account": "0000001"
		}`),
		"0000001",
		"74368f32-4e6d-4ea2-9b8f-22dac89f9ae4",
		false,
	},
	{
		"other reporter event: updated",
		"updated",
		[]byte(`{
			"type": "updated",
			"host": {
				"id": "1234",
				"account": "0000001",
				"reporter": "other",
				"system_profile": {
					"rhc_client_id": "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7",
					"rhc_config_state": "74368f32-4e6d-4ea2-9b8f-22dac89f9ae4"
				}
			}
		}`),
		"0000001",
		"74368f32-4e6d-4ea2-9b8f-22dac89f9ae4",
		false,
	},
}

func TestInventoryMessageHandler(t *testing.T) {
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmServiceMock = new(application.ConfigManagerServiceMock)

			stateUUID, _ := uuid.Parse(tt.currentState)

			invData := &message.InventoryEvent{}
			_ = json.Unmarshal(tt.data, invData)

			cmServiceMock.On(
				"GetAccountState",
				tt.account,
			).Return(&domain.AccountState{
				AccountID: tt.account,
				StateID:   stateUUID,
			}, nil)

			cmServiceMock.On(
				"ApplyState",
				ctx,
				&domain.AccountState{
					AccountID: tt.account,
					StateID:   stateUUID,
				},
				[]domain.Host{invData.Host},
			).Return([]dispatcher.RunCreated{}, nil)

			handler := &handler{
				ConfigManagerService: cmServiceMock,
			}

			handler.onMessage(ctx, newKafkaMessage(t, newKafkaHeaders(tt.eventType), tt.data))

			if tt.validEvent {
				cmServiceMock.AssertCalled(t, "GetAccountState", tt.account)
				if invData.Host.SystemProfile.RHCState != tt.currentState {
					cmServiceMock.AssertCalled(
						t,
						"ApplyState",
						ctx,
						&domain.AccountState{
							AccountID: tt.account,
							StateID:   stateUUID,
						},
						[]domain.Host{invData.Host},
					)
				} else {
					cmServiceMock.AssertNotCalled(t, "ApplyState")
				}
			} else {
				cmServiceMock.AssertNotCalled(t, "GetAccountState")
				cmServiceMock.AssertNotCalled(t, "ApplyState")
			}
		})
	}
}
