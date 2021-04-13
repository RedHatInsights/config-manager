package persistence

import (
	"bytes"
	"config-manager/domain"
	"config-manager/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type DispatcherClient struct {
	DispatcherHost string
	DispatcherPSK  string
	DispatcherImpl string
	Client         utils.HTTPClient
}

func (dc *DispatcherClient) Dispatch(
	ctx context.Context,
	inputs []domain.DispatcherInput,
) ([]domain.DispatcherResponse, error) {
	fmt.Println("Sending request to playbook dispatcher")

	if dc.DispatcherImpl == "mock" {
		expectedResponse := []byte(`[
			{"code": 200, "id": "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7"},
			{"code": 200, "id": "74368f32-4e6d-4ea2-9b8f-22dac89f9ae4"}
		]`)
		var dRes []domain.DispatcherResponse
		err := json.Unmarshal(expectedResponse, &dRes)
		return dRes, err
	}

	reqBody, err := json.Marshal(inputs)
	if err != nil {
		fmt.Println("Error marshalling inputs for request body: ", err)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", dc.DispatcherHost+"/internal/dispatch", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error constructing request to playbook-dispatcher: ", err)
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("PSK %s", dc.DispatcherPSK))
	req.Header.Set("Content-Type", "application/json")

	res, err := dc.Client.Do(req)
	if err != nil {
		fmt.Println("Error during request to playbook-dispatcher: ", err)
		return nil, err
	}
	defer res.Body.Close()

	var dRes []domain.DispatcherResponse
	err = json.NewDecoder(res.Body).Decode(&dRes)
	return dRes, err
}
