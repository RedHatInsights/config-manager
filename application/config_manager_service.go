package application

import (
	"config-manager/domain"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ConfigManagerService struct {
	AccountRepo  domain.AccountRepository
	RunRepo      domain.RunRepository
	PlaybookRepo domain.PlaybookArchiveRepository
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

func (s *ConfigManagerService) GetClients(id string) []string {
	// placeholder - request clients from external service (inventory)
	var clients []string
	clients = append(clients, "1234")
	return clients
}

func (s *ConfigManagerService) ApplyState(id, user string, clients []string) (*domain.Run, error) {
	// generate run ID and label
	// create entry in run table
	// create entry in playbook archive
	// for each client: send work request to dispatcher w/ label

	acc, _ := s.GetAccount(id)

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
		PlaybookID: playbookID.String(),
		RunID:      runID.String(),
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

	return newRun, err
}
