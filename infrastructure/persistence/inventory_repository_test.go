package persistence_test

import (
	"config-manager/infrastructure/persistence"
	"config-manager/utils"
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var tests = []struct {
	name       string
	response   string
	numResults int
	clientIDs  []string
}{
	{
		"Get inventory clients 1 result",
		`{
			"total": 1,
			"count": 1,
			"page": 1,
			"per_page": 50,
			"results": [
				{
					"id": "1234",
					"account": "0000001",
					"display_name": "test",
					"system_profile": {
						"rhc_client_id": "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7",
						"rhc_config_state": "3ef6c247-d913-491b-b3eb-56315a6e0f84"
					}
				}
			]
		}`,
		1,
		[]string{"3d711f8b-77d0-4ed5-a5b5-1d282bf930c7"},
	},
	{
		"Get inventory clients 2 results",
		`{
			"total": 2,
			"count": 2,
			"page": 1,
			"per_page": 50,
			"results": [
				{
					"id": "1234",
					"account": "0000001",
					"display_name": "test1",
					"system_profile": {
						"rhc_client_id": "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7",
						"rhc_config_state": "3ef6c247-d913-491b-b3eb-56315a6e0f84"
					}
				},
				{
					"id": "5678",
					"account": "0000001",
					"display_name": "test2",
					"system_profile": {
						"rhc_client_id": "b2df3866-cd1c-4b5f-a342-a8dca6a9eb48",
						"rhc_config_state": "3ef6c247-d913-491b-b3eb-56315a6e0f84"
					}
				}
			]
		}`,
		2,
		[]string{"3d711f8b-77d0-4ed5-a5b5-1d282bf930c7", "b2df3866-cd1c-4b5f-a342-a8dca6a9eb48"},
	},
	{
		"Get inventory clients 0 results",
		`{
			"total": 0,
			"count": 0,
			"page": 1,
			"per_page": 50,
			"results": []
		}`,
		0,
		[]string{},
	},
}

func TestGetClients(t *testing.T) {
	id := base64.StdEncoding.EncodeToString([]byte(`{ "identity": {"account_number": "540155", "type": "user", "internal": { "org_id": "1979710" } } }`))
	ctx := context.WithValue(context.Background(), "X-Rh-Identity", id)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invConnector := &persistence.InventoryClient{
				InventoryHost: "test",
				Client:        utils.SetupMockHTTPClient(tt.response, 200),
			}

			results, err := invConnector.GetInventoryClients(ctx, 1)

			assert.Nil(t, err)

			assert.Equal(t, tt.numResults, len(results.Results), fmt.Sprintf("results should have length %d", tt.numResults))

			retrievedIDs := []string{}
			for _, host := range results.Results {
				retrievedIDs = append(retrievedIDs, host.SystemProfile.RHCID)
			}

			assert.Equal(t, tt.clientIDs, retrievedIDs, "the id from the response should be included")
		})
	}
}

// func TestGetClientsSuccess(t *testing.T) {
// 	response := `{
// 		"total": 1,
// 		"count": 1,
// 		"page": 1,
// 		"per_page": 50,
// 		"results": [
// 			{
// 				"id": "1234",
// 				"account": "0000001",
// 				"display_name": "test",
// 				"system_profile": {
// 					"rhc_client_id": "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7",
// 					"rhc_config_state": "3ef6c247-d913-491b-b3eb-56315a6e0f84"
// 				}
// 			}
// 		]
// 	}`

// 	invConnector := &persistence.InventoryClient{
// 		InventoryHost: "test",
// 		Client:        utils.SetupMockHTTPClient(response, 200),
// 	}

// 	id := base64.StdEncoding.EncodeToString([]byte(`{ "identity": {"account_number": "540155", "type": "user", "internal": { "org_id": "1979710" } } }`))
// 	ctx := context.WithValue(context.Background(), "X-Rh-Identity", id)

// 	results, err := invConnector.GetInventoryClients(ctx, 1)

// 	assert.Nil(t, err)
// 	assert.Equal(t, len(results.Results), 1, "there should be one connection returned")

// 	assert.Equal(t, results.Results[0].SystemProfile.RHCID, "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7", "the id from the response should be included")
// 	assert.Equal(t, results.Results[0].SystemProfile.RHCState, "3ef6c247-d913-491b-b3eb-56315a6e0f84", "The id from the response should be included")
// }

// func TestGetClientsNotFound(t *testing.T) {
// 	response := `{
// 		"connections": []
// 	}`

// 	connector := &persistence.CloudConnectorClient{
// 		CloudConnectorHost:     "test",
// 		CloudConnectorClientID: "test",
// 		CloudConnectorPSK:      "test",
// 		Client:                 utils.SetupMockHTTPClient(response, 200),
// 	}

// 	results, err := connector.GetConnections(context.Background(), "0000001")

// 	assert.Nil(t, err)
// 	assert.Equal(t, len(results), 0, "results should exist, but there should be no connections")
// }
