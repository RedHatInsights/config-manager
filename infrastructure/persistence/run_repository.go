package persistence

import (
	"config-manager/domain"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type RunRepository struct {
	DB *sql.DB
}

func (r *RunRepository) GetRun(id uuid.UUID) (*domain.Run, error) {
	run := &domain.Run{RunID: id}
	err := r.DB.QueryRow("SELECT account_id, initiator, label, status, created_at FROM runs WHERE run_id=$1",
		run.RunID).Scan(&run.AccountID, &run.Initiator, &run.Label, &run.Status, &run.CreatedAt)

	return run, err
}

func (r *RunRepository) GetRuns(accountID string, limit, offset int) ([]domain.Run, error) {
	rows, err := r.DB.Query("SELECT run_id, account_id, initiator, label, status, created_at FROM runs WHERE account_id=$1 LIMIT $2 OFFSET $3",
		accountID, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	runs := []domain.Run{}

	for rows.Next() {
		var run domain.Run
		if err := rows.Scan(&run.RunID, &run.AccountID, &run.Initiator, &run.Label, &run.Status, &run.CreatedAt); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}

	return runs, err
}

func (r *RunRepository) UpdateRunStatus(run *domain.Run) error {
	_, err := r.DB.Exec("UPDATE runs SET status=$1 WHERE run_id=$2",
		run.Status, run.RunID)

	return err
}

func (r *RunRepository) CreateRun(run *domain.Run) error {
	fmt.Println("Creating new run entry")
	_, err := r.DB.Exec("INSERT INTO runs(run_id, account_id, initiator, label, status, created_at) VALUES($1, $2, $3, $4, $5, $6)",
		run.RunID, run.AccountID, run.Initiator, run.Label, run.Status, run.CreatedAt)

	return err
}
