package inventory

import (
	"config-manager/internal"
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
)

func TestGetInventoryClients(t *testing.T) {
	tests := []struct {
		description string
		input       struct {
			page     int
			response []byte
		}
		want InventoryResponse
	}{
		{
			description: "one result",
			input: struct {
				page     int
				response []byte
			}{
				page:     1,
				response: []byte(`{"total":1,"count":1,"page":1,"per_page":50,"results":[{"id":"1234","account":"000001","display_name":"test","system_profile":{"rhc_client_id":"3d711f8b-77d0-4ed5-a5b5-1d282bf930c7","rhc_config_state":"3ef6c247-d913-491b-b3eb-56315a6e0f84"}}]}`),
			},
			want: InventoryResponse{
				Total:   1,
				Count:   1,
				Page:    1,
				PerPage: 50,
				Results: []internal.Host{
					{
						ID:          "1234",
						Account:     "000001",
						DisplayName: "test",
						SystemProfile: struct {
							RHCID    string "json:\"rhc_client_id\""
							RHCState string "json:\"rhc_config_state\""
						}{
							RHCID:    "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7",
							RHCState: "3ef6c247-d913-491b-b3eb-56315a6e0f84",
						},
					},
				},
			},
		},
		{
			description: "zero results",
			input: struct {
				page     int
				response []byte
			}{
				page:     1,
				response: []byte(`{"total":0,"count":0,"page":1,"per_page":50,"results":[]}`),
			},
			want: InventoryResponse{
				Page:    1,
				PerPage: 50,
				Results: []internal.Host{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			rand.Seed(time.Now().UnixNano())
			port := uint32(rand.Int31n(65535-55535) + 55535)

			config.DefaultConfig.InventoryHost.Value = url.MustParse(fmt.Sprintf("http://localhost:%v", port))

			mux := staticmux.StaticMux{}
			mux.AddResponse("/api/inventory/v1/hosts", 200, test.input.response, map[string][]string{"Content-Type": {"application/json"}})
			server := http.Server{Addr: config.DefaultConfig.InventoryHost.Value.Host, Handler: &mux}
			defer server.Close()
			go func() {
				if err := server.ListenAndServe(); err != nil {
					log.Print(err)
				}
			}()

			got, err := NewInventoryClient().GetInventoryClients(context.Background(), test.input.page)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(got, test.want) {
				t.Errorf("%v", cmp.Diff(got, test.want))
			}
		})
	}
}
