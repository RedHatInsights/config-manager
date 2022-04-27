package cloudconnector

import "context"

type cloudConnectorClientMock struct{}

func NewCloudConnectorClientMock() CloudConnectorClient {
	return &cloudConnectorClientMock{}
}

func (c *cloudConnectorClientMock) GetConnections(ctx context.Context, accountID string) ([]string, error) {
	return []string{"3d711f8b-77d0-4ed5-a5b5-1d282bf930c7", "74368f32-4e6d-4ea2-9b8f-22dac89f9ae4"}, nil
}

func (c *cloudConnectorClientMock) SendMessage(ctx context.Context, accountID string, directive string, payload []byte, metadata map[string]string, recipient string) (string, error) {
	return "0afbfb55-a2af-43f2-84da-a0896f03f067", nil
}

func (c *cloudConnectorClientMock) GetConnectionStatus(ctx context.Context, accountID string, recipient string) (string, map[string]interface{}, error) {
	return "connected", map[string]interface{}{
		"rhc-worker-playbook": map[string]interface{}{},
	}, nil
}
