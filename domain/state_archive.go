package domain

import (
	"time"

	"github.com/google/uuid"
)

type StateArchive struct {
	AccountID string    `db:"account_id"`
	StateID   uuid.UUID `db:"state_id"`
	Label     string    `db:"label"`
	Initiator string    `db:"initiator"`
	CreatedAt time.Time `db:"created_at"`
	State     StateMap  `db:"state"`
}

type StateArchiveRepository interface {
	GetStateArchive(s *StateArchive) (*StateArchive, error)
	GetAllStateArchives(accountID string, limit, offset int) ([]StateArchive, error)
	DeleteStateArchive(s *StateArchive) error
	CreateStateArchive(s *StateArchive) error
}
