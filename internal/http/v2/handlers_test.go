package v2

import (
	"bytes"
	"config-manager/internal/db"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
)

type request struct {
	method, url string
	body        []byte
	headers     map[string]string
}

type response struct {
	code int
	body []byte
}

const (
	UNIXTime string = "1970-01-01T00:00:00Z00"
)

var (
	DSN  string
	port uint32
)

func TestMain(m *testing.M) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	port = uint32(r.Int31n(10000-9876) + 9876)
	DSN = fmt.Sprintf("host=localhost port=%v user=postgres password=postgres dbname=postgres sslmode=disable", port)

	runtimedir, err := os.MkdirTemp("", "config-manager-internal-http-v2.")
	if err != nil {
		log.Fatalf("cannot make temp dir: %v", err)
	}
	postgres := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().Port(port).RuntimePath(runtimedir))

	if err := postgres.Start(); err != nil {
		log.Fatalf("failed to start database: %v", err)
	}

	code := m.Run()

	if err := postgres.Stop(); err != nil {
		log.Fatalf("failed to stop database: %v", err)
	}

	if err := os.RemoveAll(runtimedir); err != nil {
		log.Fatalf("cannot remove temp dir: %v", err)
	}

	os.Exit(code)
}

func TestGetProfile(t *testing.T) {
	tests := []struct {
		description string
		seed        []byte
		input       request
		want        response
	}{
		{
			description: "get profile by ID",
			seed:        []byte(`INSERT INTO profiles (profile_id, account_id, org_id, created_at, insights, remediations, compliance) VALUES ('b5db9cbc-4ecd-464b-b416-3a6cd67af87a', '10064', '78606', '` + UNIXTime + `', FALSE, FALSE, FALSE), ('3c8859ae-ef4e-4136-ab17-ccd4ea9f36bf', '10064', '78606', '` + UNIXTime + `', TRUE, TRUE, TRUE);`),
			input: request{
				method: http.MethodGet,
				url:    "/profiles/b5db9cbc-4ecd-464b-b416-3a6cd67af87a",
				headers: map[string]string{
					"X-Rh-Identity": base64.StdEncoding.EncodeToString([]byte(`{"identity":{"account_number":"10064","auth_type":"basic","employee_account_number":"10064","internal":{"org_id":"78606"},"org_id":"78606","type":"User","user":{"email":"collett@elfreda.name","first_name":"Maricela","is_active":true,"is_internal":false,"is_org_admin":true,"last_name":"Purdy","locale":"pa","user_id":"algae","username":"torque"}}}`)),
				},
			},
			want: response{
				code: http.StatusOK,
				body: []byte(`{"id":"b5db9cbc-4ecd-464b-b416-3a6cd67af87a","account_id":"10064","org_id":"78606","created_at":"1970-01-01T00:00:00Z","active":false,"insights":false,"remediations":false,"compliance":false}`),
			},
		},
		{
			description: "get profile by current",
			seed:        []byte(`INSERT INTO profiles (profile_id, account_id, org_id, created_at, insights, remediations, compliance) VALUES ('b5db9cbc-4ecd-464b-b416-3a6cd67af87a', '10064', '78606', '` + UNIXTime + `', FALSE, FALSE, FALSE), ('3c8859ae-ef4e-4136-ab17-ccd4ea9f36bf', '10064', '78606', '` + UNIXTime + `', TRUE, TRUE, TRUE);`),
			input: request{
				method: http.MethodGet,
				url:    "/profiles/current",
				headers: map[string]string{
					"X-Rh-Identity": base64.StdEncoding.EncodeToString([]byte(`{"identity":{"account_number":"10064","auth_type":"basic","employee_account_number":"10064","internal":{"org_id":"78606"},"org_id":"78606","type":"User","user":{"email":"collett@elfreda.name","first_name":"Maricela","is_active":true,"is_internal":false,"is_org_admin":true,"last_name":"Purdy","locale":"pa","user_id":"algae","username":"torque"}}}`)),
				},
			},
			want: response{
				code: http.StatusOK,
				body: []byte(`{"id":"b5db9cbc-4ecd-464b-b416-3a6cd67af87a","account_id":"10064","org_id":"78606","created_at":"1970-01-01T00:00:00Z","active":false,"insights":false,"remediations":false,"compliance":false}`),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if err := db.Open("pgx", DSN); err != nil {
				t.Fatalf("failed to open database: %v", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Fatalf("failed to close database: %v", err)
				}
			}()

			if err := db.Migrate(true); err != nil {
				t.Fatalf("failed to migrate database: %v", err)
			}

			if err := db.SeedData(test.seed); err != nil {
				t.Fatalf("failed to seed database: %v", err)
			}

			reader := bytes.NewReader(test.input.body)
			req := httptest.NewRequest(test.input.method, test.input.url, reader)
			for k, v := range test.input.headers {
				req.Header.Add(k, v)
			}
			rr := httptest.NewRecorder()

			router := chi.NewMux()
			router.Use(identity.EnforceIdentity)
			router.Get("/profiles/{id}", getProfile)
			router.ServeHTTP(rr, req)

			got := response{rr.Code, bytes.TrimSpace(rr.Body.Bytes())}

			if !cmp.Equal(got, test.want, cmp.AllowUnexported(response{})) {
				t.Errorf("%v", cmp.Diff(got, test.want, cmp.AllowUnexported(response{})))
			}
		})
	}
}

