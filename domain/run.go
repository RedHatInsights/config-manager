package domain

type Run struct {
	RunID     string `db:"run_id"`
	AccountID string `db:"account_id"`
	Initiator string `db:"initiator"`
	Label     string `db:"label"`
	Status    string `db:"status"`
	Timestamp string `db:"timestamp"`
}

type RunRepository interface {
	GetRun(id string) (*Run, error)
	GetRuns(id string, limit, offset int) ([]Run, error)
	UpdateRunStatus(r *Run) error
	CreateRun(r *Run) error
}
