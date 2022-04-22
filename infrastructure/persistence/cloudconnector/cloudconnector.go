package cloudconnector

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

//go:generate oapi-codegen -generate client,types -package cloudconnector -o ./cloudconnector.gen.go https://github.com/RedHatInsights/cloud-connector/raw/f7b64dc76271a2293518c2da513676aa979febfd/internal/controller/api/api.spec.json

// CloundConnectorClient is an abstraction of the REST client API methods to
// interact with the platform cloud-connector application.
type CloudConnectorClient interface {
	GetConnections(ctx context.Context, accountID string) ([]string, error)
}

// cloudConnectorClientImpl implements the CloudConnectorClient interface by
// embedding a ClientWithResponses struct and calling its API methods.
type cloudConnectorClientImpl struct {
	ClientWithResponses
}

// NewCloudConnectorClientWithDoer returns a CloudConnectorClient by
// constructing a cloudconnector.Client, configured with request headers and
// host information.
func NewCloudConnectorClientWithDoer(cfg *viper.Viper, doer HttpRequestDoer) (CloudConnectorClient, error) {
	client, err := NewClientWithResponses(cfg.GetString("Cloud_Connector_Host"), WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("x-rh-cloud-connector-client-id", cfg.GetString("Cloud_Connector_Client_ID"))
		req.Header.Set("x-rh-cloud-connector-psk", cfg.GetString("Cloud_Connector_PSK"))
		return nil
	}), WithHTTPClient(doer))
	if err != nil {
		return nil, fmt.Errorf("cannot create client: %v", err)
	}

	return &cloudConnectorClientImpl{ClientWithResponses: *client}, nil
}

// NewCloudConnectorClient creates a new CloudConnectorClient.
func NewCloudConnectorClient(cfg *viper.Viper) (CloudConnectorClient, error) {
	httpClient := &http.Client{
		Timeout: time.Duration(int(time.Second) * cfg.GetInt("Cloud_Connector_Timeout")),
	}

	return NewCloudConnectorClientWithDoer(cfg, httpClient)
}

// GetConnections calls the GetConnectionAccount API method and formats the
// response.
func (c *cloudConnectorClientImpl) GetConnections(ctx context.Context, accountID string) ([]string, error) {
	resp, err := c.GetConnectionAccount(ctx, AccountID(accountID), func(ctx context.Context, req *http.Request) error {
		req.Header.Set("x-rh-cloud-connector-account", accountID)
		return nil
	})
	if err != nil {
		return nil, err
	}
	response, err := ParseGetConnectionAccountResponse(resp)
	if err != nil {
		return nil, err
	}

	if response.JSON200 != nil {
		return *response.JSON200.Connections, nil
	}

	return nil, nil
}
