package internal_test

import (
	"config-manager/config"
	"config-manager/infrastructure/persistence/dispatcher/internal"
	dispatcherPublic "config-manager/infrastructure/persistence/dispatcher/public"
	"config-manager/utils"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

var inputs = []internal.RunInput{
	internal.RunInput{
		Recipient: "276c4685-fdfb-4172-930f-4148b8340c2e",
		Account:   "0000001",
		Url:       "https://cloud.redhat.com/api/config-manager/v1/states/e417581a-d649-4cdc-9506-6eb7fdbfd66d/playbook",
		Labels: &dispatcherPublic.Labels{
			AdditionalProperties: map[string]string{
				"test": "e417581a-d649-4cdc-9506-6eb7fdbfd66d",
			},
		},
	},
	internal.RunInput{
		Recipient: "9a76b28b-0e09-41c8-bf01-79d1bef72646",
		Account:   "0000001",
		Url:       "https://cloud.redhat.com/api/config-manager/v1/states/e417581a-d649-4cdc-9506-6eb7fdbfd66d/playbook",
		Labels: &dispatcherPublic.Labels{
			AdditionalProperties: map[string]string{
				"test": "e417581a-d649-4cdc-9506-6eb7fdbfd66d",
			},
		},
	},
}

func TestDispatchSuccess(t *testing.T) {
	response := `[
		{"code": 200, "id": "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7"},
		{"code": 200, "id": "74368f32-4e6d-4ea2-9b8f-22dac89f9ae4"}
	]`

	doer := utils.SetupMockHTTPClient(response, 207)

	dispatcher := internal.NewDispatcherClientWithDoer(config.Get(), doer)

	results, err := dispatcher.Dispatch(context.Background(), inputs)

	assert.Nil(t, err)
	assert.Equal(t, len(*results), 2, "there should be two response objects")

	assert.Equal(t, (*results)[0].Code, 200, "the response code should be 200")
	assert.Equal(t, string(*(*results)[0].Id), "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7", "The id from the response should be included")

	assert.Equal(t, (*results)[1].Code, 200, "the response code should be 200")
	assert.Equal(t, string(*(*results)[1].Id), "74368f32-4e6d-4ea2-9b8f-22dac89f9ae4", "The id from the response should be included")
}

func TestDispatchNotFound(t *testing.T) {
	response := `[
		{"code": 404},
		{"code": 404}
	]`

	doer := utils.SetupMockHTTPClient(response, 207)

	dispatcher := internal.NewDispatcherClientWithDoer(config.Get(), doer)

	results, err := dispatcher.Dispatch(context.Background(), inputs)

	assert.Nil(t, err)
	assert.Equal(t, len(*results), 2, "there should be two response objects")

	assert.Equal(t, (*results)[0].Code, 404, "the response code should be 404")
	assert.Nil(t, (*results)[0].Id)

	assert.Equal(t, (*results)[1].Code, 404, "the response code should be 404")
	assert.Nil(t, (*results)[1].Id)
}
