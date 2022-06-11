package cloudconnector

import (
	"config-manager/internal/config"
	"config-manager/internal/url"
	"config-manager/utils"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConnectionsSuccess(t *testing.T) {
	response := `{
		"connections": ["3d711f8b-77d0-4ed5-a5b5-1d282bf930c7", "74368f32-4e6d-4ea2-9b8f-22dac89f9ae4"]
	}`

	config.DefaultConfig.CloudConnectorHost.Value = url.MustParse("http://test")
	config.DefaultConfig.CloudConnectorClientID = "test"
	config.DefaultConfig.CloudConnectorPSK = "test"

	connector, err := NewCloudConnectorClientWithDoer(utils.SetupMockHTTPClient(response, 200))
	if err != nil {
		t.Error(err)
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

	config.DefaultConfig.CloudConnectorHost.Value = url.MustParse("http://test")
	config.DefaultConfig.CloudConnectorClientID = "test"
	config.DefaultConfig.CloudConnectorPSK = "test"

	connector, err := NewCloudConnectorClientWithDoer(utils.SetupMockHTTPClient(response, 200))
	if err != nil {
		t.Error(err)
	}

	results, err := connector.GetConnections(context.Background(), "0000001")

	assert.Nil(t, err)
	assert.Equal(t, len(results), 0, "results should exist, but there should be no connections")
}

func TestSendMessageSuccess(t *testing.T) {
	response := `{"id": "0afbfb55-a2af-43f2-84da-a0896f03f067"}`

	config.DefaultConfig.CloudConnectorHost.Value = url.MustParse("http://test")
	config.DefaultConfig.CloudConnectorClientID = "test"
	config.DefaultConfig.CloudConnectorPSK = "test"

	connector, err := NewCloudConnectorClientWithDoer(utils.SetupMockHTTPClient(response, 201))
	if err != nil {
		t.Error(err)
	}

	results, err := connector.SendMessage(context.Background(), "0000001", "test", []byte(`"test"`), nil, "test")

	assert.Nil(t, err)
	assert.Equal(t, "0afbfb55-a2af-43f2-84da-a0896f03f067", results, "the id from the response should be 0afbfb55-a2af-43f2-84da-a0896f03f067")
}

func TestGetConnectionStatus(t *testing.T) {
	response := `{"status":"connected","dispatchers": {"rhc-worker-playbook": {}}}`

	config.DefaultConfig.CloudConnectorHost.Value = url.MustParse("http://test")
	config.DefaultConfig.CloudConnectorClientID = "test"
	config.DefaultConfig.CloudConnectorPSK = "test"

	connector, err := NewCloudConnectorClientWithDoer(utils.SetupMockHTTPClient(response, 200))
	if err != nil {
		t.Error(err)
	}

	status, results, err := connector.GetConnectionStatus(context.Background(), "0000001", "test")

	assert.Nil(t, err)
	assert.Equal(t, "connected", status, "status should be connected")
	assert.Equal(t, map[string]interface{}{"rhc-worker-playbook": map[string]interface{}{}}, results, "the dispatchers from the response should contain rhc-worker-playbook")
}
