package persistence

import (
	"config-manager/domain"
	"database/sql"
)

// NOTE possibly not needed

type SystemArchiveRepository struct {
	DB *sql.DB
}

func (r *SystemArchiveRepository) GetSystemsByRunID(runID string, limit, offset int) ([]domain.System, error) {
	rows, err := r.DB.Query("SELECT run_id, system_id, system_name, updated, status FROM run_systems LIMIT $1 OFFSET $2 WHERE run_id=$3",
		limit, offset, runID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	systems := []domain.System{}

	for rows.Next() {
		var s domain.System
		if err := rows.Scan(&s.RunID, &s.SystemID, &s.SystemName, &s.Updated, &s.Status); err != nil {
			return nil, err
		}
		systems = append(systems, s)
	}

	return systems, nil
}

func (r *SystemArchiveRepository) UpdateSystem(s *domain.System) error {
	_, err := r.DB.Exec("UPDATE run_systems SET updated=$1 status=$2 logs=$3 WHERE system_id=$4",
		s.Updated, s.Status, s.Logs, s.SystemID)

	return err
}

func (r *SystemArchiveRepository) CreateSystem(s *domain.System) error {
	_, err := r.DB.Exec("INSERT INTO run_systems(run_id, system_id, system_name, updated, status, logs) VALUES($1, $2, $3, $4, $5, $6)",
		s.RunID, s.SystemID, s.SystemName, s.Updated, s.Status, s.Logs)

	return err
}
