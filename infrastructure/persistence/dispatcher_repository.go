package persistence

import (
	"config-manager/domain"
	"fmt"
)

type DispatcherRepository struct {
	DispatcherURL string
	RunStatusURL  string
	PlaybookURL   string
}

// Placeholder
func (r *DispatcherRepository) Dispatch(clientID string) (*domain.DispatcherResponse, error) {
	fmt.Println("Sending request to playbook dispatcher for client: ", clientID)
	res := &domain.DispatcherResponse{
		Code:  200,
		RunID: "id-from-playbook-dispatcher",
	}
	return res, nil
}

// Placeholder
func (r *DispatcherRepository) GetStatus(label string) ([]domain.DispatcherRun, error) {
	fmt.Println("Getting status for hosts using label: ", label)
	return nil, nil
}
