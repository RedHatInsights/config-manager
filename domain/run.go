package domain

import (
	"time"

	"github.com/google/uuid"
)

type Run struct {
	RunID     uuid.UUID `db:"run_id"`
	AccountID string    `db:"account_id"`
	Hostname  string    `db:"hostname"`
	Initiator string    `db:"initiator"`
	Label     string    `db:"label"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type RunRepository interface {
	GetRun(id uuid.UUID) (*Run, error)
	GetRunsByLabel(label string, limit, offset int) ([]Run, error)
	UpdateRunStatus(r *Run) error
	CreateRun(r *Run) error
}
