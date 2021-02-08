package persistence

import (
	"config-manager/domain"
	"database/sql"
)

type StateArchiveRepository struct {
	DB *sql.DB
}

func (r *StateArchiveRepository) GetStateArchive(s *domain.StateArchive) (*domain.StateArchive, error) {
	err := r.DB.QueryRow("SELECT account_id, label, initiator, created_at, state FROM state_archive WHERE state_id=$1",
		s.StateID).Scan(&s.AccountID, &s.Label, &s.Initiator, &s.CreatedAt, &s.State)

	return s, err
}

func (r *StateArchiveRepository) GetAllStateArchives(accountID string, limit, offset int) ([]domain.StateArchive, error) {
	rows, err := r.DB.Query("SELECT account_id, state_id, label, initiator, created_at, state "+
		"FROM state_archive WHERE account_id=$1 LIMIT $2 OFFSET $3", accountID, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	states := []domain.StateArchive{}

	for rows.Next() {
		var state domain.StateArchive
		if err := rows.Scan(&state.AccountID, &state.StateID, &state.Label, &state.Initiator,
			&state.CreatedAt, &state.State); err != nil {
			return nil, err
		}
		states = append(states, state)
	}

	return states, err
}

func (r *StateArchiveRepository) DeleteStateArchive(s *domain.StateArchive) error {
	_, err := r.DB.Exec("DELETE FROM state_archive WHERE state_id=$1", s.StateID)

	return err
}

func (r *StateArchiveRepository) CreateStateArchive(s *domain.StateArchive) error {
	_, err := r.DB.Exec("INSERT INTO State_archive(state_id, account_id, label, initiator, created_at, state) VALUES($1, $2, $3, $4, $5, $6)",
		s.StateID, s.AccountID, s.Label, s.Initiator, s.CreatedAt, s.State)

	return err
}
