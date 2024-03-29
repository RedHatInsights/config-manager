package inventory

import (
	"config-manager/internal"
	"config-manager/internal/config"
	"config-manager/internal/http/staticmux"
	"config-manager/internal/url"
	"context"

	"net/http/httptest"
	"testing"

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

			mux := staticmux.StaticMux{}
			responseBody := test.input.response
			headers := map[string][]string{"Content-Type": {"application/json"}}
			mux.AddResponse("/api/inventory/v1/hosts", 200, responseBody, headers)

			server := httptest.NewServer(&mux)
			defer server.Close()

			config.DefaultConfig.InventoryHost.Value = url.MustParse(server.URL)

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
