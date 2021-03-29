package domain

import "context"

type CloudConnectorConnections struct {
	Connections []string `json:"connections"`
}

type CloudConnectorClient interface {
	GetConnections(ctx context.Context, accountID string) ([]string, error)
}
