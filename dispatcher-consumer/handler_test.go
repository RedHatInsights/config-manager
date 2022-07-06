package dispatcherconsumer

import (
	"config-manager/domain/message"
	kafkaUtils "config-manager/infrastructure/kafka"
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func newKafkaMessage(t *testing.T, headers []kafka.Header, data []byte) kafka.Message {
	return kafka.Message{
		Headers: headers,
		Value:   data,
	}
}

func newKafkaHeaders(eventService string) []kafka.Header {
	var headers []kafka.Header

	headers = append(headers, kafka.Header{
		Key:   "service",
		Value: []byte(eventService),
	})

	return headers
}

var tests = []struct {
	name         string
	eventService string
	requestID    string
	data         []byte
	org_id       string
	stateID      string
	invMsgSent   bool
	validEvent   bool
}{
	{
		"dispatcher event: running",
		"config_manager",
		"acc2e229-5e26-4985-bea5-c47ce467ea2f",
		[]byte(`{
			"event_type": "update",
			"payload": {
				"id": "1234",
				"org_id": "5318290",
				"recipient": "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7",
				"correlation_id": "62156f8e-9dfd-4103-a60d-31f6090a3241",
				"service": "config_manager",
				"labels": {
					"state_id": "88d2706a-a9da-4aa4-a6fd-9750bcb4714f",
					"id": "e554c871-9d1d-41c5-acda-229180facf0d"
				},
				"status": "running"
			}
		}`),
		"5318290",
		"88d2706a-a9da-4aa4-a6fd-9750bcb4714f",
		false,
		true,
	},
	{
		"dispatcher event: success",
		"config_manager",
		"acc2e229-5e26-4985-bea5-c47ce467ea2f",
		[]byte(`{
			"event_type": "update",
			"payload": {
				"id": "1234",
				"org_id": "5318290",
				"recipient": "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7",
				"correlation_id": "62156f8e-9dfd-4103-a60d-31f6090a3241",
				"service": "config_manager",
				"labels": {
					"state_id": "88d2706a-a9da-4aa4-a6fd-9750bcb4714f",
					"id": "e554c871-9d1d-41c5-acda-229180facf0d"
				},
				"status": "success"
			}
		}`),
		"5318290",
		"88d2706a-a9da-4aa4-a6fd-9750bcb4714f",
		true,
		true,
	},
	{
		"dispatcher event: failed",
		"config_manager",
		"acc2e229-5e26-4985-bea5-c47ce467ea2f",
		[]byte(`{
			"event_type": "update",
			"payload": {
				"id": "1234",
				"org_id": "5318290",
				"recipient": "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7",
				"correlation_id": "62156f8e-9dfd-4103-a60d-31f6090a3241",
				"service": "config_manager",
				"labels": {
					"state_id": "88d2706a-a9da-4aa4-a6fd-9750bcb4714f",
					"id": "e554c871-9d1d-41c5-acda-229180facf0d"
				},
				"status": "failure"
			}
		}`),
		"5318290",
		"88d2706a-a9da-4aa4-a6fd-9750bcb4714f",
		false,
		true,
	},
	{
		"dispatcher event: timeout",
		"remediations",
		"acc2e229-5e26-4985-bea5-c47ce467ea2f",
		[]byte(`{
			"event_type": "update",
			"payload": {
				"id": "1234",
				"org_id": "5318290",
				"recipient": "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7",
				"correlation_id": "62156f8e-9dfd-4103-a60d-31f6090a3241",
				"service": "config_manager",
				"labels": {
					"state_id": "88d2706a-a9da-4aa4-a6fd-9750bcb4714f",
					"id": "e554c871-9d1d-41c5-acda-229180facf0d"
				},
				"status": "timeout"
			}
		}`),
		"5318290",
		"88d2706a-a9da-4aa4-a6fd-9750bcb4714f",
		false,
		false,
	},
}

func TestDispatcherMessageHandler(t *testing.T) {
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var value = &message.DispatcherEvent{}
			err := json.Unmarshal(tt.data, value)
			assert.Nil(t, err)

			reqID, err := uuid.Parse(tt.requestID)
			assert.Nil(t, err)

			producerMsg, err := buildMessage(value.Payload, reqID)
			assert.Nil(t, err)

			var mockProducer = new(kafkaUtils.MockWriter)
			mockProducer.On(
				"WriteMessages",
				ctx,
				[]kafka.Message{
					kafka.Message{
						Key:   []byte("cm-" + value.Payload.Labels["id"]),
						Value: producerMsg,
					},
				},
			).Return(nil)

			handler := &handler{
				producer: mockProducer,
				uuidGenerator: func() uuid.UUID {
					mockUUID, _ := uuid.Parse(tt.requestID)
					return mockUUID
				},
			}

			handler.onMessage(ctx, newKafkaMessage(t, newKafkaHeaders(tt.eventService), tt.data))

			if tt.validEvent {
				if tt.invMsgSent {
					mockProducer.AssertCalled(
						t,
						"WriteMessages",
						ctx,
						[]kafka.Message{
							kafka.Message{
								Key:   []byte("cm-" + value.Payload.Labels["id"]),
								Value: producerMsg,
							},
						},
					)
				} else {
					mockProducer.AssertNotCalled(t, "WriteMessages")
				}
			} else {
				mockProducer.AssertNotCalled(t, "WriteMessages")
			}
		})
	}
}

func TestDispatcherMessageBuilder(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var value = &message.DispatcherEvent{}
			err := json.Unmarshal(tt.data, value)
			assert.Nil(t, err)

			reqID, err := uuid.Parse(tt.requestID)
			assert.Nil(t, err)

			producerMsg, err := buildMessage(value.Payload, reqID)
			assert.Nil(t, err)

			var invMsg = &message.InventoryUpdate{}
			err = json.Unmarshal(producerMsg, invMsg)
			assert.Nil(t, err)

			assert.Equal(t, invMsg.Metadata.RequestID, tt.requestID)
			assert.Equal(t, invMsg.Data.OrgID, tt.org_id)
			assert.Equal(t, invMsg.Data.ID, value.Payload.Labels["id"])
			assert.Equal(t, invMsg.Data.SystemProfile.RHCState, tt.stateID)
		})
	}
}
