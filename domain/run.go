package domain

type Run struct {
	RunID     string `db:"run_id"`
	AccountID string `db:"account_id"`
	Initiator string `db:"initiator"`
	Status    string `db:"status"`
	Playbook  string `db:"playbook"`
	Timestamp string `db:"timestamp"`
}

type RunRepository interface {
	GetRun(id string) (*Run, error)
	UpdateRun(r *Run) error
	DeleteRun(r *Run) error
	CreateRun(r *Run) error
}
