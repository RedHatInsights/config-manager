package persistence

import (
	"config-manager/domain"
	"database/sql"
	"fmt"
)

type PlaybookArchiveRepository struct {
	DB *sql.DB
}

func (r *PlaybookArchiveRepository) GetPlaybookArchive(pb *domain.PlaybookArchive) (*domain.PlaybookArchive, error) {
	err := r.DB.QueryRow("SELECT account_id, playbook FROM playbook_archive WHERE state_id=$1",
		pb.StateID).Scan(&pb.AccountID, &pb.Playbook)

	return pb, err
}

func (r *PlaybookArchiveRepository) DeletePlaybookArchive(pb *domain.PlaybookArchive) error {
	_, err := r.DB.Exec("DELETE FROM playbook_archive WHERE state_id=$1", pb.StateID)

	return err
}

func (r *PlaybookArchiveRepository) CreatePlaybookArchive(pb *domain.PlaybookArchive) error {
	fmt.Println("Creating playbook entry into db")
	_, err := r.DB.Exec("INSERT INTO playbook_archive(account_id, state_id, playbook) VALUES($1, $2, $3)",
		pb.AccountID, pb.StateID, pb.Playbook)

	return err
}
