package domain

// PLACEHOLDER - Need to see what actually gets returned by inventory when requesting connected systems
// NOTE potentially not needed

type System struct {
	RunID      string `db:"run_id"`
	SystemID   string `db:"system_id"`
	SystemName string `db:"system_name"`
	Updated    string `db:"updated"`
	Status     string `db:"status"`
	Logs       string `db:"logs"`
}

type SystemRepository interface {
	GetSystemsByRunID(runID string, limit, offset int) ([]System, error)
	UpdateSystem(s *System) error
	CreateSystem(s *System) error
}
