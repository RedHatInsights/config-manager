package persistence

import (
	"config-manager/domain"
	"database/sql"
)

type AccountRepository struct {
	DB *sql.DB
}

func (r *AccountRepository) GetAccount(id string) (*domain.Account, error) {
	acc := &domain.Account{AccountID: id}
	err := r.DB.QueryRow("SELECT state FROM accounts WHERE account_id=$1",
		acc.AccountID).Scan(&acc.State)

	return acc, err
}

func (r *AccountRepository) UpdateAccount(acc *domain.Account) error {
	_, err := r.DB.Exec("UPDATE accounts SET state=$1 WHERE account_id=$2",
		acc.State, acc.AccountID)

	return err
}

func (r *AccountRepository) DeleteAccount(acc *domain.Account) error {
	_, err := r.DB.Exec("DELETE FROM accounts WHERE account_id=$1", acc.AccountID)

	return err
}

func (r *AccountRepository) CreateAccount(acc *domain.Account) error {
	_, err := r.DB.Exec("INSERT INTO accounts(account_id, state) VALUES($1, $2)",
		acc.AccountID, acc.State)

	return err
}
