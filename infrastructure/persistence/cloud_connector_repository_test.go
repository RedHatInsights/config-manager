package persistence_test

import (
	"config-manager/infrastructure/persistence"
	"config-manager/utils"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConnectionsSuccess(t *testing.T) {
	response := `{
		"connections": ["3d711f8b-77d0-4ed5-a5b5-1d282bf930c7", "74368f32-4e6d-4ea2-9b8f-22dac89f9ae4"]
	}`

	connector := &persistence.CloudConnectorClient{
		CloudConnectorHost:     "test",
		CloudConnectorClientID: "test",
		CloudConnectorPSK:      "test",
		Client:                 utils.SetupMockHTTPClient(response, 200),
	}

	results, err := connector.GetConnections(context.Background(), "0000001")

	assert.Nil(t, err)
	assert.Equal(t, len(results), 2, "there should be two connections returned")

	assert.Equal(t, results[0], "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7", "the id from the response should be included")
	assert.Equal(t, results[1], "74368f32-4e6d-4ea2-9b8f-22dac89f9ae4", "The id from the response should be included")
}

func TestGetConnectionsAccountNotFound(t *testing.T) {
	response := `{
		"connections": []
	}`

	connector := &persistence.CloudConnectorClient{
		CloudConnectorHost:     "test",
		CloudConnectorClientID: "test",
		CloudConnectorPSK:      "test",
		Client:                 utils.SetupMockHTTPClient(response, 200),
	}

	results, err := connector.GetConnections(context.Background(), "0000001")

	assert.Nil(t, err)
	assert.Equal(t, len(results), 0, "results should exist, but there should be no connections")
}
