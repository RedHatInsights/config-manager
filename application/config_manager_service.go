package application

import (
	"config-manager/domain"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/viper"

	"github.com/google/uuid"
)

// ConfigManagerService enables communication between the api and other resources (db + other apis)
type ConfigManagerService struct {
	Cfg              *viper.Viper
	AccountStateRepo domain.AccountStateRepository
	StateArchiveRepo domain.StateArchiveRepository
	ClientListRepo   domain.ClientListRepository
	DispatcherRepo   domain.DispatcherRepository
	PlaybookRepo     domain.PlaybookArchiveRepository
	PBGenerator      Generator
}

// GetAccountState retrieves the current state for the account
func (s *ConfigManagerService) GetAccountState(id string) (*domain.AccountState, error) {
	acc := &domain.AccountState{AccountID: id}
	acc, err := s.AccountStateRepo.GetAccountState(acc)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			fmt.Println("Creating new account entry with default values")
			defaultState := s.Cfg.GetString("DefaultServiceEnablement")
			state := domain.StateMap{}
			json.Unmarshal([]byte(defaultState), &state)
			acc, err = s.UpdateAccountState(id, "redhat", state)
		default:
			return nil, err
		}
	}

	return acc, err
}

// UpdateAccountState updates the current state for the account and creates a new state archive
// TODO refactor
func (s *ConfigManagerService) UpdateAccountState(id, user string, payload map[string]string) (*domain.AccountState, error) {
	newStateID := uuid.New()
	newLabel := id + "-" + uuid.New().String()
	acc := &domain.AccountState{
		AccountID: id,
		State:     payload,
		StateID:   newStateID,
		Label:     newLabel,
	}

	pb, err := s.PBGenerator.GeneratePlaybook(acc.State)
	if err != nil {
		return nil, err // TODO improve errors
	}

	err = s.AccountStateRepo.UpdateAccountState(acc)
	if err != nil {
		return nil, err
	}

	stateArchive := &domain.StateArchive{
		AccountID: acc.AccountID,
		StateID:   acc.StateID,
		Label:     acc.Label,
		Initiator: user,
		CreatedAt: time.Now(),
		State:     acc.State,
	}

	err = s.StateArchiveRepo.CreateStateArchive(stateArchive)
	if err != nil {
		return nil, err
	}

	playbookArchive := &domain.PlaybookArchive{
		AccountID: acc.AccountID,
		StateID:   acc.StateID,
		Playbook:  pb,
	}

	err = s.PlaybookRepo.CreatePlaybookArchive(playbookArchive)
	if err != nil {
		return nil, err
	}

	return acc, err
}

// DeleteAccount TODO
func (s *ConfigManagerService) DeleteAccount(id string) error {
	return nil
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

func (s *ConfigManagerService) GetPlaybook(stateID string) (string, error) {
	id, err := uuid.Parse(stateID)
	if err != nil {
		return "", err
	}

	archive := &domain.PlaybookArchive{StateID: id}
	playbook, err := s.PlaybookRepo.GetPlaybookArchive(archive)
	if err != nil {
		fmt.Println("Could not retrieve playbook")
		return "", err
	}

	return playbook.Playbook, err
}
