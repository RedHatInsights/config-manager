package persistence

import (
	"config-manager/domain"
	"config-manager/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	headerCloudConnectorClientID = "x-rh-cloud-connector-client-id"
	headerCloudConnectorAccount  = "x-rh-cloud-connector-account"
	headerCloudConnectorPSK      = "x-rh-cloud-connector-psk"
)

type CloudConnectorClient struct {
	CloudConnectorHost     string
	CloudConnectorClientID string
	CloudConnectorPSK      string
	Client                 utils.HTTPClient
}

func (c *CloudConnectorClient) GetConnections(
	ctx context.Context,
	accountID string,
) ([]string, error) {
	fmt.Println("Sending request to cloud connector")

	req, err := http.NewRequestWithContext(ctx, "GET", c.CloudConnectorHost+"/api/cloud-connector/v1/connection/"+accountID, nil)
	if err != nil {
		fmt.Println("Error constructing request to cloud-connector: ", err)
		return nil, err
	}
	req.Header.Set(headerCloudConnectorClientID, c.CloudConnectorClientID)
	req.Header.Set(headerCloudConnectorPSK, c.CloudConnectorPSK)
	req.Header.Set(headerCloudConnectorAccount, accountID)

	res, err := c.Client.Do(req)
	if err != nil {
		fmt.Println("Error during request to cloud-connector: ", err)
		return nil, err
	}
	defer res.Body.Close()

	var cloudConnectorRes domain.CloudConnectorConnections
	err = json.NewDecoder(res.Body).Decode(&cloudConnectorRes)
	return cloudConnectorRes.Connections, err
}
