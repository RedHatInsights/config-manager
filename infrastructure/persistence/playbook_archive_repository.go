package persistence

import (
	"config-manager/domain"
	"database/sql"

	"github.com/google/uuid"
)

type PlaybookArchiveRepository struct {
	DB *sql.DB
}

func (r *PlaybookArchiveRepository) GetPlaybookArchiveByRunID(runID uuid.UUID) (*domain.PlaybookArchive, error) {
	p := &domain.PlaybookArchive{RunID: runID}
	err := r.DB.QueryRow("SELECT playbook_id, account_id, filename, created_at, state FROM playbook_archive WHERE run_id=$1",
		p.RunID).Scan(&p.PlaybookID, &p.AccountID, &p.Filename, &p.CreatedAt, &p.State)

	return p, err
}

func (r *PlaybookArchiveRepository) DeletePlaybookArchive(p *domain.PlaybookArchive) error {
	_, err := r.DB.Exec("DELETE FROM playbook_archive WHERE playbook_id=$1", p.PlaybookID)

	return err
}

func (r *PlaybookArchiveRepository) CreatePlaybookArchive(p *domain.PlaybookArchive) error {
	_, err := r.DB.Exec("INSERT INTO playbook_archive(playbook_id, account_id, run_id, filename, created_at, state) VALUES($1, $2, $3, $4, $5, $6)",
		p.PlaybookID, p.AccountID, p.RunID, p.Filename, p.CreatedAt, p.State)

	return err
}
