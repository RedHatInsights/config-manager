package dispatcher

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
)

func TestDispatch(t *testing.T) {
	tests := []struct {
		description string
		input       struct {
			runs     []RunInput
			response []byte
		}
		want []RunCreated
	}{
		{
			description: "two responses",
			input: struct {
				runs     []RunInput
				response []byte
			}{
				runs: []RunInput{
					{
						Recipient: "276c4685-fdfb-4172-930f-4148b8340c2e",
						Account:   "0000001",
						Url:       "https://cloud.redhat.com/api/config-manager/v1/states/e417581a-d649-4cdc-9506-6eb7fdbfd66d/playbook",
						Labels: &RunInput_Labels{
							AdditionalProperties: map[string]string{
								"test": "e417581a-d649-4cdc-9506-6eb7fdbfd66d",
							},
						},
					},
					{
						Recipient: "9a76b28b-0e09-41c8-bf01-79d1bef72646",
						Account:   "0000001",
						Url:       "https://cloud.redhat.com/api/config-manager/v1/states/e417581a-d649-4cdc-9506-6eb7fdbfd66d/playbook",
						Labels: &RunInput_Labels{
							AdditionalProperties: map[string]string{
								"test": "e417581a-d649-4cdc-9506-6eb7fdbfd66d",
							},
						},
					},
				},
				response: []byte(`[{"code":200,"id":"3d711f8b-77d0-4ed5-a5b5-1d282bf930c7"},{"code":200,"id":"74368f32-4e6d-4ea2-9b8f-22dac89f9ae4"}]`),
			},
			want: []RunCreated{
				{
					Code: 200,
					Id:   func(s string) *string { return &s }("3d711f8b-77d0-4ed5-a5b5-1d282bf930c7"),
				},
				{
					Code: 200,
					Id:   func(s string) *string { return &s }("74368f32-4e6d-4ea2-9b8f-22dac89f9ae4"),
				},
			},
		},
		{
			description: "missing responses",
			input: struct {
				runs     []RunInput
				response []byte
			}{
				runs: []RunInput{
					{
						Recipient: "276c4685-fdfb-4172-930f-4148b8340c2e",
						Account:   "0000001",
						Url:       "https://cloud.redhat.com/api/config-manager/v1/states/e417581a-d649-4cdc-9506-6eb7fdbfd66d/playbook",
						Labels: &RunInput_Labels{
							AdditionalProperties: map[string]string{
								"test": "e417581a-d649-4cdc-9506-6eb7fdbfd66d",
							},
						},
					},
					{
						Recipient: "9a76b28b-0e09-41c8-bf01-79d1bef72646",
						Account:   "0000001",
						Url:       "https://cloud.redhat.com/api/config-manager/v1/states/e417581a-d649-4cdc-9506-6eb7fdbfd66d/playbook",
						Labels: &RunInput_Labels{
							AdditionalProperties: map[string]string{
								"test": "e417581a-d649-4cdc-9506-6eb7fdbfd66d",
							},
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
			rand.Seed(time.Now().UnixNano())
			port := uint32(rand.Int31n(65535-55535) + 55535)

			config.DefaultConfig.DispatcherHost.Value = url.MustParse(fmt.Sprintf("http://localhost:%v", port))

			mux := staticmux.StaticMux{}
			mux.AddResponse("/internal/dispatch", 207, test.input.response, map[string][]string{"Content-Type": {"application/json"}})
			server := http.Server{Addr: config.DefaultConfig.DispatcherHost.Value.Host, Handler: &mux}
			defer server.Close()
			go func() {
				if err := server.ListenAndServe(); err != nil {
					log.Print(err)
				}
			}()

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
