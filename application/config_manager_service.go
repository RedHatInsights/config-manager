package application

import (
	"config-manager/domain"
	"config-manager/infrastructure/persistence/cloudconnector"
	"config-manager/infrastructure/persistence/dispatcher"
	"config-manager/internal/config"
	"config-manager/internal/db"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
)

// ConfigManagerInterface is an abstraction around a subset of the
// ConfigManagerService methods.
type ConfigManagerInterface interface {
	GetAccountState(id string) (*domain.AccountState, error)
	ApplyState(ctx context.Context, acc *domain.AccountState, clients []domain.Host) ([]dispatcher.RunCreated, error)
	GetSingleStateChange(stateID string) (*domain.StateArchive, error)
	SetupHost(ctx context.Context, host domain.Host) (string, error)
}

// ConfigManagerService provides an API for interacting with backend services
// such as the local storage database, inventory, cloud-connector, and
// playbook-dispatcher.
type ConfigManagerService struct {
	AccountStateRepo   domain.AccountStateRepository
	StateArchiveRepo   domain.StateArchiveRepository
	CloudConnectorRepo cloudconnector.CloudConnectorClient
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
			if err != nil {
				return nil, err
			}
		default:
			return nil, err
		}
	}

	return acc, err
}

func (s *ConfigManagerService) setupDefaultState(acc *domain.AccountState) (*domain.AccountState, error) {
	log.Info().Msgf("Creating new account entry with default values")
	err := s.AccountStateRepo.CreateAccountState(acc)
	if err != nil {
		return nil, err
	}

	defaultState := config.DefaultConfig.ServiceConfig
	state := domain.StateMap{}
	if err := json.Unmarshal([]byte(defaultState), &state); err != nil {
		return nil, err
	}
	acc, err = s.UpdateAccountState(acc.AccountID, "redhat", state, db.JSONNullBool{NullBool: sql.NullBool{Valid: true, Bool: true}})

	return acc, err
}

// UpdateAccountState updates the current state for the account and creates a new state archive
func (s *ConfigManagerService) UpdateAccountState(id, user string, payload domain.StateMap, applyState db.JSONNullBool) (*domain.AccountState, error) {
	newStateID := uuid.New()
	newLabel := id + "-" + uuid.New().String()
	acc := &domain.AccountState{
		AccountID:  id,
		State:      payload,
		StateID:    newStateID,
		Label:      newLabel,
		ApplyState: applyState,
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

	if acc.ApplyState.Valid && !acc.ApplyState.Bool {
		log.Info().Msgf("account_state.apply_state is false; skipping configuration")
		return []dispatcher.RunCreated{}, nil
	}

	log.Info().Msgf("applying state for %v clients: %v", len(clients), acc.State)
	for i, client := range clients {
		if client.Reporter == "cloud-connector" {
			log.Debug().Msgf("setting up host %v for playbook execution", client.DisplayName)
			if _, err := s.SetupHost(context.Background(), client); err != nil {
				log.Info().Msgf("error setting up host '%v': %v", client, err)
				continue
			}
		}

		log.Info().Msgf("Dispatching work for client %s", client.SystemProfile.RHCID)
		input := dispatcher.RunInput{
			Recipient: client.SystemProfile.RHCID,
			Account:   acc.AccountID,
			Url:       config.DefaultConfig.PlaybookHost.String() + fmt.Sprintf(config.DefaultConfig.PlaybookPath, acc.StateID),
			Labels: &dispatcher.RunInput_Labels{
				AdditionalProperties: map[string]string{
					"state_id": acc.StateID.String(),
					"id":       client.ID,
				},
			},
		}

		inputs = append(inputs, input)

		if len(inputs) == config.DefaultConfig.DispatcherBatchSize || i == len(clients)-1 {
			if inputs != nil {
				res, err := s.DispatcherRepo.Dispatch(ctx, inputs)
				if err != nil {
					log.Error().Err(err) // TODO what happens if a message can't be dispatched? Retry?
				}

				results = append(results, res...)
				inputs = nil
			} else {
				log.Info().Msg("Nothing sent to playbook dispatcher - no systems currently connected")
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

// SetApplyState sets the apply_state field to skipApplyState
func (s *ConfigManagerService) SetApplyState(accountID string, applyState bool) error {
	return s.AccountStateRepo.UpdateAccountStateApplyState(accountID, applyState)
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

// SetupHost messages a host to install the rhc-worker-playbook RPM to enable it
// to receive and execute playbooks from the playbook-dispatcher service.
func (s *ConfigManagerService) SetupHost(ctx context.Context, host domain.Host) (string, error) {
	status, dispatchers, err := s.CloudConnectorRepo.GetConnectionStatus(ctx, host.Account, host.SystemProfile.RHCID)
	if err != nil {
		return "", err
	}

	if status != "connected" {
		return "", fmt.Errorf("cannot set up host: host connection status = %v", status)
	}

	if _, has := dispatchers["package-manager"]; !has {
		return "", fmt.Errorf("host %v missing required directive 'package-manager'", host.SystemProfile.RHCID)
	}

	if _, has := dispatchers["rhc-worker-playbook"]; has {
		return "", nil
	}

	payload := struct {
		Command string `json:"command"`
		Name    string `json:"name"`
	}{
		Command: "install",
		Name:    "rhc-worker-playbook",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("cannot marshal payload: %v", err)
	}

	messageID, err := s.CloudConnectorRepo.SendMessage(ctx, host.Account, "package-manager", data, nil, host.SystemProfile.RHCID)
	if err != nil {
		return "", err
	}

	for {
		status, dispatchers, err := s.CloudConnectorRepo.GetConnectionStatus(ctx, host.Account, host.SystemProfile.RHCID)
		if err != nil {
			return "", err
		}
		if status == "disconnected" {
			return messageID, fmt.Errorf("host disconnected while waiting for connection status")
		}
		if _, has := dispatchers["rhc-worker-playbook"]; has {
			break
		}
		time.Sleep(30 * time.Second)
	}

	return messageID, nil

}
