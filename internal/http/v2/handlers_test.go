package v2

import (
	"bytes"
	"config-manager/internal/config"
	"config-manager/internal/db"
	"config-manager/internal/util"
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

func TestGetProfiles(t *testing.T) {
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
				url:    "/profiles",
				headers: map[string]string{
					"X-Rh-Identity": base64.StdEncoding.EncodeToString([]byte(`{"identity":{"account_number":"10064","auth_type":"basic","employee_account_number":"10064","internal":{"org_id":"78606"},"org_id":"78606","type":"User","user":{"email":"collett@elfreda.name","first_name":"Maricela","is_active":true,"is_internal":false,"is_org_admin":true,"last_name":"Purdy","locale":"pa","user_id":"algae","username":"torque"}}}`)),
				},
			},
			want: response{
				code: http.StatusOK,
				body: []byte(`{"count":2,"limit":0,"offset":0,"total":2,"results":[{"id":"b5db9cbc-4ecd-464b-b416-3a6cd67af87a","account_id":"10064","org_id":"78606","created_at":"1970-01-01T00:00:00Z","active":false,"insights":false,"remediations":false,"compliance":false},{"id":"3c8859ae-ef4e-4136-ab17-ccd4ea9f36bf","account_id":"10064","org_id":"78606","created_at":"1970-01-01T00:00:00Z","active":false,"insights":true,"remediations":true,"compliance":true}]}`),
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
			router.Get("/profiles", getProfiles)
			router.ServeHTTP(rr, req)

			got := response{rr.Code, bytes.TrimSpace(rr.Body.Bytes())}

			if !cmp.Equal(got, test.want, cmp.AllowUnexported(response{})) {
				t.Errorf("%v", cmp.Diff(got, test.want, cmp.AllowUnexported(response{})))
			}
		})
	}
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

