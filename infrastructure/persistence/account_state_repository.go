package persistence

import (
	"config-manager/domain"
	"database/sql"
)

type AccountStateRepository struct {
	DB *sql.DB
}

func (r *AccountStateRepository) GetAccountState(acc *domain.AccountState) (*domain.AccountState, error) {
	err := r.DB.QueryRow("SELECT state, state_id, label FROM account_states WHERE account_id=$1",
		acc.AccountID).Scan(&acc.State, &acc.StateID, &acc.Label)

	return acc, err
}

func (r *AccountStateRepository) UpdateAccountState(acc *domain.AccountState) error {
	_, err := r.DB.Exec("UPDATE account_states SET state=$1, state_id=$2, label=$3 WHERE account_id=$4",
		acc.State, acc.StateID, acc.Label, acc.AccountID)

	return err
}

func (r *AccountStateRepository) DeleteAccountState(acc *domain.AccountState) error {
	_, err := r.DB.Exec("DELETE FROM account_states WHERE account_id=$1", acc.AccountID)

	return err
}

func (r *AccountStateRepository) CreateAccountState(acc *domain.AccountState) error {
	_, err := r.DB.Exec("INSERT INTO account_states(account_id, state, state_id, label) VALUES($1, $2, $3, $4)",
		acc.AccountID, acc.State, acc.StateID, acc.Label)

	return err
}
