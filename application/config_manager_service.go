package application

import (
	"config-manager/domain"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ConfigManagerService struct {
	AccountStateRepo domain.AccountStateRepository
	RunRepo          domain.RunRepository
	StateArchiveRepo domain.StateArchiveRepository
	ClientListRepo   domain.ClientListRepository
	DispatcherRepo   domain.DispatcherRepository
}

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

func (s *ConfigManagerService) UpdateAccountState(id, user string, payload map[string]interface{}) (*domain.AccountState, error) {
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

func (s *ConfigManagerService) GetClients(id string) (*domain.ClientList, error) {
	clients, err := s.ClientListRepo.GetConnectedClients(id)
	if err != nil {
		return nil, err
	}
	return clients, nil
}

func (s *ConfigManagerService) ApplyState(id, user string, clients []domain.Client) (*domain.AccountState, error) {

	acc, _ := s.GetAccountState(id) // GetAccount or just have state passed in via the api call?

	// construct and send work request to playbook dispatcher
	// includes url to retrieve the playbook, url to upload results, and which client to send work to
	var err error
	for _, client := range clients {
		res, err := s.DispatcherRepo.Dispatch(client.ClientID)
		if err != nil {
			fmt.Println(err) // TODO what happens if a message can't be dispatched? Retry?
		}
		fmt.Println(res.Code)

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
	}

	return acc, err
}

func (s *ConfigManagerService) GetStateChanges(accountID string, limit, offset int) ([]domain.StateArchive, error) {
	states, err := s.StateArchiveRepo.GetAllStateArchives(accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	return states, err
}

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

func (s *ConfigManagerService) GetRunsByLabel(label string, limit, offset int) ([]domain.Run, error) {
	runs, err := s.RunRepo.GetRunsByLabel(label, limit, offset)
	if err != nil {
		return nil, err
	}

	return runs, err
}

func (s *ConfigManagerService) GetRunStatus(label string) ([]domain.DispatcherRun, error) {
	statusList, err := s.DispatcherRepo.GetStatus(label)
	if err != nil {
		return nil, err
	}

	return statusList, err
}
