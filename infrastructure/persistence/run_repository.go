package persistence

import (
	"config-manager/domain"
	"database/sql"
)

type RunRepository struct {
	DB *sql.DB
}

func (r *RunRepository) GetRun(id string) (*domain.Run, error) {
	return nil, nil
}

func (r *RunRepository) UpdateRun(run *domain.Run) error {
	return nil
}

func (r *RunRepository) DeleteRun(run *domain.Run) error {
	return nil
}

func (r *RunRepository) CreateRun(run *domain.Run) error {
	return nil
}
