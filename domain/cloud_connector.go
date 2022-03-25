package domain

import "context"

// CloudConnectorConnections represents the connections received from the
// cloud-connector application.
type CloudConnectorConnections struct {
	Connections []string `json:"connections"`
}

// CloundConnectorClient is an abstraction of the REST client API methods to
// interact with the platform cloud-connector application.
type CloudConnectorClient interface {
	GetConnections(ctx context.Context, accountID string) ([]string, error)
}
