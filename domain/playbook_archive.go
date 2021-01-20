package domain

import (
	"time"

	"github.com/google/uuid"
)

type PlaybookArchive struct {
	PlaybookID uuid.UUID `db:"playbook_id"`
	AccountID  string    `db:"account_id"`
	RunID      uuid.UUID `db:"run_id"`
	Filename   string    `db:"filename"`
	CreatedAt  time.Time `db:"created_at"`
	State      StateMap  `db:"state"`
}

type PlaybookArchiveRepository interface {
	GetPlaybookArchiveByRunID(runID uuid.UUID) (*PlaybookArchive, error)
	DeletePlaybookArchive(p *PlaybookArchive) error
	CreatePlaybookArchive(p *PlaybookArchive) error
}
