package application

import (
	"config-manager/domain"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ConfigManagerService enables communication between the api and other resources (db + other apis)
type ConfigManagerService struct {
	AccountStateRepo domain.AccountStateRepository
	RunRepo          domain.RunRepository
	StateArchiveRepo domain.StateArchiveRepository
	ClientListRepo   domain.ClientListRepository
	DispatcherRepo   domain.DispatcherRepository
}

// GetAccountState retrieves the current state for the account
func (s *ConfigManagerService) GetAccountState(id string) (*domain.AccountState, error) {
	acc := &domain.AccountState{AccountID: id}
	acc, err := s.AccountStateRepo.GetAccountState(acc)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			fmt.Println("Creating new account entry")
			acc, err = s.createAccountState(id)
		default:
			return nil, err
		}
	}

	return acc, err
}

// UpdateAccountState updates the current state for the account and creates a new state archive
func (s *ConfigManagerService) UpdateAccountState(id, user string, payload map[string]string) (*domain.AccountState, error) {
	newStateID := uuid.New()
	newLabel := id + "-" + uuid.New().String()
	acc := &domain.AccountState{
		AccountID: id,
		State:     payload,
		StateID:   newStateID,
		Label:     newLabel,
	}

	err := s.AccountStateRepo.UpdateAccountState(acc)
	if err != nil {
		return nil, err
	}

	archive := &domain.StateArchive{
		AccountID: acc.AccountID,
		StateID:   acc.StateID,
		Label:     acc.Label,
		Initiator: user,
		CreatedAt: time.Now(),
		State:     acc.State,
	}

	err = s.StateArchiveRepo.CreateStateArchive(archive)
	if err != nil {
		return nil, err
	}

	return acc, err
}

// DeleteAccount TODO
func (s *ConfigManagerService) DeleteAccount(id string) error {
	return nil
}

func (s *ConfigManagerService) createAccountState(id string) (*domain.AccountState, error) {
	stateID := uuid.New()
	label := id + "-default"
	acc := &domain.AccountState{
		AccountID: id,
		State: domain.StateMap{
			"insights":   "enabled",
			"advisor":    "enabled",
			"compliance": "enabled",
		},
		StateID: stateID,
		Label:   label,
	}

	err := s.AccountStateRepo.CreateAccountState(acc)
	if err != nil {
		return nil, err
	}

	archive := &domain.StateArchive{
		AccountID: acc.AccountID,
		StateID:   acc.StateID,
		Label:     acc.Label,
		Initiator: "redhat",
		CreatedAt: time.Now(),
		State:     acc.State,
	}

	err = s.StateArchiveRepo.CreateStateArchive(archive)
	if err != nil {
		return nil, err
	}

	return acc, err
}

// GetClients TODO: Retrieve clients from inventory
func (s *ConfigManagerService) GetClients(id string) (*domain.ClientList, error) {
	clients, err := s.ClientListRepo.GetConnectedClients(id)
	if err != nil {
		return nil, err
	}
	return clients, nil
}

// ApplyState applies the current state to selected clients
// TODO: Change return type to satisfy openapi response
// TODO: Separate application function for automatic applications via kafka?
func (s *ConfigManagerService) ApplyState(acc *domain.AccountState, user string, clients []domain.Client) ([]*domain.DispatcherResponse, error) {
	// construct and send work request to playbook dispatcher
	// includes url to retrieve the playbook, url to upload results, and which client to send work to
	var err error
	var results []*domain.DispatcherResponse
	for _, client := range clients {
		res, err := s.DispatcherRepo.Dispatch(client.ClientID)
		if err != nil {
			fmt.Println(err) // TODO what happens if a message can't be dispatched? Retry?
		}

		runID := uuid.New()
		initialTime := time.Now()

		newRun := &domain.Run{
			RunID:     runID,
			AccountID: acc.AccountID, // Could runID come from dispatcher response?
			Hostname:  client.Hostname,
			Initiator: user,
			Label:     acc.Label,
			Status:    "in progress",
			CreatedAt: initialTime,
			UpdatedAt: initialTime,
		}

		err = s.RunRepo.CreateRun(newRun)
		if err != nil {
			return nil, err
		}

		results = append(results, res)
	}

	return results, err
}

// GetStateChanges gets list of state archives/changes
// TODO: Add sorting and filtering
// Sorting: currently only ascending
// Filtering idea: may need to filter on user/initiator
func (s *ConfigManagerService) GetStateChanges(accountID string, limit, offset int) ([]domain.StateArchive, error) {
	states, err := s.StateArchiveRepo.GetAllStateArchives(accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	return states, err
}

// GetSingleStateChange gets a single state archive by state_id
// TODO: Function to get current state?
// State archives contain additional information over the AccountState so this could be useful
func (s *ConfigManagerService) GetSingleStateChange(stateID string) (*domain.StateArchive, error) {
	id, err := uuid.Parse(stateID)
	if err != nil {
		return nil, err
	}

	archive := &domain.StateArchive{StateID: id}
	state, err := s.StateArchiveRepo.GetStateArchive(archive)
	if err != nil {
		return nil, err
	}

	return state, err
}

// GetSingleRun gets a single run entry by run_id
func (s *ConfigManagerService) GetSingleRun(runID string) (*domain.Run, error) {
	id, err := uuid.Parse(runID)
	if err != nil {
		return nil, err
	}
	run, err := s.RunRepo.GetRun(id)
	if err != nil {
		return nil, err
	}

	return run, err
}

// GetRuns gets runs for the account
// TODO: Expand on filter - allow filtering by hostname/status/label/user
func (s *ConfigManagerService) GetRuns(accountID, filter, sortBy string, limit, offset int) ([]domain.Run, error) {
	runs, err := s.RunRepo.GetRuns(accountID, filter, sortBy, limit, offset)
	if err != nil {
		return nil, err
	}

	return runs, err
}

// GetRunStatus TODO: Get status updates from playbook dispatcher
// PLACEHOLDER
func (s *ConfigManagerService) GetRunStatus(label string) ([]domain.DispatcherRun, error) {
	statusList, err := s.DispatcherRepo.GetStatus(label)
	if err != nil {
		return nil, err
	}

	return statusList, err
}
