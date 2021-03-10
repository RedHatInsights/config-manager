package persistence

import (
	"bytes"
	"config-manager/domain"
	"encoding/json"
	"fmt"
	"net/http"
)

type DispatcherRepository struct {
	DispatcherURL string
	DispatcherPSK string
	PlaybookURL   string
}

// Placeholder
func (r *DispatcherRepository) Dispatch(clientID string, acc *domain.AccountState) (*domain.DispatcherResponse, error) {
	fmt.Println("Sending request to playbook dispatcher for client: ", clientID)

	input := &domain.DispatcherInput{
		Recipient: clientID,
		Account:   acc.AccountID,
		URL:       fmt.Sprintf(r.PlaybookURL, acc.StateID),
		Labels: map[string]string{
			"playbook-cm": acc.StateID.String(),
		},
	}

	reqBody, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}

	client := http.DefaultClient

	req, err := http.NewRequest("POST", r.DispatcherURL, bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", fmt.Sprintf("PSK %s", r.DispatcherPSK))

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	var dr *domain.DispatcherResponse

	err = json.NewDecoder(res.Body).Decode(&dr)

	return dr, nil
}

// Placeholder
func (r *DispatcherRepository) GetStatus(label string) ([]domain.DispatcherRun, error) {
	fmt.Println("Getting status for hosts using label: ", label)
	return nil, nil
}
