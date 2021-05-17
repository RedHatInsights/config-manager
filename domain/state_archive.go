package domain

import (
	"time"

	"github.com/google/uuid"
)

type StateArchive struct {
	AccountID string    `db:"account_id" json:"account"`
	StateID   uuid.UUID `db:"state_id" json:"id"`
	Label     string    `db:"label" json:"label"`
	Initiator string    `db:"initiator" json:"initiator"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	State     StateMap  `db:"state" json:"state"`
}

type StateArchives struct {
	Count  int            `json:"count"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
	Total  int            `json:"total"`
	States []StateArchive `json:"results"`
}

type StateArchiveRepository interface {
	GetStateArchive(s *StateArchive) (*StateArchive, error)
	GetAllStateArchives(accountID, sortBy string, limit, offset int) (*StateArchives, error)
	DeleteStateArchive(s *StateArchive) error
	CreateStateArchive(s *StateArchive) error
}
