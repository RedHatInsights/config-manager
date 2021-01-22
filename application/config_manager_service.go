package application

import (
	"config-manager/domain"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ConfigManagerService struct {
	AccountRepo    domain.AccountRepository
	RunRepo        domain.RunRepository
	PlaybookRepo   domain.PlaybookArchiveRepository
	ClientListRepo domain.ClientListRepository
	DispatcherRepo domain.DispatcherRepository
}

func (s *ConfigManagerService) GetAccount(id string) (*domain.Account, error) {
	acc := &domain.Account{AccountID: id}
	acc, err := s.AccountRepo.GetAccount(acc)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			fmt.Println("Creating new account entry")
			acc, err = s.createAccount(id)
		default:
			return nil, err
		}
	}

	return acc, err
}

func (s *ConfigManagerService) UpdateAccount(id string, payload map[string]interface{}) (*domain.Account, error) {
	acc := &domain.Account{
		AccountID: id,
		State:     payload,
	}

	err := s.AccountRepo.UpdateAccount(acc)
	if err != nil {
		return nil, err
	}

	return acc, err
}

func (s *ConfigManagerService) DeleteAccount(id string) error {
	return nil

}

func (s *ConfigManagerService) createAccount(id string) (*domain.Account, error) {
	acc := &domain.Account{
		AccountID: id,
		State: domain.StateMap{
			"insights":   "enabled",
			"advisor":    "enabled",
			"compliance": "enabled",
		},
	}

	err := s.AccountRepo.CreateAccount(acc)
	if err != nil {
		return nil, err
	}

	return acc, err
}

func (s *ConfigManagerService) GetClients(id string) ([]string, error) {
	clients, err := s.ClientListRepo.GetConnectedClients(id)
	if err != nil {
		return nil, err
	}
	return clients.Clients, nil
}

func (s *ConfigManagerService) ApplyState(id, user string, clients []string) (*domain.Run, error) {
	// generate run ID and label
	// create entry in run table
	// create entry in playbook archive
	// for each client: send work request to dispatcher w/ label

	acc, _ := s.GetAccount(id) // GetAccount or just have state passed in via the api call?

	runID := uuid.New()
	label := runID.String() + "-demo-label"

	newRun := &domain.Run{
		AccountID: id,
		RunID:     runID,
		Initiator: user,
		Label:     label,
		Status:    "in progress",
		CreatedAt: time.Now(),
	}

	err := s.RunRepo.CreateRun(newRun)
	if err != nil {
		return nil, err
	}

	playbookID := uuid.New()

	playbookArchive := &domain.PlaybookArchive{
		PlaybookID: playbookID,
		RunID:      runID,
		AccountID:  id,
		Filename:   "test",
		CreatedAt:  time.Now(),
		State:      acc.State,
	}

	err = s.PlaybookRepo.CreatePlaybookArchive(playbookArchive)
	if err != nil {
		return nil, err
	}

	// construct and send work request to playbook dispatcher
	// includes url to retrieve the playbook, url to upload results, and which client to send work to
	for _, client := range clients {
		res, err := s.DispatcherRepo.Dispatch(client)
		if err != nil {
			fmt.Println(err) // TODO what happens if a message can't be dispatched? Retry?
		}
		fmt.Println(res.Code)
	}

	return newRun, err
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

func (s *ConfigManagerService) GetRuns(accountID string, limit, offset int) ([]domain.Run, error) {
	runs, err := s.RunRepo.GetRuns(accountID, limit, offset)
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
