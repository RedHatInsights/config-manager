package application

import (
	"config-manager/domain"
	"config-manager/infrastructure/persistence/dispatcher"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// ConfigManagerInterface is an abstraction around a subset of the
// ConfigManagerService methods.
type ConfigManagerInterface interface {
	GetAccountState(id string) (*domain.AccountState, error)
	ApplyState(ctx context.Context, acc *domain.AccountState, clients []domain.Host) ([]dispatcher.RunCreated, error)
	GetSingleStateChange(stateID string) (*domain.StateArchive, error)
}

// ConfigManagerService provides an API for interacting with backend services
// such as the local storage database, inventory, cloud-connector, and
// playbook-dispatcher.
type ConfigManagerService struct {
	Cfg                *viper.Viper
	AccountStateRepo   domain.AccountStateRepository
	StateArchiveRepo   domain.StateArchiveRepository
	CloudConnectorRepo domain.CloudConnectorClient
	InventoryRepo      domain.InventoryClient
	DispatcherRepo     dispatcher.DispatcherClient
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
	log.Println("Creating new account entry with default values")
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

// GetConnectedClients Retrieve clients from cloud-connector
func (s *ConfigManagerService) GetConnectedClients(ctx context.Context, id string) (map[string]bool, error) {
	connected := make(map[string]bool)

	clients, err := s.CloudConnectorRepo.GetConnections(ctx, id)
	if err != nil {
		return nil, err
	}

	for _, client := range clients {
		connected[client] = true
	}
	return connected, nil
}

// GetInventoryClients Retrieve clients from inventory
func (s *ConfigManagerService) GetInventoryClients(ctx context.Context, page int) (domain.InventoryResponse, error) {
	res, err := s.InventoryRepo.GetInventoryClients(ctx, page)
	if err != nil {
		return res, err
	}
	return res, nil
}

// ApplyState applies the current state to selected clients
func (s *ConfigManagerService) ApplyState(
	ctx context.Context,
	acc *domain.AccountState,
	clients []domain.Host,
) ([]dispatcher.RunCreated, error) {
	var err error
	var results []dispatcher.RunCreated
	var inputs []dispatcher.RunInput

	for i, client := range clients {
		log.Println(fmt.Sprintf("Dispatching work for client %s", client.SystemProfile.RHCID))
		input := dispatcher.RunInput{
			Recipient: client.SystemProfile.RHCID,
			Account:   acc.AccountID,
			Url:       s.Cfg.GetString("Playbook_Host") + fmt.Sprintf(s.Cfg.GetString("Playbook_Path"), acc.StateID),
			Labels: &dispatcher.RunInput_Labels{
				AdditionalProperties: map[string]string{
					"state_id": acc.StateID.String(),
					"id":       client.ID,
				},
			},
		}

		inputs = append(inputs, input)

		if len(inputs) == s.Cfg.GetInt("Dispatcher_Batch_Size") || i == len(clients)-1 {
			if inputs != nil {
				res, err := s.DispatcherRepo.Dispatch(ctx, inputs)
				if err != nil {
					log.Println(err) // TODO what happens if a message can't be dispatched? Retry?
				}

				results = append(results, res...)
				inputs = nil
			} else {
				log.Println("Nothing sent to playbook dispatcher - no systems currently connected")
			}
		}
	}

	return results, err
}

// GetStateChanges gets list of state archives/changes
// TODO: Add sorting and filtering
// Sorting: currently only ascending
// Filtering idea: may need to filter on user/initiator
func (s *ConfigManagerService) GetStateChanges(accountID, sortBy string, limit, offset int) (*domain.StateArchives, error) {
	states, err := s.StateArchiveRepo.GetAllStateArchives(accountID, sortBy, limit, offset)
	if err != nil {
		return nil, err
	}

	return states, err
}

// GetSingleStateChange gets a single state archive by state_id
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

// GetPlaybook gets a playbook by state_id
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
