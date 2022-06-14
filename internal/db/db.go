package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Open opens a database specified by dataSourceName using driverName.
//
// Open adheres to all database/sql driver expectations. For example, it is an
// error to request a dataSourceName of ":memory:" with the "pgx" driver.
func Open(driverName, dataSourceName string) error {
	handle, err := sqlx.Open(driverName, dataSourceName)
	if err != nil {
		return fmt.Errorf("cannot open database: %w", err)
	}

	if err := handle.Ping(); err != nil {
		return fmt.Errorf("cannot ping database: %w", err)
	}

	stddb = &dB{
		handle:     handle,
		statements: make(map[string]*sqlx.Stmt),
		driverName: driverName,
	}

	return nil
}

func Close() error {
	return stddb.Close()
}

func InsertProfile(profile Profile) error {
	return stddb.insertProfile(profile)
}

func GetCurrentProfile(accountID string) (*Profile, error) {
	return stddb.getCurrentProfile(accountID)
}

func GetOrInsertCurrentProfile(accountID string, newProfile *Profile) (*Profile, error) {
	return stddb.getOrInsertCurrentProfile(accountID, newProfile)
}

func GetProfile(profileID string) (*Profile, error) {
	return stddb.getProfile(profileID)
}

func GetProfiles(accountID string, orderBy string, limit int, offset int) ([]Profile, error) {
	return stddb.getProfiles(accountID, orderBy, limit, offset)
}

func CountProfiles(accountID string) (int, error) {
	return stddb.countProfiles(accountID)
}

func Migrate(migrationsPath string, reset bool) error {
	return stddb.migrate(migrationsPath, reset)
}

func Seed(path string) error {
	return stddb.seed(path)
}
