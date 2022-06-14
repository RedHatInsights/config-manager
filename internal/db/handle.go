package db

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	fields = `profile_id, name, label, account_id, org_id, created_at, active, creator, insights, remediations, compliance`
)

var stddb *dB

// ParseError represents an error that ocurred when parsing URL query parameters
// such as offset, orderBy and limit.
type ParseError struct {
	msg string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("parse error: '%v'", e.msg)
}

// dB wraps a sql.dB handle, providing an application-specific, higher-level API
// around the standard sql.dB interface.
type dB struct {
	handle     *sqlx.DB
	statements map[string]*sqlx.Stmt
	driverName string
}

// Close closes all open prepared statements and returns the connection to the
// connection pool.
func (db *dB) Close() error {
	for _, stmt := range db.statements {
		stmt.Close()
	}
	return db.handle.Close()
}

// Handle returns a handle to the wrapped *sql.DB.
func (db *dB) Handle() *sql.DB {
	return db.handle.DB
}

// insertProfile creates a new record in the profiles table from profile.
func (db *dB) insertProfile(profile Profile) error {
	stmt, err := db.preparedStatement(`INSERT INTO profiles (profile_id, name, label, account_id, org_id, insights, remediations, compliance, active, creator) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);`)
	if err != nil {
		return fmt.Errorf("cannot prepare INSERT: %w", err)
	}

	_, err = stmt.Exec(profile.ID, profile.Name, profile.Label, profile.AccountID, profile.OrgID, profile.Insights, profile.Remediations, profile.Compliance, profile.Active, profile.Creator)
	if err != nil {
		return fmt.Errorf("cannot execute INSERT: %w", err)
	}

	return nil
}

// getCurrentProfile retrieves the current profile for the given account ID from
// the database.
func (db *dB) getCurrentProfile(accountID string) (*Profile, error) {
	query := fmt.Sprintf("SELECT %v FROM profiles WHERE account_id = $1 ORDER BY created_at DESC LIMIT 1;", fields)
	stmt, err := db.preparedStatement(query)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare SELECT: %w", err)
	}

	var profile Profile
	if err := stmt.Get(&profile, accountID); err != nil {
		return nil, fmt.Errorf("cannot execute SELECT: %w", err)
	}

	return &profile, nil
}

// getOrInsertCurrentProfile attempts to retrieve a profile for the account ID.
// If no row is returned, a new one is created and inserted using newProfile as
// a template. A new row is then queried from the database and returned.
func (db *dB) getOrInsertCurrentProfile(accountID string, newProfile *Profile) (*Profile, error) {
	query := fmt.Sprintf("SELECT %v FROM profiles WHERE account_id = $1 ORDER BY created_at DESC LIMIT 1;", fields)
	stmt, err := db.preparedStatement(query)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare SELECT: %w", err)
	}

	var profile Profile
	err = stmt.Get(&profile, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			if err := db.insertProfile(*newProfile); err != nil {
				return nil, fmt.Errorf("cannot perform INSERT: %w", err)
			}
			if err := stmt.Get(&profile, accountID); err != nil {
				return nil, fmt.Errorf("cannot perform SELECT: %w", err)
			}
		} else {
			return nil, fmt.Errorf("cannot execute SELECT: %w", err)
		}
	}

	return &profile, nil
}

// getProfile retrieves the profile for the given profile ID from the database.
func (db *dB) getProfile(profileID string) (*Profile, error) {
	query := fmt.Sprintf("SELECT %v FROM profiles WHERE profile_id = $1;", fields)
	stmt, err := db.preparedStatement(query)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare SELECT: %w", err)
	}

	var profile Profile
	if err := stmt.Get(&profile, profileID); err != nil {
		return nil, fmt.Errorf("cannot execute SELECT: %w", err)
	}

	return &profile, nil
}

// getProfiles retrieves all profiles for the given account ID from the
// database.
func (db *dB) getProfiles(accountID string, orderBy string, limit int, offset int) ([]Profile, error) {
	orderColumn, orderDirection, err := parseOrderBy(orderBy)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT %v FROM profiles WHERE account_id = $1", fields)

	if orderColumn != "" {
		query += " ORDER BY " + orderColumn
	}

	if orderDirection != "" {
		query += " " + orderDirection
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %v", limit)
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %v", offset)
	}

	query += ";"

	stmt, err := db.preparedStatement(query)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare SELECT: %w", err)
	}

	profiles := []Profile{}
	if err := stmt.Select(&profiles, accountID); err != nil {
		return nil, fmt.Errorf("cannot execute SELECT: %w", err)
	}

	return profiles, nil
}

