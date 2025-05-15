package authorization

import (
	"config-manager/internal/config"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
	v1beta1 "github.com/project-kessel/inventory-client-go/v1beta2"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"google.golang.org/grpc"
)

type mockKesselInventoryServiceClient struct {
	request       *kesselv2.CheckRequest
	response      *kesselv2.CheckResponse
	responseError error
}

func (m *mockKesselInventoryServiceClient) Check(ctx context.Context, in *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
	m.request = in
	return m.response, m.responseError
}

func (m *mockKesselInventoryServiceClient) CheckForUpdate(ctx context.Context, in *kesselv2.CheckForUpdateRequest, opts ...grpc.CallOption) (*kesselv2.CheckForUpdateResponse, error) {
	panic("unimplemented")
}

func (m *mockKesselInventoryServiceClient) DeleteResource(ctx context.Context, in *kesselv2.DeleteResourceRequest, opts ...grpc.CallOption) (*kesselv2.DeleteResourceResponse, error) {
	panic("unimplemented")
}

func (m *mockKesselInventoryServiceClient) ReportResource(ctx context.Context, in *kesselv2.ReportResourceRequest, opts ...grpc.CallOption) (*kesselv2.ReportResourceResponse, error) {
	panic("unimplemented")
}

func (m *mockKesselInventoryServiceClient) StreamedListObjects(ctx context.Context, in *kesselv2.StreamedListObjectsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[kesselv2.StreamedListObjectsResponse], error) {
	panic("unimplemented")
}

var _ kesselv2.KesselInventoryServiceClient = &mockKesselInventoryServiceClient{}

func configWithKesselEnabled(value bool) config.Config {
	return config.Config{
		KesselEnabled: value,
	}
}

func mockClient(allowed kesselv2.Allowed) *mockKesselInventoryServiceClient {
	return &mockKesselInventoryServiceClient{
		response: &kesselv2.CheckResponse{
			Allowed: allowed,
		},
	}
}

func TestKesselMiddleware(t *testing.T) {

	tests := []struct {
		description string
		config      config.Config
		client      *mockKesselInventoryServiceClient
		identity    identity.Identity
		permission  string
		want        int
	}{
		{
			description: "return 403 if kessel is enabled and kessel returns false for a user",
			config:      configWithKesselEnabled(true),
			client:      mockClient(kesselv2.Allowed_ALLOWED_FALSE),
			identity: identity.Identity{
				OrgID: "540155",
				User: &identity.User{
					Username: "user",
				},
				Type: "User",
			},
			permission: "config_manager_profile_view",
			want:       403,
		},
		{
			description: "return 200 if kessel is enabled and kessel returns true a user",
			config:      configWithKesselEnabled(true),
			client:      mockClient(kesselv2.Allowed_ALLOWED_TRUE),
			identity: identity.Identity{
				OrgID: "540155",
				User: &identity.User{
					Username: "user",
				},
				Type: "User",
			},
			permission: "config_manager_profile_view",
			want:       200,
		},
		{
			description: "return 200 if kessel is disabled",
			config:      configWithKesselEnabled(false),
			client:      mockClient(kesselv2.Allowed_ALLOWED_FALSE),
			identity: identity.Identity{
				OrgID: "540155",
				User: &identity.User{
					Username: "user",
				},
				Type: "User",
			},
			permission: "config_manager_profile_view",
			want:       200,
		},
		{
			description: "return 500 on kessel error",
			config:      configWithKesselEnabled(true),
			client: &mockKesselInventoryServiceClient{
				response: &kesselv2.CheckResponse{
					Allowed: kesselv2.Allowed_ALLOWED_FALSE,
				},
				responseError: context.Canceled,
			},
			identity: identity.Identity{
				OrgID: "540155",
				User: &identity.User{
					Username: "user",
				},
				Type: "User",
			},
			permission: "config_manager_profile_view",
			want:       500,
		},

		// TODO: service account
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			middlewareBuilder := &kesselMiddlewareBuilderImpl{
				config: test.config,
				client: &v1beta1.InventoryClient{
					KesselInventoryService: test.client,
				},
			}

			sampleHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("OK"))
			})

			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/profiles", nil)
			req = req.WithContext(identity.WithIdentity(req.Context(), identity.XRHID{Identity: test.identity}))

			middlewareBuilder.EnforceOrgPermission(test.permission)(sampleHandler).ServeHTTP(rr, req)

			if test.config.KesselEnabled {
				assertEquals(t, "response status code", test.want, rr.Code)
				assertEquals(t, "tenant id", test.identity.OrgID, test.client.request.Object.ResourceId)
				assertEquals(t, "tenant resource type", "tenant", test.client.request.Object.ResourceType)
				assertEquals(t, "tenant reporter type", "rbac", test.client.request.Object.Reporter.Type)
				assertEquals(t, "relation", test.permission, test.client.request.Relation)
				assertEquals(t, "principal id", test.identity.User.Username, test.client.request.Subject.Resource.ResourceId)
				assertEquals(t, "principal resource type", "principal", test.client.request.Subject.Resource.ResourceType)
				assertEquals(t, "principal reporter type", "rbac", test.client.request.Subject.Resource.Reporter.Type)
			}
		})
	}
}

func assertEquals[T comparable](t *testing.T, field string, want, got T) {
	if got != want {
		t.Errorf("expected %s %v, got %v", field, want, got)
	}
}
