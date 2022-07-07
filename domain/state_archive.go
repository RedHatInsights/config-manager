package domain

import (
	"time"

	"github.com/google/uuid"
)

// StateArchive represents both a state_archive database record and JSON API
// object.
type StateArchive struct {
	AccountID string    `db:"account_id" json:"account"`
	OrgID     string    `db:"org_id" json:"org_id"`
	StateID   uuid.UUID `db:"state_id" json:"id"`
	Label     string    `db:"label" json:"label"`
	Initiator string    `db:"initiator" json:"initiator"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	State     StateMap  `db:"state" json:"state"`
}

// StateArchives represents the collection of archives retrieved from the
// database.
type StateArchives struct {
	Count  int            `json:"count"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
	Total  int            `json:"total"`
	States []StateArchive `json:"results"`
}

// StateArchiveRepository is an abstraction of the CRUD API methods for
// accessing state archive information.
type StateArchiveRepository interface {
	GetStateArchive(s *StateArchive) (*StateArchive, error)
	GetAllStateArchives(accountID, sortBy string, limit, offset int) (*StateArchives, error)
	DeleteStateArchive(s *StateArchive) error
	CreateStateArchive(s *StateArchive) error
}
