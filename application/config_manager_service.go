package application

import (
	"config-manager/domain"
	"database/sql"
	"fmt"
)

type ConfigManagerService struct {
	AccountRepo domain.AccountRepository
	RunRepo     domain.RunRepository
}

func (s *ConfigManagerService) GetAccount(id string) (*domain.Account, error) {
	acc, err := s.AccountRepo.GetAccount(id)

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
