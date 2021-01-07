package main

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type account struct {
	AccountID string
	State     StateMap
}

type host struct {
	ClientID    string            `json:"client-id"`
	InventoryID string            `json:"inventory-id"`
	Account     string            `json:"account"`
	RunID       string            `json:"run-id"`
	State       map[string]string `json:"state"`
}

type run struct {
	RunID     string `json:"run-id"`
	Account   string `json:"account"`
	Initiator string `json:"initiator"`
	Status    string `json:"status"`
	Playbook  string `json:"playbook"`
	Timestamp string `json:"timestamp"`
}

// StateMap is used to assist in converting from JSONB to Map for postgres
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

func (a *account) getAccount(db *sql.DB) error {
	return db.QueryRow("SELECT state FROM accounts WHERE account_id=$1",
		a.AccountID).Scan(&a.State)
}

func (a *account) updateAccount(db *sql.DB) error {
	_, err := db.Exec("UPDATE accounts SET state=$1 WHERE account_id=$2",
		a.State, a.AccountID)

	return err
}

func (a *account) deleteAccount(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM accounts WHERE account_id=$1", a.AccountID)

	return err
}

func (a *account) createAccount(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO accounts(account_id, state) VALUES($1, $2)",
		a.AccountID, a.State)

	return err
}

func (h *host) getHost(db *sql.DB) error {
	return errors.New("Not implemented")
}

func (h *host) updateHost(db *sql.DB) error {
	return errors.New("Not implemented")
}

func (h *host) deleteHost(db *sql.DB) error {
	return errors.New("Not implemented")
}

func (h *host) createHost(db *sql.DB) error {
	return errors.New("Not implemented")
}

func (h *host) getAllHosts(db *sql.DB) error {
	return errors.New("Not implemented")
}

func (r *run) getRun(db *sql.DB) error {
	return errors.New("Not implemented")
}

func (r *run) updateRun(db *sql.DB) error {
	return errors.New("Not implemented")
}

func (r *run) createRun(db *sql.DB) error {
	return errors.New("Not implemented")
}
