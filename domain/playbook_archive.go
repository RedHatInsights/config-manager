package domain

import (
	"github.com/google/uuid"
)

type PlaybookArchive struct {
	AccountID string    `db:"account_id" json:"account"`
	StateID   uuid.UUID `db:"state_id" json:"id"`
	Playbook  string    `db:"playbook" json:"playbook"`
}

type PlaybookArchiveRepository interface {
	GetPlaybookArchive(pb *PlaybookArchive) (*PlaybookArchive, error)
	DeletePlaybookArchive(pb *PlaybookArchive) error
	CreatePlaybookArchive(pb *PlaybookArchive) error
}
