package v1

import (
	"bytes"
	"config-manager/internal/config"
	"config-manager/internal/db"
	"config-manager/internal/http/staticmux"
	"config-manager/internal/url"
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
	"github.com/redhatinsights/platform-go-middlewares/identity"

	_ "github.com/golang-migrate/migrate/v4/source/file"
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
	rand.Seed(time.Now().UnixNano())
	port = uint32(rand.Int31n(10000-9876) + 9876)
	DSN = fmt.Sprintf("host=localhost port=%v user=postgres password=postgres dbname=postgres sslmode=disable", port)

	runtimedir, err := os.MkdirTemp("", "config-manager-internal-http-v1.")
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

func TestGetStates(t *testing.T) {
	tests := []struct {
		description string
		seed        []byte
		input       request
		want        response
	}{
		{
			seed: []byte(`INSERT INTO profiles (profile_id, account_id, org_id, created_at, insights, remediations, compliance) VALUES ('b5db9cbc-4ecd-464b-b416-3a6cd67af87a', '10064', '78606', '` + UNIXTime + `', FALSE, FALSE, FALSE), ('3c8859ae-ef4e-4136-ab17-ccd4ea9f36bf', '10064', '78606', '` + UNIXTime + `', TRUE, TRUE, TRUE);`),
			input: request{
				method: http.MethodGet,
				url:    "/states",
				headers: map[string]string{
					"X-Rh-Identity": base64.StdEncoding.EncodeToString([]byte(`{"identity":{"account_number":"10064","auth_type":"basic","employee_account_number":"10064","internal":{"org_id":"78606"},"org_id":"78606","type":"User","user":{"email":"collett@elfreda.name","first_name":"Maricela","is_active":true,"is_internal":false,"is_org_admin":true,"last_name":"Purdy","locale":"pa","user_id":"algae","username":"torque"}}}`)),
				},
			},
			want: response{
				code: http.StatusOK,
				body: []byte(`{"count":2,"limit":0,"offset":0,"total":2,"results":[{"account":"10064","id":"b5db9cbc-4ecd-464b-b416-3a6cd67af87a","created_at":"1970-01-01T00:00:00Z","state":{"compliance_openscap":"disabled","insights":"disabled","remediations":"disabled"},"org_id":"78606"},{"account":"10064","id":"3c8859ae-ef4e-4136-ab17-ccd4ea9f36bf","created_at":"1970-01-01T00:00:00Z","state":{"compliance_openscap":"enabled","insights":"enabled","remediations":"enabled"},"org_id":"78606"}]}`),
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
			router.Get("/states", getStates)
			router.ServeHTTP(rr, req)

			got := response{rr.Code, bytes.TrimSpace(rr.Body.Bytes())}

			if !cmp.Equal(got, test.want, cmp.AllowUnexported(response{})) {
				t.Errorf("%v", cmp.Diff(got, test.want, cmp.AllowUnexported(response{})))
			}
		})
	}
}

func TestPostStates(t *testing.T) {
	type response struct {
		code int
		body map[string]interface{}
	}

	tests := []struct {
		description string
		seed        []byte
		input       request
		want        response
	}{
		{
			seed: []byte(`INSERT INTO profiles (profile_id, account_id, org_id, created_at, insights, remediations, compliance) VALUES ('b5db9cbc-4ecd-464b-b416-3a6cd67af87a', '10064', '78606', '` + UNIXTime + `', FALSE, FALSE, FALSE), ('3c8859ae-ef4e-4136-ab17-ccd4ea9f36bf', '10064', '78606', '` + UNIXTime + `', TRUE, TRUE, TRUE);`),
			input: request{
				method: http.MethodPost,
				url:    "/states",
				body:   []byte(`{"insights":"enabled","remediations":"enabled","compliance_openscap":"enabled"}`),
				headers: map[string]string{
					"X-Rh-Identity": base64.StdEncoding.EncodeToString([]byte(`{"identity":{"account_number":"10064","auth_type":"basic","employee_account_number":"10064","internal":{"org_id":"78606"},"org_id":"78606","type":"User","user":{"email":"collett@elfreda.name","first_name":"Maricela","is_active":true,"is_internal":false,"is_org_admin":true,"last_name":"Purdy","locale":"pa","user_id":"algae","username":"torque"}}}`)),
				},
			},
			want: response{
				code: http.StatusOK,
				body: map[string]interface{}{
					"account":     "10064",
					"apply_state": false,
					"id":          "b5db9cbc-4ecd-464b-b416-3a6cd67af87a",
					"label":       "",
					"org_id":      "78606",
					"state": map[string]interface{}{
						"compliance_openscap": "enabled",
						"insights":            "enabled",
						"remediations":        "enabled",
					},
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

			config.DefaultConfig.InventoryHost.Value = url.MustParse("http://localhost:8000")
			staticInventoryMux := staticmux.StaticMux{}
			staticInventoryMux.AddResponse("/api/inventory/v1/hosts", 200, []byte(`{"count":1,"limit":0","offset":0","total":1,"page":1,"per_page":50","results":[{"id":"6a46563f-6c26-449a-89c4-de902d8c5ceb","account":"10064","org_id":"78606","display_name":"test","reporter":"test","system_profile":{"rhc_client_id":"7eb87461-a49b-4ce6-8042-2494200f6bf6","rhc_config_state":"connected"}}]}`), map[string][]string{"Content-Type": {"application/json"}})
			go func() {
				if err := http.ListenAndServe(config.DefaultConfig.InventoryHost.Value.Host, &staticInventoryMux); err != nil {
					log.Print(err)
				}
			}()

			reader := bytes.NewReader(test.input.body)
			req := httptest.NewRequest(test.input.method, test.input.url, reader)
			for k, v := range test.input.headers {
				req.Header.Add(k, v)
			}
			rr := httptest.NewRecorder()

			router := chi.NewMux()
			router.Use(identity.EnforceIdentity)
			router.Post("/states", postStates)
			router.ServeHTTP(rr, req)

			got := response{rr.Code, make(map[string]interface{})}

			if got.code != test.want.code {
				t.Fatalf("%v != %v (%v)", got.code, test.want.code, rr.Body.String())
			}

			if err := json.Unmarshal(rr.Body.Bytes(), &got.body); err != nil {
				t.Fatal(err)
			}

			ignoreMapEntriesOpt := cmpopts.IgnoreMapEntries(func(k string, v interface{}) bool {
				return k == "label" || k == "id"
			})

			if !cmp.Equal(got, test.want, cmp.AllowUnexported(response{}), ignoreMapEntriesOpt) {
				t.Errorf("%v", cmp.Diff(got, test.want, cmp.AllowUnexported(response{}), ignoreMapEntriesOpt))
			}
		})
	}
}

func TestGetCurrentState(t *testing.T) {
	tests := []struct {
		description string
		seed        []byte
		input       request
		want        response
	}{
		{
			seed: []byte(`INSERT INTO profiles (profile_id, account_id, org_id, created_at, insights, remediations, compliance) VALUES ('b5db9cbc-4ecd-464b-b416-3a6cd67af87a', '1', '78606', '` + UNIXTime + `', FALSE, FALSE, FALSE), ('3c8859ae-ef4e-4136-ab17-ccd4ea9f36bf', '1', '78606', '` + time.Now().Format(time.RFC3339) + `', TRUE, TRUE, TRUE);`),
			input: request{
				method: http.MethodGet,
				url:    "/states/current",
				headers: map[string]string{
					"X-Rh-Identity": base64.StdEncoding.EncodeToString([]byte(`{"identity":{"account_number":"10064","auth_type":"basic","employee_account_number":"10064","internal":{"org_id":"78606"},"org_id":"78606","type":"User","user":{"email":"collett@elfreda.name","first_name":"Maricela","is_active":true,"is_internal":false,"is_org_admin":true,"last_name":"Purdy","locale":"pa","user_id":"algae","username":"torque"}}}`)),
				},
			},
			want: response{
				code: http.StatusOK,
				body: []byte(`{"account":"1","state":{"compliance_openscap":"enabled","insights":"enabled","remediations":"enabled"},"id":"3c8859ae-ef4e-4136-ab17-ccd4ea9f36bf","apply_state":false,"org_id":"78606"}`),
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

			router := chi.NewRouter()
			router.Use(identity.EnforceIdentity)
			router.Get("/states/current", getCurrentState)
			router.ServeHTTP(rr, req)

			got := response{rr.Code, bytes.TrimSpace(rr.Body.Bytes())}

			if !cmp.Equal(got, test.want, cmp.AllowUnexported(response{})) {
				t.Errorf("%v", cmp.Diff(got, test.want, cmp.AllowUnexported(response{})))
			}
		})
	}
}

func TestGetStateByID(t *testing.T) {
	tests := []struct {
		description string
		seed        []byte
		input       request
		want        response
	}{
		{
			seed: []byte(`INSERT INTO profiles (profile_id, account_id, org_id, created_at, insights, remediations, compliance) VALUES ('b5db9cbc-4ecd-464b-b416-3a6cd67af87a', '10064', '78606', '` + UNIXTime + `', FALSE, FALSE, FALSE), ('3c8859ae-ef4e-4136-ab17-ccd4ea9f36bf', '10064', '78606', '` + UNIXTime + `', TRUE, TRUE, TRUE);`),
			input: request{
				method: http.MethodGet,
				url:    "/states/b5db9cbc-4ecd-464b-b416-3a6cd67af87a",
				headers: map[string]string{
					"X-Rh-Identity": base64.StdEncoding.EncodeToString([]byte(`{"identity":{"account_number":"10064","auth_type":"basic","employee_account_number":"10064","internal":{"org_id":"78606"},"org_id":"78606","type":"User","user":{"email":"collett@elfreda.name","first_name":"Maricela","is_active":true,"is_internal":false,"is_org_admin":true,"last_name":"Purdy","locale":"pa","user_id":"algae","username":"torque"}}}`)),
				},
			},
			want: response{
				code: http.StatusOK,
				body: []byte(`{"account":"10064","id":"b5db9cbc-4ecd-464b-b416-3a6cd67af87a","created_at":"1970-01-01T00:00:00Z","state":{"compliance_openscap":"disabled","insights":"disabled","remediations":"disabled"},"org_id":"78606"}`),
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
			router.Get("/states/{id}", getStateByID)
			router.ServeHTTP(rr, req)

			got := response{rr.Code, bytes.TrimSpace(rr.Body.Bytes())}

			if !cmp.Equal(got, test.want, cmp.AllowUnexported(response{})) {
				t.Errorf("%v", cmp.Diff(got, test.want, cmp.AllowUnexported(response{})))
			}
		})
	}
}

func TestPostManage(t *testing.T) {
	tests := []struct {
		description string
		seed        []byte
		input       request
		want        response
	}{
		{
			seed: []byte(`INSERT INTO profiles (profile_id, account_id, org_id, created_at, insights, remediations, compliance) VALUES ('b5db9cbc-4ecd-464b-b416-3a6cd67af87a', '10064', '78606', '` + UNIXTime + `', FALSE, FALSE, FALSE);`),
			input: request{
				method: http.MethodPost,
				url:    "/manage",
				headers: map[string]string{
					"X-Rh-Identity": base64.StdEncoding.EncodeToString([]byte(`{"identity":{"account_number":"10064","auth_type":"basic","employee_account_number":"10064","internal":{"org_id":"78606"},"org_id":"78606","type":"User","user":{"email":"collett@elfreda.name","first_name":"Maricela","is_active":true,"is_internal":false,"is_org_admin":true,"last_name":"Purdy","locale":"pa","user_id":"algae","username":"torque"}}}`)),
				},
				body: []byte{'t', 'r', 'u', 'e'},
			},
			want: response{
				code: http.StatusOK,
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
			router.Post("/manage", postManage)
			router.ServeHTTP(rr, req)

			got := response{rr.Code, rr.Body.Bytes()}

			if !cmp.Equal(got, test.want, cmp.AllowUnexported(response{})) {
				t.Errorf("%v", cmp.Diff(got, test.want, cmp.AllowUnexported(response{})))
			}
		})
	}
}

func TestGetStatesIDPlaybook(t *testing.T) {
	tests := []struct {
		description string
		seed        []byte
		input       request
		want        response
	}{
		{
			seed: []byte(`INSERT INTO profiles (profile_id, account_id, created_at, insights, remediations, compliance) VALUES ('b5db9cbc-4ecd-464b-b416-3a6cd67af87a', '1', '` + UNIXTime + `', FALSE, FALSE, FALSE);`),
			input: request{
				method: http.MethodGet,
				url:    "/states/b5db9cbc-4ecd-464b-b416-3a6cd67af87a/playbook",
				headers: map[string]string{
					"X-Rh-Identity": base64.StdEncoding.EncodeToString([]byte(`{"identity":{"account_number":"1","auth_type":"basic","employee_account_number":"10064","internal":{"org_id":"78606"},"org_id":"78606","type":"User","user":{"email":"collett@elfreda.name","first_name":"Maricela","is_active":true,"is_internal":false,"is_org_admin":true,"last_name":"Purdy","locale":"pa","user_id":"algae","username":"torque"}}}`)),
				},
				body: []byte{},
			},
			want: response{
				code: http.StatusOK,
				body: []byte("---\n# Service Enablement playbook\n"),
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
			router.Get("/states/{id}/playbook", getStatesIDPlaybook)
			router.ServeHTTP(rr, req)

			got := response{rr.Code, rr.Body.Bytes()}

			if !cmp.Equal(got, test.want, cmp.AllowUnexported(response{})) {
				t.Errorf("%v", cmp.Diff(got, test.want, cmp.AllowUnexported(response{})))
			}
		})
	}
}

func TestPostStatesPreview(t *testing.T) {
	tests := []struct {
		description string
		seed        []byte
		input       request
		want        response
	}{
		{
			seed: []byte(`INSERT INTO profiles (profile_id, account_id, created_at, insights, remediations, compliance) VALUES ('b5db9cbc-4ecd-464b-b416-3a6cd67af87a', '1', '` + UNIXTime + `', FALSE, FALSE, FALSE);`),
			input: request{
				method: http.MethodPost,
				url:    "/states/preview",
				headers: map[string]string{
					"X-Rh-Identity": base64.StdEncoding.EncodeToString([]byte(`{"identity":{"account_number":"1","auth_type":"basic","employee_account_number":"10064","internal":{"org_id":"78606"},"org_id":"78606","type":"User","user":{"email":"collett@elfreda.name","first_name":"Maricela","is_active":true,"is_internal":false,"is_org_admin":true,"last_name":"Purdy","locale":"pa","user_id":"algae","username":"torque"}}}`)),
				},
				body: []byte(`{"compliance_openscap":"enabled","remediations":"disabled","insights":"enabled"}`),
			},
			want: response{
				code: http.StatusOK,
				body: []byte("---\n# Service Enablement playbook\n"),
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
			router.Post("/states/preview", postStatesPreview)
			router.ServeHTTP(rr, req)

			got := response{rr.Code, rr.Body.Bytes()}

			if !cmp.Equal(got, test.want, cmp.AllowUnexported(response{})) {
				t.Errorf("%v", cmp.Diff(got, test.want, cmp.AllowUnexported(response{})))
			}
		})
	}
}
