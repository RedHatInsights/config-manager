package application

import (
	"config-manager/domain"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// ConfigManagerService enables communication between the api and other resources (db + other apis)
type ConfigManagerService struct {
	Cfg                *viper.Viper
	AccountStateRepo   domain.AccountStateRepository
	StateArchiveRepo   domain.StateArchiveRepository
	CloudConnectorRepo domain.CloudConnectorClient
	DispatcherRepo     domain.DispatcherClient
	PlaybookGenerator  Generator
}

// GetAccountState retrieves the current state for the account
func (s *ConfigManagerService) GetAccountState(id string) (*domain.AccountState, error) {
	acc := &domain.AccountState{AccountID: id}
	acc, err := s.AccountStateRepo.GetAccountState(acc)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			acc, err = s.setupDefaultState(acc)
		default:
			return nil, err
		}
	}

	return acc, err
}

func (s *ConfigManagerService) setupDefaultState(acc *domain.AccountState) (*domain.AccountState, error) {
	fmt.Println("Creating new account entry with default values")
	err := s.AccountStateRepo.CreateAccountState(acc)
	if err != nil {
		return nil, err
	}

	defaultState := s.Cfg.GetString("Service_Config")
	state := domain.StateMap{}
	json.Unmarshal([]byte(defaultState), &state)
	acc, err = s.UpdateAccountState(acc.AccountID, "redhat", state)

	return acc, err
}

// UpdateAccountState updates the current state for the account and creates a new state archive
func (s *ConfigManagerService) UpdateAccountState(id, user string, payload domain.StateMap) (*domain.AccountState, error) {
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

// GetClients TODO: Retrieve clients from inventory
func (s *ConfigManagerService) GetConnectorClients(ctx context.Context, id string) ([]string, error) {
	clients, err := s.CloudConnectorRepo.GetConnections(ctx, id)
	if err != nil {
		return nil, err
	}
	return clients, nil
}

// ApplyState applies the current state to selected clients
// TODO: Separate application function for automatic applications via kafka?
func (s *ConfigManagerService) ApplyState(
	ctx context.Context,
	acc *domain.AccountState,
	clients []string,
) ([]domain.DispatcherResponse, error) {
	var err error
	var results []domain.DispatcherResponse
	var inputs []domain.DispatcherInput
	for i, client := range clients {
		input := domain.DispatcherInput{
			Recipient: client,
			Account:   acc.AccountID,
			URL:       fmt.Sprintf(s.Cfg.GetString("Playbook_URL"), acc.StateID),
			Labels: map[string]string{
				"cm-playbook": acc.StateID.String(),
			},
		}

		inputs = append(inputs, input)

		if len(inputs) == s.Cfg.GetInt("Dispatcher_Batch_Size") || i == len(clients)-1 {
			res, err := s.DispatcherRepo.Dispatch(ctx, inputs)
			if err != nil {
				fmt.Println(err) // TODO what happens if a message can't be dispatched? Retry?
			}

			results = append(results, res...)
			inputs = nil
		}
	}

	return results, err
}

// GetStateChanges gets list of state archives/changes
// TODO: Add sorting and filtering
// Sorting: currently only ascending
// Filtering idea: may need to filter on user/initiator
func (s *ConfigManagerService) GetStateChanges(accountID string, limit, offset int) (*domain.StateArchives, error) {
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

	archive := &domain.StateArchive{StateID: id}
	archive, err = s.StateArchiveRepo.GetStateArchive(archive)
	if err != nil {
		return "", err
	}

	playbook, err := s.PlaybookGenerator.GeneratePlaybook(archive.State)
	if err != nil {
		return "", err
	}

	return playbook, err
}
