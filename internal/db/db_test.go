package db

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	UNIXTime string = "1970-01-01T00:00:00Z"
	DSN      string = "host=localhost port=9876 user=postgres password=postgres dbname=postgres sslmode=disable"
)

func TestMain(m *testing.M) {
	postgres := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().Port(9876))

	if err := postgres.Start(); err != nil {
		log.Fatalf("failed to start database: %v", err)
	}

	code := m.Run()

	if err := postgres.Stop(); err != nil {
		log.Fatalf("failed to stop database: %v", err)
	}

	os.Exit(code)
}

func TestInsertProfile(t *testing.T) {
	tests := []struct {
		description string
		input       Profile
		want        error
	}{
		{
			input: *NewProfile("1", map[string]string{"insights": "enabled", "remediations": "enabled", "compliance_openscap": "enabled"}),
			want:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			db, err := Open("pgx", DSN)
			if err != nil {
				t.Fatalf("failed to open database: %v", err)
			}
			defer db.Close()

			if err := db.Migrate("file://../../db/migrations", true); err != nil {
				t.Fatalf("failed to migrate database: %v", err)
			}

			if err := db.InsertProfile(test.input); err != test.want {
				t.Errorf("got error: %v", err)
			}
		})
	}
}

func TestGetCurrentProfile(t *testing.T) {
	tests := []struct {
		description string
		seed        []byte
		input       string
		want        *Profile
	}{
		{
			seed:  []byte(`INSERT INTO profiles (profile_id, account_id, created_at) VALUES ('84d3724c-1944-41d1-a12a-235eddca7771', '1', '` + UNIXTime + `');`),
			input: "1",
			want: &Profile{
				ID:        uuid.MustParse("84d3724c-1944-41d1-a12a-235eddca7771"),
				AccountID: sql.NullString{Valid: true, String: "1"},
				CreatedAt: time.Unix(0, 0),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			db, err := Open("pgx", DSN)
			if err != nil {
				t.Fatalf("failed to open database: %v", err)
			}
			defer db.Close()

			if err := db.Migrate("file://../../db/migrations", true); err != nil {
				t.Fatalf("failed to migrate database: %v", err)
			}

			if err := db.seedData(test.seed); err != nil {
				t.Fatalf("failed to seed database: %v", err)
			}

			got, err := db.GetCurrentProfile(test.input)
			if err != nil {
				t.Fatalf("failed to get current profile: %v", err)
			}

			if !cmp.Equal(got, test.want) {
				t.Errorf("---got\n+++want\n%v", cmp.Diff(got, test.want))
			}
		})
	}
}

func TestGetProfile(t *testing.T) {
	tests := []struct {
		description string
		seed        []byte
		input       string
		want        *Profile
	}{
		{
			seed:  []byte(`INSERT INTO profiles (profile_id, account_id, created_at) VALUES ('84d3724c-1944-41d1-a12a-235eddca7771', '1', '` + UNIXTime + `');`),
			input: "84d3724c-1944-41d1-a12a-235eddca7771",
			want: &Profile{
				ID:        uuid.MustParse("84d3724c-1944-41d1-a12a-235eddca7771"),
				AccountID: sql.NullString{Valid: true, String: "1"},
				CreatedAt: time.Unix(0, 0),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			db, err := Open("pgx", DSN)
			if err != nil {
				t.Fatalf("failed to open database: %v", err)
			}
			defer db.Close()

			if err := db.Migrate("file://../../db/migrations", true); err != nil {
				t.Fatalf("failed to migrate database: %v", err)
			}

			if err := db.seedData(test.seed); err != nil {
				t.Fatalf("failed to seed database: %v", err)
			}

			got, err := db.GetProfile(test.input)
			if err != nil {
				t.Fatalf("failed to get profile: %v", err)
			}

			if !cmp.Equal(got, test.want) {
				t.Errorf("---got\n+++want\n%v", cmp.Diff(got, test.want))
			}
		})
	}
}

