package persistence

import (
	"bytes"
	"config-manager/domain"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type DispatcherClient struct {
	DispatcherHost string
	DispatcherPSK  string
	Client         http.Client
}

func (r *DispatcherClient) Dispatch(
	ctx context.Context,
	input domain.DispatcherInput,
) (*domain.DispatcherResponse, error) {
	fmt.Println("Sending request to playbook dispatcher for client: ", input.Recipient)

	reqBody, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.DispatcherHost+"/internal/dispatch", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", fmt.Sprintf("PSK %s", r.DispatcherPSK))

	res, err := r.Client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	var dRes *domain.DispatcherResponse
	err = json.NewDecoder(res.Body).Decode(&dRes)
	return dRes, err
}

// Placeholder
// func (r *DispatcherRepository) GetStatus(label string) ([]domain.DispatcherRun, error) {
// 	fmt.Println("Getting status for hosts using label: ", label)
// 	return nil, nil
// }
