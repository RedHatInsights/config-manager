package cloudconnector

import "context"

type cloudConnectorClientMock struct{}

func NewCloudConnectorClientMock() CloudConnectorClient {
	return &cloudConnectorClientMock{}
}

func (c *cloudConnectorClientMock) GetConnections(ctx context.Context, accountID string) ([]string, error) {
	return []string{"3d711f8b-77d0-4ed5-a5b5-1d282bf930c7", "74368f32-4e6d-4ea2-9b8f-22dac89f9ae4"}, nil
}
