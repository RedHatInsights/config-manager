package cloudconnector

import (
	"config-manager/internal/config"
	"config-manager/internal/http/staticmux"
	"config-manager/internal/url"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestSendMessage(t *testing.T) {
	tests := []struct {
		description string
		input       struct {
			orgID     string
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
				orgID     string
				directive string
				payload   []byte
				metadata  map[string]string
				recipient string
				response  []byte
			}{
				orgID:     "000001",
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
			rand.Seed(time.Now().UnixNano())
			port := uint32(rand.Int31n(65535-55535) + 55535)

			config.DefaultConfig.CloudConnectorHost.Value = url.MustParse(fmt.Sprintf("http://localhost:%v", port))
			config.DefaultConfig.CloudConnectorClientID = "test"
			config.DefaultConfig.CloudConnectorPSK = "test"

			mux := staticmux.StaticMux{}
			mux.AddResponse("/v2/connections/"+test.input.recipient+"/message", 201, test.input.response, map[string][]string{"Content-Type": {"application/json"}})
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

			got, err := connector.SendMessage(context.Background(), test.input.orgID, test.input.directive, test.input.payload, test.input.metadata, test.input.recipient)

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
			orgID     string
			recipient string
			response  []byte
		}
		want      response
		wantError error
	}{
		{
			description: "connected with rhc-worker-playbook dispatcher",
			input: struct {
				orgID     string
				recipient string
				response  []byte
			}{
				orgID:     "000001",
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
			rand.Seed(time.Now().UnixNano())
			port := uint32(rand.Int31n(65535-55535) + 55535)

			config.DefaultConfig.CloudConnectorHost.Value = url.MustParse(fmt.Sprintf("http://localhost:%v", port))
			config.DefaultConfig.CloudConnectorClientID = "test"
			config.DefaultConfig.CloudConnectorPSK = "test"

			mux := staticmux.StaticMux{}
			mux.AddResponse("/v2/connections/"+test.input.recipient+"/status", 200, test.input.response, map[string][]string{"Content-Type": {"application/json"}})
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
			got.status, got.dispatchers, err = connector.GetConnectionStatus(context.Background(), test.input.orgID, test.input.recipient)

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