func TestCreateProfile(t *testing.T) {
	type response struct {
		code int
		body map[string]interface{}
	}
	tests := []struct {
		description      string
		seed             []byte
		ignoreMapEntries func(k string, v interface{}) bool
		input            request
		want             response
	}{
		{
			description: "new profile values",
			seed:        []byte(`INSERT INTO profiles (profile_id, account_id, org_id, created_at, insights, remediations, compliance) VALUES ('b5db9cbc-4ecd-464b-b416-3a6cd67af87a', '10064', '78606', '` + UNIXTime + `', FALSE, FALSE, FALSE);`),
			ignoreMapEntries: func(k string, v interface{}) bool {
				return k == "id" || k == "label" || k == "created_at"
			},
			input: request{
				method: http.MethodGet,
				url:    "/profiles",
				body:   []byte(`{"active":true,"insights":true,"compliance":true,"remediations":true}`),
				headers: map[string]string{
					"X-Rh-Identity": base64.StdEncoding.EncodeToString([]byte(`{"identity":{"account_number":"10064","auth_type":"basic","employee_account_number":"10064","internal":{"org_id":"78606"},"org_id":"78606","type":"User","user":{"email":"collett@elfreda.name","first_name":"Maricela","is_active":true,"is_internal":false,"is_org_admin":true,"last_name":"Purdy","locale":"pa","user_id":"algae","username":"torque"}}}`)),
				},
			},
			want: response{
				code: http.StatusCreated,
				body: map[string]interface{}{
					"account_id":   "10064",
					"org_id":       "78606",
					"insights":     true,
					"compliance":   true,
					"remediations": true,
					"active":       true,
				},
			},
		},
		{
			description: "identical profile values",
			seed:        []byte(`INSERT INTO profiles (profile_id, account_id, org_id, created_at, insights, remediations, compliance) VALUES ('b5db9cbc-4ecd-464b-b416-3a6cd67af87a', '10064', '78606', '` + UNIXTime + `', FALSE, FALSE, FALSE);`),
			ignoreMapEntries: func(k string, v interface{}) bool {
				return k == "label" || k == "created_at"
			},
			input: request{
				method: http.MethodGet,
				url:    "/profiles",
				body:   []byte(`{"active":false,"insights":false,"compliance":false,"remediations":false}`),
				headers: map[string]string{
					"X-Rh-Identity": base64.StdEncoding.EncodeToString([]byte(`{"identity":{"account_number":"10064","auth_type":"basic","employee_account_number":"10064","internal":{"org_id":"78606"},"org_id":"78606","type":"User","user":{"email":"collett@elfreda.name","first_name":"Maricela","is_active":true,"is_internal":false,"is_org_admin":true,"last_name":"Purdy","locale":"pa","user_id":"algae","username":"torque"}}}`)),
				},
			},
			want: response{
				code: http.StatusNotModified,
				body: map[string]interface{}{
					"account_id":   "10064",
					"org_id":       "78606",
					"id":           "b5db9cbc-4ecd-464b-b416-3a6cd67af87a",
					"insights":     false,
					"compliance":   false,
					"remediations": false,
					"active":       false,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if err := db.Open("pgx", DSN); err != nil {
				t.Fatalf("failed to open database: %v", err)
			}
			defer func() {
				if err := db.Close(); err != nil {
					t.Fatalf("failed to close database: %v", err)
				}
			}()

			if err := db.Migrate(true); err != nil {
				t.Fatalf("failed to migrate database: %v", err)
			}

			if err := db.SeedData(test.seed); err != nil {
				t.Fatalf("failed to seed database: %v", err)
			}

			reader := bytes.NewReader(test.input.body)
			req := httptest.NewRequest(test.input.method, test.input.url, reader)
			for k, v := range test.input.headers {
				req.Header.Add(k, v)
			}
			rr := httptest.NewRecorder()

			router := chi.NewMux()
			router.Use(identity.EnforceIdentity)
			router.Get("/profiles", createProfile)
			router.ServeHTTP(rr, req)

			got := response{rr.Code, map[string]interface{}{}}

			if got.code != test.want.code {
				t.Fatalf("%v != %v (%v)", got.code, test.want.code, rr.Body.String())
			}

			if err := json.Unmarshal(rr.Body.Bytes(), &got.body); err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(got, test.want, cmp.AllowUnexported(response{}), cmpopts.IgnoreMapEntries(test.ignoreMapEntries)) {
				t.Errorf("%v", cmp.Diff(got, test.want, cmp.AllowUnexported(response{})))
			}
		})
	}
}
