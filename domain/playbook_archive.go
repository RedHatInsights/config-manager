package domain

type PlaybookArchive struct {
	PlaybookID string   `db:"playbook_id"`
	AccountID  string   `db:"account_id"`
	RunID      string   `db:"run_id"`
	Filename   string   `db:"filename"`
	CreatedAt  string   `db:"created_at"`
	State      StateMap `db:"state"`
}

type PlaybookArchiveRepository interface {
	GetPlaybookArchiveByRunID(runID string) (*PlaybookArchive, error)
	DeletePlaybookArchive(p *PlaybookArchive) error
	CreatePlaybookArchive(p *PlaybookArchive) error
}
