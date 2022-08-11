package cloudconnector

import (
	"config-manager/internal/config"
	"config-manager/internal/http/staticmux"
	"config-manager/internal/url"
	"context"
	"log"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestGetConnections(t *testing.T) {
	tests := []struct {
		description string
		input       struct {
			accountID string
			response  []byte
		}
		want      []string
		wantError error
	}{
		{
			description: "two connections",
			input: struct {
				accountID string
				response  []byte
			}{
				accountID: "000001",
				response:  []byte(`{"connections":["3d711f8b-77d0-4ed5-a5b5-1d282bf930c7", "74368f32-4e6d-4ea2-9b8f-22dac89f9ae4"]}`),
			},
			want: []string{"3d711f8b-77d0-4ed5-a5b5-1d282bf930c7", "74368f32-4e6d-4ea2-9b8f-22dac89f9ae4"},
		},
		{
			description: "zero connections",
			input: struct {
				accountID string
				response  []byte
			}{
				accountID: "000001",
				response:  []byte(`{"connections":[]}`),
			},
			want: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			config.DefaultConfig.CloudConnectorHost.Value = url.MustParse("http://localhost:8080")
			config.DefaultConfig.CloudConnectorClientID = "test"
			config.DefaultConfig.CloudConnectorPSK = "test"

			mux := staticmux.StaticMux{}
			mux.AddResponse("/connection/"+test.input.accountID, 200, test.input.response, map[string][]string{"Content-Type": {"application/json"}})
			server := http.Server{Addr: config.DefaultConfig.CloudConnectorHost.Value.Host, Handler: &mux}
			defer server.Close()
			go func() {
				if err := server.ListenAndServe(); err != nil {
					log.Print(err)
				}
			}()

			connector, err := NewCloudConnectorClient()
			if err != nil {
				t.Fatal(err)
			}

			got, err := connector.GetConnections(context.Background(), test.input.accountID)

			if test.wantError != nil {
				if !cmp.Equal(err, test.wantError, cmpopts.EquateErrors()) {
					t.Errorf("%#v != %#v", err, test.wantError)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(got, test.want) {
					t.Errorf("%v", cmp.Diff(got, test.want))
				}
			}
		})
	}
}

func TestSendMessage(t *testing.T) {
	tests := []struct {
		description string
		input       struct {
			accountID string
			directive string
			payload   []byte
			metadata  map[string]string
			recipient string
			response  []byte
		}
		want      string
		wantError error
	}{
		{
			description: "successful",
			input: struct {
				accountID string
				directive string
				payload   []byte
				metadata  map[string]string
				recipient string
				response  []byte
			}{
				accountID: "000001",
				directive: "test",
				payload:   []byte(`"test"`),
				metadata:  nil,
				recipient: "test",
				response:  []byte(`{"id":"0afbfb55-a2af-43f2-84da-a0896f03f067"}`),
			},
			want: `0afbfb55-a2af-43f2-84da-a0896f03f067`,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			config.DefaultConfig.CloudConnectorHost.Value = url.MustParse("http://localhost:8080")
			config.DefaultConfig.CloudConnectorClientID = "test"
			config.DefaultConfig.CloudConnectorPSK = "test"

			mux := staticmux.StaticMux{}
			mux.AddResponse("/message", 201, test.input.response, map[string][]string{"Content-Type": {"application/json"}})
			server := http.Server{Addr: config.DefaultConfig.CloudConnectorHost.Value.Host, Handler: &mux}
			defer server.Close()
			go func() {
				if err := server.ListenAndServe(); err != nil {
					log.Print(err)
				}
			}()

			connector, err := NewCloudConnectorClient()
			if err != nil {
				t.Fatal(err)
			}

			got, err := connector.SendMessage(context.Background(), test.input.accountID, test.input.directive, test.input.payload, test.input.metadata, test.input.recipient)

			if test.wantError != nil {
				if !cmp.Equal(err, test.wantError, cmpopts.EquateErrors()) {
					t.Errorf("%#v != %#v", err, test.wantError)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(got, test.want) {
					t.Errorf("%v", cmp.Diff(got, test.want))
				}
			}
		})
	}
}

func TestGetConnectionStatus(t *testing.T) {
	type response struct {
		status      string
		dispatchers map[string]interface{}
	}

	tests := []struct {
		description string
		input       struct {
			accountID string
			recipient string
			response  []byte
		}
		want      response
		wantError error
	}{
		{
			description: "connected with rhc-worker-playbook dispatcher",
			input: struct {
				accountID string
				recipient string
				response  []byte
			}{
				accountID: "000001",
				recipient: "test",
				response:  []byte(`{"status":"connected","dispatchers":{"rhc-worker-playbook":{}}}`),
			},
			want: response{
				status: "connected",
				dispatchers: map[string]interface{}{
					"rhc-worker-playbook": map[string]interface{}{},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			config.DefaultConfig.CloudConnectorHost.Value = url.MustParse("http://localhost:8080")
			config.DefaultConfig.CloudConnectorClientID = "test"
			config.DefaultConfig.CloudConnectorPSK = "test"

			mux := staticmux.StaticMux{}
			mux.AddResponse("/connection/status", 200, test.input.response, map[string][]string{"Content-Type": {"application/json"}})
			server := http.Server{Addr: config.DefaultConfig.CloudConnectorHost.Value.Host, Handler: &mux}
			defer server.Close()
			go func() {
				if err := server.ListenAndServe(); err != nil {
					log.Print(err)
				}
			}()

			connector, err := NewCloudConnectorClient()
			if err != nil {
				t.Fatal(err)
			}

			var got response
			got.status, got.dispatchers, err = connector.GetConnectionStatus(context.Background(), test.input.accountID, test.input.recipient)

			if test.wantError != nil {
				if !cmp.Equal(err, test.wantError, cmpopts.EquateErrors()) {
					t.Errorf("%#v != %#v", err, test.wantError)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(got, test.want, cmp.AllowUnexported(response{})) {
					t.Errorf("%v", cmp.Diff(got, test.want))
				}
			}
		})
	}
}
