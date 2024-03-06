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

//go:generate oapi-codegen -config oapi-codegen.yml https://raw.githubusercontent.com/RedHatInsights/cloud-connector/master/internal/controller/api/api.spec.json

// CloundConnectorClient is an abstraction of the REST client API methods to
// interact with the platform cloud-connector application.
type CloudConnectorClient interface {
	GetConnectionStatus(ctx context.Context, orgID string, recipient string) (string, map[string]interface{}, error)
	SendMessage(ctx context.Context, orgID string, directive string, payload []byte, metadata map[string]string, recipient string) (string, error)
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

func (c *cloudConnectorClientImpl) SendMessage(ctx context.Context, orgID string, directive string, payload []byte, metadata map[string]string, recipient string) (string, error) {
	logger := log.With().Str("http_client", "cloud-connector").Logger()

	body := struct {
		Directive string            `json:"directive"`
		Metadata  map[string]string `json:"metadata"`
		Payload   json.RawMessage   `json:"payload"`
	}{
		Directive: directive,
		Metadata:  metadata,
		Payload:   payload,
	}
	data, err := json.Marshal(body)
	if err != nil {
		logger.Error().Err(err).Msg("cannot marshal JSON body")
		return "", err
	}

	resp, err := c.PostV2ConnectionsClientIdMessageWithBody(ctx, recipient, "application/json", bytes.NewReader(data), func(ctx context.Context, req *http.Request) error {
		req.Header.Set("x-rh-cloud-connector-org-id", orgID)
		logger.Trace().Str("method", req.Method).Str("url", req.URL.String()).Interface("headers", req.Header).Msg("sending HTTP request")
		return nil
	})
	if err != nil {
		logger.Error().Err(err).Msg("cannot post message")
		return "", err
	}
	logger.Trace().Str("http_status", http.StatusText(resp.StatusCode)).Interface("headers", resp.Header).Msg("received HTTP response")
	response, err := ParsePostV2ConnectionsClientIdMessageResponse(resp)
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

func (c *cloudConnectorClientImpl) GetConnectionStatus(ctx context.Context, orgID string, recipient string) (string, map[string]interface{}, error) {
	logger := log.With().Str("http_client", "cloud-connector").Logger()

	resp, err := c.V2ConnectionStatusMultiorg(ctx, recipient, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("x-rh-cloud-connector-org-id", orgID)
		logger = logger.With().Str("method", req.Method).Str("url", req.URL.String()).Interface("headers", req.Header).Logger()
		logger.Trace().Msg("sending HTTP request")
		return nil
	})
	if err != nil {
		logger.Error().Err(err).Msg("cannot get connection status")
		return "unknown", nil, err
	}
	logger = logger.With().Str("http_status", http.StatusText(resp.StatusCode)).Interface("headers", resp.Header).Logger()
	response, err := ParseV2ConnectionStatusMultiorgResponse(resp)
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
