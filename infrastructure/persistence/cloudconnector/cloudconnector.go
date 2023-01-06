package cloudconnector

import (
	"bytes"
	"config-manager/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

//go:generate oapi-codegen -config oapi-codegen.yml https://github.com/RedHatInsights/cloud-connector/raw/f7b64dc76271a2293518c2da513676aa979febfd/internal/controller/api/api.spec.json

// CloundConnectorClient is an abstraction of the REST client API methods to
// interact with the platform cloud-connector application.
type CloudConnectorClient interface {
	GetConnections(ctx context.Context, accountID string) ([]string, error)
	GetConnectionStatus(ctx context.Context, accountID string, recipient string) (string, map[string]interface{}, error)
	SendMessage(ctx context.Context, accountID string, directive string, payload []byte, metadata map[string]string, recipient string) (string, error)
}

// cloudConnectorClientImpl implements the CloudConnectorClient interface by
// embedding a ClientWithResponses struct and calling its API methods.
type cloudConnectorClientImpl struct {
	ClientWithResponses
}

// NewCloudConnectorClientWithDoer returns a CloudConnectorClient by
// constructing a cloudconnector.Client, configured with request headers and
// host information.
func NewCloudConnectorClientWithDoer(doer HttpRequestDoer) (CloudConnectorClient, error) {
	client, err := NewClientWithResponses(config.DefaultConfig.CloudConnectorHost.Value.String(), WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("x-rh-cloud-connector-client-id", config.DefaultConfig.CloudConnectorClientID)
		req.Header.Set("x-rh-cloud-connector-psk", config.DefaultConfig.CloudConnectorPSK)
		req.Header.Set("x-rh-insights-request-id", uuid.New().String())
		return nil
	}), WithHTTPClient(doer))
	if err != nil {
		return nil, fmt.Errorf("cannot create client: %v", err)
	}

	return &cloudConnectorClientImpl{ClientWithResponses: *client}, nil
}

// NewCloudConnectorClient creates a new CloudConnectorClient.
func NewCloudConnectorClient() (CloudConnectorClient, error) {
	httpClient := &http.Client{
		Timeout: time.Duration(int(time.Second) * config.DefaultConfig.CloudConnectorTimeout),
	}

	return NewCloudConnectorClientWithDoer(httpClient)
}

// GetConnections calls the GetConnectionAccount API method and formats the
// response.
func (c *cloudConnectorClientImpl) GetConnections(ctx context.Context, accountID string) ([]string, error) {
	logger := log.With().Str("http_client", "cloud-connector").Logger()

	resp, err := c.GetConnectionAccount(ctx, AccountID(accountID), func(ctx context.Context, req *http.Request) error {
		req.Header.Set("x-rh-cloud-connector-account", accountID)
		logger.Debug().Str("method", req.Method).Str("url", req.URL.String()).Interface("headers", req.Header).Msg("sending HTTP request")
		return nil
	})
	if err != nil {
		logger.Error().Err(err).Msg("cannot get connections from cloud-connector")
		return nil, err
	}
	logger.Debug().Str("http_status", http.StatusText(resp.StatusCode)).Interface("headers", resp.Header).Msg("recieved HTTP response from cloud-connector")
	response, err := ParseGetConnectionAccountResponse(resp)
	logger.Debug().Str("response", string(response.Body)).Msg("parsed HTTP response")
	if err != nil {
		logger.Error().Err(err).Msg("cannot parse get connection account response")
		return nil, err
	}

	if response.JSON200 != nil {
		return *response.JSON200.Connections, nil
	}

	return nil, nil
}

func (c *cloudConnectorClientImpl) SendMessage(ctx context.Context, accountID string, directive string, payload []byte, metadata map[string]string, recipient string) (string, error) {
	logger := log.With().Str("http_client", "cloud-connector").Logger()

	body := struct {
		Account   string            `json:"account"`
		Directive string            `json:"directive"`
		Metadata  map[string]string `json:"metadata"`
		Payload   json.RawMessage   `json:"payload"`
		Recipient string            `json:"recipient"`
	}{
		Account:   accountID,
		Directive: directive,
		Metadata:  metadata,
		Payload:   payload,
		Recipient: recipient,
	}
	data, err := json.Marshal(body)
	if err != nil {
		logger.Error().Err(err).Msg("cannot marshal JSON body")
		return "", err
	}
	// Using PostMessageWithBody here because the MessageRequest is incorrectly
	// typed as *string rather than interface{}. Perhaps a newer version of the
	// OpenAPI spec will correct that, and PostMessage with a MessageRequest
	// struct can be used instead.
	resp, err := c.PostMessageWithBody(ctx, "application/json", bytes.NewReader(data), func(ctx context.Context, req *http.Request) error {
		req.Header.Set("x-rh-cloud-connector-account", accountID)
		logger.Trace().Str("method", req.Method).Str("url", req.URL.String()).Interface("headers", req.Header).Msg("sending HTTP request")
		return nil
	})
	if err != nil {
		logger.Error().Err(err).Msg("cannot post message")
		return "", err
	}
	logger.Trace().Str("http_status", http.StatusText(resp.StatusCode)).Interface("headers", resp.Header).Msg("received HTTP response")
	response, err := ParsePostMessageResponse(resp)
	if err != nil {
		logger.Error().Err(err).Msg("cannot parse post message response")
		return "", err
	}
	logger.Debug().Str("response", string(response.Body)).Msg("parsed HTTP response")

	if response.JSON201 != nil {
		return response.JSON201.Id.String(), nil
	}

	return "", nil
}

func (c *cloudConnectorClientImpl) GetConnectionStatus(ctx context.Context, accountID string, recipient string) (string, map[string]interface{}, error) {
	logger := log.With().Str("http_client", "cloud-connector").Logger()

	body := ConnectionStatusRequest{
		Account: &accountID,
		NodeId:  &recipient,
	}
	resp, err := c.PostConnectionStatus(ctx, PostConnectionStatusJSONRequestBody(body), func(ctx context.Context, req *http.Request) error {
		req.Header.Set("x-rh-cloud-connector-account", accountID)
		logger = logger.With().Str("method", req.Method).Str("url", req.URL.String()).Interface("headers", req.Header).Interface("body", body).Logger()
		logger.Trace().Msg("sending HTTP request")
		return nil
	})
	if err != nil {
		logger.Error().Err(err).Msg("cannot get connection status")
		return "unknown", nil, err
	}
	logger = logger.With().Str("http_status", http.StatusText(resp.StatusCode)).Interface("headers", resp.Header).Logger()
	response, err := ParsePostConnectionStatusResponse(resp)
	if err != nil {
		logger.Error().Err(err).Msg("cannot parse connection status response")
		return "unknown", nil, err
	}
	logger = logger.With().Str("response", string(response.Body)).Logger()
	logger.Trace().Msg("received HTTP response")

	if response.JSON200 != nil {
		var status string
		var dispatchers map[string]interface{}

		if response.JSON200.Status != nil {
			status = string(*response.JSON200.Status)
		}

		if response.JSON200.Dispatchers != nil {
			dispatchers = map[string]interface{}(*response.JSON200.Dispatchers)
		}

		return status, dispatchers, nil
	}

	return "", map[string]interface{}{}, fmt.Errorf("unknown connection status: %v", fmt.Errorf("%v: %v", response.Status(), string(response.Body)))
}
