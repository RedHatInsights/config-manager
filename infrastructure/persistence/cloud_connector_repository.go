package persistence

import (
	"config-manager/domain"
	"config-manager/utils"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	CloudConnectorImpl     string
}

func (c *CloudConnectorClient) GetConnections(
	ctx context.Context,
	accountID string,
) ([]string, error) {
	fmt.Println("Sending request to cloud connector")

	if c.CloudConnectorImpl == "mock" {
		expectedResponse := []byte(`{
			"connections": ["3d711f8b-77d0-4ed5-a5b5-1d282bf930c7", "74368f32-4e6d-4ea2-9b8f-22dac89f9ae4"]
		}`)
		var cloudConnectorRes domain.CloudConnectorConnections
		err := json.Unmarshal(expectedResponse, &cloudConnectorRes)
		return cloudConnectorRes.Connections, err
	}

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
	if err != nil {
		body, _ := ioutil.ReadAll(res.Body)
		log.Println("Error decoding cloud-connector response: ", string(body))
	}
	return cloudConnectorRes.Connections, err
}
