package dispatcher

import (
	"config-manager/internal/config"
	"config-manager/internal/http/staticmux"
	"config-manager/internal/url"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestDispatch(t *testing.T) {
	tests := []struct {
		description string
		input       struct {
			runs     []RunInputV2
			response []byte
		}
		want []RunCreated
	}{
		{
			description: "two responses",
			input: struct {
				runs     []RunInputV2
				response []byte
			}{
				runs: []RunInputV2{
					{
						Recipient: uuid.MustParse("276c4685-fdfb-4172-930f-4148b8340c2e"),
						OrgId:     "0000001",
						Principal: "test_user",
						Url:       "https://cloud.redhat.com/api/config-manager/v1/states/e417581a-d649-4cdc-9506-6eb7fdbfd66d/playbook",
						Name:      "Apply fix",
						Labels: &Labels{
							"test": "e417581a-d649-4cdc-9506-6eb7fdbfd66d",
						},
					},
					{
						Recipient: uuid.MustParse("9a76b28b-0e09-41c8-bf01-79d1bef72646"),
						OrgId:     "0000001",
						Principal: "test_user",
						Url:       "https://cloud.redhat.com/api/config-manager/v1/states/e417581a-d649-4cdc-9506-6eb7fdbfd66d/playbook",
						Name:      "Apply fix",
						Labels: &Labels{
							"test": "e417581a-d649-4cdc-9506-6eb7fdbfd66d",
						},
					},
				},
				response: []byte(`[{"code":200,"id":"3d711f8b-77d0-4ed5-a5b5-1d282bf930c7"},{"code":200,"id":"74368f32-4e6d-4ea2-9b8f-22dac89f9ae4"}]`),
			},
			want: []RunCreated{
				{
					Code: 200,
					Id:   func(s uuid.UUID) *uuid.UUID { return &s }(uuid.MustParse("3d711f8b-77d0-4ed5-a5b5-1d282bf930c7")),
				},
				{
					Code: 200,
					Id:   func(s uuid.UUID) *uuid.UUID { return &s }(uuid.MustParse("74368f32-4e6d-4ea2-9b8f-22dac89f9ae4")),
				},
			},
		},
		{
			description: "missing responses",
			input: struct {
				runs     []RunInputV2
				response []byte
			}{
				runs: []RunInputV2{
					{
						Recipient: uuid.MustParse("276c4685-fdfb-4172-930f-4148b8340c2e"),
						OrgId:     "0000001",
						Principal: "test_user",
						Url:       "https://cloud.redhat.com/api/config-manager/v1/states/e417581a-d649-4cdc-9506-6eb7fdbfd66d/playbook",
						Name:      "Apply Fix",
						Labels: &Labels{
							"test": "e417581a-d649-4cdc-9506-6eb7fdbfd66d",
						},
					},
					{
						Recipient: uuid.MustParse("9a76b28b-0e09-41c8-bf01-79d1bef72646"),
						OrgId:     "0000001",
						Principal: "test_user",
						Url:       "https://cloud.redhat.com/api/config-manager/v1/states/e417581a-d649-4cdc-9506-6eb7fdbfd66d/playbook",
						Name:      "Apply Fix",
						Labels: &Labels{
							"test": "e417581a-d649-4cdc-9506-6eb7fdbfd66d",
						},
					},
				},
				response: []byte(`[{"code":404},{"code":404}]`),
			},
			want: []RunCreated{
				{
					Code: 404,
				},
				{
					Code: 404,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			mux := staticmux.StaticMux{}
			responseBody := test.input.response
			headers := map[string][]string{"Content-Type": {"application/json"}}
			mux.AddResponse("/internal/v2/dispatch", 207, responseBody, headers)

			server := httptest.NewServer(&mux)
			defer server.Close()

			config.DefaultConfig.DispatcherHost.Value = url.MustParse(server.URL)

			got, err := NewDispatcherClient().Dispatch(context.Background(), test.input.runs)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(got, test.want) {
				t.Errorf("%v", cmp.Diff(got, test.want))
			}
		})
	}
}
