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
	Client         utils.HTTPClient
}

func (r *DispatcherClient) Dispatch(
	ctx context.Context,
	inputs []domain.DispatcherInput,
) ([]domain.DispatcherResponse, error) {
	fmt.Println("Sending request to playbook dispatcher")

	reqBody, err := json.Marshal(inputs)
	if err != nil {
		fmt.Println("Error marshalling inputs for request body: ", err)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.DispatcherHost+"/internal/dispatch", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error constructing request to playbook-dispatcher: ", err)
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("PSK %s", r.DispatcherPSK))
	req.Header.Set("Content-Type", "application/json")

	res, err := r.Client.Do(req)
	if err != nil {
		fmt.Println("Error during request to playbook-dispatcher: ", err)
		return nil, err
	}
	defer res.Body.Close()

	var dRes []domain.DispatcherResponse
	err = json.NewDecoder(res.Body).Decode(&dRes)
	return dRes, err
}