// countProfiles returns a count of all rows with an account_id column equal to
// accountID.
func (db *dB) countProfiles(accountID string) (int, error) {

	query := "SELECT COUNT(*) FROM profiles WHERE account_id = $1;"

	stmt, err := db.preparedStatement(query)
	if err != nil {
		return -1, fmt.Errorf("cannot prepare SELECT: %w", err)
	}

	var count int
	if err := stmt.Get(&count, accountID); err != nil {
		return -1, fmt.Errorf("cannot execute SELECT: %w", err)
	}

	return count, nil
}

// migrate inspects the current active migration version and runs all necessary
// steps to migrate all the way up. If reset is true, everything is deleted in
// the database before applying migrations.
func (db *dB) migrate(migrationsPath string, reset bool) error {
	m, err := newMigrate(db.handle.DB, db.driverName, migrationsPath)
	if err != nil {
		return fmt.Errorf("cannot create migration: %w", err)
	}

	if reset {
		if err := m.Drop(); err != nil {
			return fmt.Errorf("cannot drop database during migration: %w", err)
		}
		// After calling Drop, we need to ensure the schema_migrations table
		// exists. In the postgres driver, an unexported function, ensureVersionTable,
		// is called inside WithInstance. So we just reinitialize m to a new
		// Migrate instance.
		m, err = newMigrate(db.handle.DB, db.driverName, migrationsPath)
		if err != nil {
			return fmt.Errorf("cannot create migration: %w", err)
		}
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return fmt.Errorf("cannot migrate up: %w", err)
	}
	return nil
}

// seed executes the SQL contained in path in order to seed the database.
func (db *dB) seed(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}
	return db.seedData(data)
}

func (db *dB) seedData(data []byte) error {
	_, err := db.handle.Exec(string(data))
	if err != nil {
		return fmt.Errorf("cannot execute seed SQL: %w", err)
	}
	return nil
}

// preparedNamedStatement creates a prepared statement for the given query, caches
// it in a map and returns the prepared statement. If a statement already exists
// for query, the cached statement is returned.
func (db *dB) preparedStatement(query string) (*sqlx.Stmt, error) {
	stmt, has := db.statements[query]
	if has && stmt != nil {
		return stmt, nil
	}
	stmt, err := db.handle.Preparex(query)
	if err != nil {
		return nil, fmt.Errorf("cannot prepare statement: %w", err)
	}
	db.statements[query] = stmt
	return stmt, nil
}

func newMigrate(db *sql.DB, driverName string, migrationsPath string) (*migrate.Migrate, error) {
	var driver database.Driver
	var err error
	switch driverName {
	case "pgx":
		driver, err = postgres.WithInstance(db, &postgres.Config{})
		if err != nil {
			return nil, fmt.Errorf("cannot create database driver: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported driver: %v", driverName)
	}

	m, err := migrate.NewWithDatabaseInstance(migrationsPath, driverName, driver)
	if err != nil {
		return nil, fmt.Errorf("cannot create migration: %w", err)
	}

	return m, nil
}

// parseOrderBy parses a colon-separated orderBy query parameter format and
// returns properly formatted SQL syntax.
func parseOrderBy(orderBy string) (column string, direction string, err error) {
	// validate the column name in an effort to mitigate SQL injection
	r := regexp.MustCompile(`[^a-zA-Z0-9.\-_]+`)

	var columns []string

	order := strings.Split(orderBy, ":")
	switch len(order) {
	case 1:
		columns = strings.Split(order[0], ",")
	case 2:
		columns = strings.Split(order[0], ",")

		switch strings.ToUpper(order[1]) {
		case "":
			direction = ""
		case "ASC":
			direction = "ASC"
		case "DESC":
			direction = "DESC"
		default:
			return "", "", ParseError{msg: fmt.Sprintf("invalid order direction: %v", strings.ToUpper(order[1]))}
		}
	default:
		return "", "", ParseError{msg: fmt.Sprintf("cannot parse order: %v", orderBy)}
	}

	for _, column := range columns {
		if len(column) > 0 {
			if r.MatchString(column) {
				return "", "", ParseError{msg: fmt.Sprintf("invalid column field: %v", column)}
			}
		}
	}

	if len(columns) > 1 {
		column = "(" + strings.Join(columns, ", ") + ")"
	} else {
		column = columns[0]
	}

	return
}
