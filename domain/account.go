package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Account struct {
	AccountID string   `db:"account_id"`
	State     StateMap `db:"state"`
}

type AccountRepository interface {
	GetAccount(acc *Account) (*Account, error)
	UpdateAccount(acc *Account) error
	DeleteAccount(acc *Account) error
	CreateAccount(acc *Account) error
}

type StateMap map[string]interface{}

// Value interface for StateMap
func (s StateMap) Value() (driver.Value, error) {
	j, err := json.Marshal(s)
	return j, err
}

// Scan interface for StateMap
func (s *StateMap) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &s)
}
