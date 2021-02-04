package domain

import (
	"time"

	"github.com/google/uuid"
)

type Run struct {
	RunID     uuid.UUID `db:"run_id" json:"id"`
	AccountID string    `db:"account_id" json:"account"`
	Hostname  string    `db:"hostname" json:"hostname"`
	Initiator string    `db:"initiator" json:"initiator"`
	Label     string    `db:"label" json:"label"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type RunRepository interface {
	GetRun(id uuid.UUID) (*Run, error)
	GetRuns(accountID, filter, sortBy string, limit, offset int) ([]Run, error)
	// GetRunsByLabel(label string, limit, offset int) ([]Run, error)
	UpdateRunStatus(r *Run) error
	CreateRun(r *Run) error
}
