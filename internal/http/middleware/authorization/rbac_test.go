package authorization

import (
	"config-manager/internal/config"
	"config-manager/internal/http/staticmux"
	"config-manager/internal/url"
	"context"
	"errors"

	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestGetInventoryClients(t *testing.T) {
	tests := []struct {
		description  string
		response     string
		responseCode int
		want         string
		wantErr      error
	}{
		{
			description:  "OK",
			response:     `{"meta":{"count":1,"limit":10,"offset":0},"links":{"first":"/api/rbac/v2/workspaces/?limit=10&offset=0&type=default","next":null,"previous":null,"last":"/api/rbac/v2/workspaces/?limit=10&offset=0&type=default"},"data":[{"name":"Default Workspace","id":"01973edb-2cbd-7be1-9901-fa3c23ead696","parent_id":"01973edb-2cbd-7be1-9901-fa22d3c104a3","description":null,"created":"2025-06-05T06:50:40.701926Z","modified":"2025-06-05T06:50:40.778886Z","type":"default"}]}`,
			responseCode: 200,
			want:         "01973edb-2cbd-7be1-9901-fa3c23ead696",
		},
		{
			description:  "unexpected status code",
			response:     `{"meta":{"count":1,"limit":10,"offset":0},"links":{"first":"/api/rbac/v2/workspaces/?limit=10&offset=0&type=default","next":null,"previous":null,"last":"/api/rbac/v2/workspaces/?limit=10&offset=0&type=default"},"data":[{"name":"Default Workspace","id":"01973edb-2cbd-7be1-9901-fa3c23ead696","parent_id":"01973edb-2cbd-7be1-9901-fa22d3c104a3","description":null,"created":"2025-06-05T06:50:40.701926Z","modified":"2025-06-05T06:50:40.778886Z","type":"default"}]}`,
			responseCode: 403,
			wantErr:      errors.New("unexpected status code: 403"),
		},
		{
			description:  "no default workspace",
			response:     `{"meta":{"count":0,"limit":10,"offset":0},"links":{"first":"/api/rbac/v2/workspaces/?limit=10&offset=0&type=default","next":null,"previous":null,"last":"/api/rbac/v2/workspaces/?limit=10&offset=0&type=default"},"data":[]}`,
			responseCode: 200,
			wantErr:      errors.New("unexpected number of default workspaces: 0"),
		},
		{
			description:  "too many workspaces",
			response:     `{"meta":{"count":2,"limit":10,"offset":0},"links":{"first":"/api/rbac/v2/workspaces/?limit=10&offset=0&type=default","next":null,"previous":null,"last":"/api/rbac/v2/workspaces/?limit=10&offset=0&type=default"},"data":[{"name":"Default Workspace","id":"01973edb-2cbd-7be1-9901-fa3c23ead696","parent_id":"01973edb-2cbd-7be1-9901-fa22d3c104a3","description":null,"created":"2025-06-05T06:50:40.701926Z","modified":"2025-06-05T06:50:40.778886Z","type":"default"}, {"name":"Default Workspace","id":"8039249f-63f9-4565-871d-41c05c678628","parent_id":"01973edb-2cbd-7be1-9901-fa22d3c104a3","description":null,"created":"2025-06-05T06:50:40.701926Z","modified":"2025-06-05T06:50:40.778886Z","type":"default"}]}`,
			responseCode: 200,
			wantErr:      errors.New("unexpected number of default workspaces: 2"),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

			mux := staticmux.StaticMux{}
			headers := map[string][]string{"content-type": {"application/json"}}
			mux.AddResponse("/api/rbac/v2/workspaces/", test.responseCode, []byte(test.response), headers)

			server := httptest.NewServer(&mux)
			defer server.Close()

			config.DefaultConfig.InventoryHost.Value = url.MustParse(server.URL)

			workspaceID, err := newRbacClient(server.URL, nil).GetDefaultWorkspaceID(context.TODO(), "540155")

			if test.wantErr != nil {
				if err.Error() != test.wantErr.Error() {
					t.Errorf("unexpected error: %s", cmp.Diff(err, test.wantErr, cmpopts.EquateErrors()))
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if workspaceID != test.want {
					t.Errorf("%v", cmp.Diff(workspaceID, test.want))
				}
			}
		})
	}
}
