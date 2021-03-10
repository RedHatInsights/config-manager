package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

type AccountState struct {
	AccountID string    `db:"account_id" json:"account"`
	State     StateMap  `db:"state" json:"state"`
	StateID   uuid.UUID `db:"state_id" json:"id"`
	Label     string    `db:"label" json:"label"`
}

type AccountStateRepository interface {
	GetAccountState(acc *AccountState) (*AccountState, error)
	UpdateAccountState(acc *AccountState) error
	DeleteAccountState(acc *AccountState) error
	CreateAccountState(acc *AccountState) error
}

type StateMap map[string]string

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

func (s StateMap) GetKeys() []string {
	keys := make([]string, 0, len(s))
	for key := range s {
		keys = append(keys, key)
	}

	return keys
}