func TestPlaybooks(t *testing.T) {
	config.DefaultConfig.PlaybookFiles = "../../../playbooks/"

	tests := []struct {
		description string
		seed        []byte
		input       request
		want        string
	}{
		{
			description: "get playbook by profile_id",
			seed:        []byte(`INSERT INTO profiles (profile_id, account_id, org_id, created_at, insights, remediations, compliance) VALUES ('b5db9cbc-4ecd-464b-b416-3a6cd67af87a', '10064', '78606', '` + UNIXTime + `', FALSE, FALSE, FALSE);`),
			input: request{
				method: http.MethodGet,
				url:    "/playbooks?profile_id=b5db9cbc-4ecd-464b-b416-3a6cd67af87a",
				headers: map[string]string{
					"X-Rh-Identity": base64.StdEncoding.EncodeToString([]byte(`{"identity":{"account_number":"10064","auth_type":"basic","employee_account_number":"10064","internal":{"org_id":"78606"},"org_id":"78606","type":"User","user":{"email":"collett@elfreda.name","first_name":"Maricela","is_active":true,"is_internal":false,"is_org_admin":true,"last_name":"Purdy","locale":"pa","user_id":"algae","username":"torque"}}}`)),
				},
			},
			want: `---
				# Service Enablement playbook

				# This playbook will take care of all steps required to disable
				# Insights Client
				- name: Insights Disable
				  hosts: localhost
				  become: yes
				  vars:
						insights_signature_exclude: /hosts,/vars/insights_signature
						insights_signature: !!binary |
						TFMwdExTMUNSVWRKVGlCUVIxQWdVMGxIVGtGVVZWSkZMUzB0TFMwS1ZtVnljMmx2YmpvZ1IyNTFV
						RWNnZGpFS0NtbFJTVlpCZDFWQldVODNjekU0ZG5jMU9FUXJhalZ3VGtGUmFGcGlhRUZCYkdkdGFI
						WXpZVTR5WjFJM2EwRXdiRVZMSzNNeVVVeGtiSE00YkhSaVZXZ0tORkZoWlZaSVNWa3pPRTVsTXpG
						aFJUTkRURFJZV0ZneVVuSlhUbk5QVG5GbFZFUmpPV2xOVERjM05uZzFjbTE2VWk5bFVrbG5NbXg1
						UVRoQkwwOWpOd3A0ZGtkcE1uaHBSRkZVWWtsVE9XaFRTM04yZEZKVllXbHdWWEIwUkV0TVlVcHZN
						VTl2Ulhkd2JqQXhUVGMyZDJOQlZqSmxUR1Y0YkhweU5TOXpOazlMQ25oRU4waFFiMjlpU0RGblVG
						QjNVbmszZDFadVdIUXhSbE5DYVVKUlYzcE9XRGRzU0hOR1RUaHVjbE01UlhaMWJ6VjBTMmh6Y1Zo
						U2VqQnNXR0prWVZnS1NVeERiVWhMVkdjd2JESm9iRTA1V25sS1JqTllNRUpLWVV0dFRWRjFibVpL
						Wkd0NlMxSlpOR2QyUTBaTFZGbHBWMEZxZDNFelRreFNTMmQ0Wlhwd1VBcDVlV2xVVTBoRlRrTlZP
						RXB2V0Rsa1FuWm5UbUl5TWpreFZIUmxSbGRSVTFGcVlUazRLeXQ2VGpKV2JqVlFNbmN5TlZFd2Iw
						ZzBNRGs1Ym5kclVEazJDakptVHk5aVJpOTFTM0l4V1RBelFsSmhaRFEwWmxneGVFYzNlbXBVYUZw
						WmNYUjFUM2hyUkVKVk5USkpTRlpaYWxVMFNsVmpPWFUzYUdOTFRYRlNhSG9LVVdKc1EwSnVNMDV2
						YUVsbWEySjFNSGxqVldwQldIcHVOR3hJVTJaNFFreHFOM3BYUVU4MWEwTnNVbm8xVTJScWFIVnFk
						bUl3Tms4MlJIRkZWU3MzWkFwVWVVSTRVVXd4Y1VRclp5dFFSV3d2U0RVclZtTm1NRlJST0dnd05G
						bHBiVUpOYWpkWVFuQkxVSFpWTlc1WlJVRmtiMVIxWkUwMlpWSk1aRUl2VG5aakNtZExXV1pJTm1G
						eGNFMXRiVTFVUTFwTVRFZENLM05yY1ZwdFFVSlJTazV5VlcxM2NYRnlSakJYVVZGMk9HSkxZMFpp
						Tm1kb0swbzBlalJLVW5nM1dqWUtkVU5GV2tsRlFVWnRSbkkwTDNjcmJ6QndaM1ZJYlZCRVZrNUZZ
						WGhTVWpWMlNFSm9Xa2xRV25wNlUwNXRhMDAwWTNWblZHbDZUM0JMVUhoTVRYWlJTZ3A0U0ZCUFZq
						SjFXRUZXTkQwS1BWQkphRThLTFMwdExTMUZUa1FnVUVkUUlGTkpSMDVCVkZWU1JTMHRMUzB0Q2c9
						PQ==   
				  tasks:
					- name: Disable the insights-client
					  command: insights-client --disable-schedule

				# This playbook will perform steps required to remove Insights Compliance service from your system.
				- name: Compliance OpenSCAP Remove
				  hosts: localhost
				  become: yes
				  vars:
					insights_signature_exclude: /hosts,/vars/insights_signature
					insights_signature: !!binary |
					TFMwdExTMUNSVWRKVGlCUVIxQWdVMGxIVGtGVVZWSkZMUzB0TFMwS1ZtVnljMmx2YmpvZ1IyNTFV
					RWNnZGpFS0NtbFJTVlpCZDFWQldVbENSa2c0ZG5jMU9FUXJhalZ3VGtGUmFuSXlRUzhyU3pOa1kw
					dGtUVkZpYTNVellWRmFVV1UzY0VSWVdVdEdabXhsU1M5dlFra0tWR1lyZURkYVkyNUpaMk5WWjBo
					MlpWWlRiVFpvVWtzM2QycDZhWGxGV0RGTlZHUTFPVGhTYUhaU2VGRjZVMUZvTVd0V1EzQkROWG94
					SzBob2FXTkxjd3B4TjFGV2RXZ3lNWFpFVm10TGRIazNSemhKZVhKT2J6UnlXbTl4ZG5kSWJUaDNh
					VFJVV0RKTVNFcDRhM0JaSzBGSldFMVVPR1JXZEU1VGQySXZRekYwQ2tFek1sUjNMMmx2Y0hONlpq
					SkdOWGxsZWpWcU9YbEZaa0ZGSzNocE9YWnZjMlpFZFV4blRXWjVUVVV6UjJZeVdYRm1OVFpETVZa
					NlJUWmthVU4wWm00S2JsQjZTR2xOUW1kTVVVZHpORmR4V0VwUGVrc3dZbU4zZW5VNVdUTXdZVXhQ
					VUZneWNHWm9hMEpQWW5wbGRDOXBOR2RhTTBoNmRWTkJRamxaYmpOYUx3cEJaa2xSZW5OQ2JrMUNV
					a05LYTNOeVJrZzVPV0V5YzAwdlVIUXJaR2hMZW01Q1lXOUpRbGhwY25GMFF6aEVNbEpHUVROdFpY
					bzBPVlpyVjB4V1drUkhDalZQTTBrMFNXMDNSVGMyYkdGaUsxWk5jbmRQTVV4S2NrVnRNazFGVWtK
					aFRrWm1ObGhSV0ZkVlpqQTVURXBRYmpSd2NXd3diRTgwYjNORE1sWkJkVGtLVHpOcGFuaDNWVEJ2
					YTA1dGJsUm5hVFpqTTFBNGRubEtNbTVvTkRscFQzbHJhRmxRY0ROa1ExVmlZMGRCTDBsM1UzSnFk
					MmRZSzFSck9ISmpMMDVuV1FwTmVVUlNRalZtWlc1UGFrdElPRVp1V2tvM1VXY3dWekUzUzI5WVox
					VjBaa2w2WmxORFZqTnpiMVJ3TVVKcVZ6STVURWhyWlV4dVZIbDFZbTVCVlRCc0NtVnlPRmhoV2s5
					RmVTOXpZbmhPVEZKSWVVb3ZUMkpWUVc5NlQzcGFORUZzVmswMk0zUTNVbUZzTTJvMlkyMW1VbXhD
					VlhSQlJsaDNjVFpDYWpSQ2JYUUtSV1ZpY1N0UFl6aDZZM3BIZEVaNU55dHJSa1F6V0VKbGJubFJW
					emxrTWk4eU9XWkdNVzF2VmxsVVRXazRiQzlLYVZKU1ZtSlJUelI0TWtwMmVWUTNTQXBMVTBwUlRE
					SnZUMnc1T0QwS1BXcEpaa3dLTFMwdExTMUZUa1FnVUVkUUlGTkpSMDVCVkZWU1JTMHRMUzB0Q2c9
					PQ==
				  tasks:
					- name: Compliance OpenSCAP disabled
					debug:
						msg: "Compliance OpenSCAP is not enabled. Nothing to do"`,
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

			req := httptest.NewRequest(test.input.method, test.input.url, nil)
			for k, v := range test.input.headers {
				req.Header.Add(k, v)
			}
			rr := httptest.NewRecorder()

			router := chi.NewMux()
			router.Use(identity.EnforceIdentity)
			router.Get("/playbooks", getPlaybook)
			router.ServeHTTP(rr, req)

			got := util.NormalizeWhitespace(rr.Body.String())
			test.want = util.NormalizeWhitespace(test.want)
			if !cmp.Equal(got, test.want) {
				t.Errorf("%v", cmp.Diff(got, test.want))
			}
		})
	}
}