func TestGetProfiles(t *testing.T) {
	tests := []struct {
		description string
		seed        []byte
		input       struct {
			accountID string
			orderBy   string
			limit     int
			offset    int
		}
		want []Profile
	}{
		{
			seed: []byte(`INSERT INTO profiles (profile_id, account_id, created_at, insights, remediations, compliance) VALUES ('b5db9cbc-4ecd-464b-b416-3a6cd67af87a', '1', '` + UNIXTime + `', FALSE, FALSE, FALSE), ('3c8859ae-ef4e-4136-ab17-ccd4ea9f36bf', '1', '` + UNIXTime + `', TRUE, TRUE, TRUE);`),
			input: struct {
				accountID string
				orderBy   string
				limit     int
				offset    int
			}{
				accountID: "1",
				orderBy:   "",
				limit:     -1,
				offset:    -1,
			},
			want: []Profile{
				{
					ID:           uuid.MustParse("b5db9cbc-4ecd-464b-b416-3a6cd67af87a"),
					AccountID:    sql.NullString{Valid: true, String: "1"},
					CreatedAt:    time.Unix(0, 0),
					Insights:     false,
					Remediations: false,
					Compliance:   false,
				},
				{
					ID:           uuid.MustParse("3c8859ae-ef4e-4136-ab17-ccd4ea9f36bf"),
					AccountID:    sql.NullString{Valid: true, String: "1"},
					CreatedAt:    time.Unix(0, 0),
					Insights:     true,
					Remediations: true,
					Compliance:   true,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			db, err := Open("pgx", DSN)
			if err != nil {
				t.Fatalf("failed to open database: %v", err)
			}
			defer db.Close()

			if err := db.Migrate("file://../../db/migrations", true); err != nil {
				t.Fatalf("failed to migrate database: %v", err)
			}

			if err := db.seedData(test.seed); err != nil {
				t.Fatalf("failed to seed database: %v", err)
			}

			got, err := db.GetProfiles(test.input.accountID, test.input.orderBy, test.input.limit, test.input.offset)
			if err != nil {
				t.Fatalf("failed to get profile: %v", err)
			}

			if !cmp.Equal(got, test.want) {
				t.Errorf("---got\n+++want\n%v", cmp.Diff(got, test.want))
			}
		})
	}
}

func TestCountProfiles(t *testing.T) {
	tests := []struct {
		description string
		seed        []byte
		input       string
		want        int
	}{
		{
			seed:  []byte(`INSERT INTO profiles (profile_id, account_id, created_at, insights, remediations, compliance) VALUES ('b5db9cbc-4ecd-464b-b416-3a6cd67af87a', '1', '` + UNIXTime + `', FALSE, FALSE, FALSE), ('3c8859ae-ef4e-4136-ab17-ccd4ea9f36bf', '1', '` + UNIXTime + `', TRUE, TRUE, TRUE);`),
			input: "1",
			want:  2,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			db, err := Open("pgx", DSN)
			if err != nil {
				t.Fatalf("failed to open database: %v", err)
			}
			defer db.Close()

			if err := db.Migrate("file://../../db/migrations", true); err != nil {
				t.Fatalf("failed to migrate database: %v", err)
			}

			if err := db.seedData(test.seed); err != nil {
				t.Fatalf("failed to seed database: %v", err)
			}

			got, err := db.CountProfiles(test.input)
			if err != nil {
				t.Fatalf("failed to get profile: %v", err)
			}

			if !cmp.Equal(got, test.want) {
				t.Errorf("---got\n+++want\n%v", cmp.Diff(got, test.want))
			}
		})
	}
}

func TestParseOrderBy(t *testing.T) {
	type orderBy struct {
		column    string
		direction string
	}

	tests := []struct {
		description string
		input       string
		want        orderBy
		wantError   error
	}{
		{
			description: "empty input",
			input:       "",
			want: orderBy{
				column:    "",
				direction: "",
			},
		},
		{
			description: "only column",
			input:       "created_at",
			want: orderBy{
				column:    "created_at",
				direction: "",
			},
		},
		{
			description: "empty direction",
			input:       "created_at:",
			want: orderBy{
				column:    "created_at",
				direction: "",
			},
		},
		{
			description: "column and direction",
			input:       "last_updated:asc",
			want: orderBy{
				column:    "last_updated",
				direction: "ASC",
			},
		},
		{
			description: "multiple columns",
			input:       "created_at,last_updated:desc",
			want: orderBy{
				column:    "(created_at, last_updated)",
				direction: "DESC",
			},
		},
		{
			description: "invalid column name",
			input:       "created_at;TRUNCATE",
			wantError:   ParseError{msg: "invalid column field: created_at;TRUNCATE"},
		},
		{
			description: "invalid direction",
			input:       "created_at:;truncate",
			wantError:   ParseError{msg: "invalid order direction: ;TRUNCATE"},
		},
		{
			description: "invalid format",
			input:       "field_a:field_b:field_c",
			wantError:   ParseError{msg: "cannot parse order: field_a:field_b:field_c"},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var err error
			got := orderBy{}

			got.column, got.direction, err = parseOrderBy(test.input)

			if test.wantError != nil {
				if !cmp.Equal(err, test.wantError, cmpopts.EquateErrors()) {
					t.Errorf("%#v != %#v", err, test.wantError)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(got, test.want, cmp.AllowUnexported(orderBy{})) {
					t.Errorf("%#v != %#v", got, test.want)
				}
			}
		})
	}
}
