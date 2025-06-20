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
	request                *kesselv2.CheckRequest
	response               *kesselv2.CheckResponse
	responseError          error
	forUpdateRequest       *kesselv2.CheckForUpdateRequest
	forUpdateResponse      *kesselv2.CheckForUpdateResponse
	forUpdateResponseError error
}

func (m *mockKesselInventoryServiceClient) Check(ctx context.Context, in *kesselv2.CheckRequest, opts ...grpc.CallOption) (*kesselv2.CheckResponse, error) {
	m.request = in
	return m.response, m.responseError
}

func (m *mockKesselInventoryServiceClient) CheckForUpdate(ctx context.Context, in *kesselv2.CheckForUpdateRequest, opts ...grpc.CallOption) (*kesselv2.CheckForUpdateResponse, error) {
	m.forUpdateRequest = in
	return m.forUpdateResponse, m.forUpdateResponseError
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

type mockRbacClient struct {
	orgID string
	id    string
	err   error
}

func (m *mockRbacClient) GetDefaultWorkspaceID(ctx context.Context, orgID string) (string, error) {
	m.orgID = orgID
	return m.id, m.err
}

func TestKesselMiddleware(t *testing.T) {

	tests := []struct {
		description     string
		config          config.Config
		allowed         kesselv2.Allowed
		err             error
		rbacClient      RbacClient
		identity        identity.Identity
		permission      string
		want            int
		wantWorkspaceID string
		wantPrincipalID string
	}{
		{
			description: "return 403 if kessel is enabled and kessel returns false for a user",
			config:      configWithKesselEnabled(true),
			allowed:     kesselv2.Allowed_ALLOWED_FALSE,
			rbacClient: &mockRbacClient{
				id:  "019496b6-ff35-71a0-8bb4-ff7f0579a4c2",
				err: nil,
			},
			identity: identity.Identity{
				OrgID: "540155",
				User: &identity.User{
					Username: "user",
					UserID:   "1212",
				},
				Type: "User",
			},
			permission:      "config_manager_profile_view",
			want:            403,
			wantWorkspaceID: "019496b6-ff35-71a0-8bb4-ff7f0579a4c2",
			wantPrincipalID: "redhat/1212",
		},
		{
			description: "return 403 if kessel is enabled and kessel returns false for a service account",
			config:      configWithKesselEnabled(true),
			allowed:     kesselv2.Allowed_ALLOWED_FALSE,
			rbacClient: &mockRbacClient{
				id:  "019496b6-ff35-71a0-8bb4-ff7f0579a4c2",
				err: nil,
			},
			identity: identity.Identity{
				OrgID: "540155",
				ServiceAccount: &identity.ServiceAccount{
					Username: "service-account-b69eaf9e-e6a6-4f9e-805e-02987daddfbd",
					UserId:   "60ce65dc-4b5a-4812-8b65-b48178d92b12",
				},
				Type: "ServiceAccount",
			},
			permission:      "config_manager_profile_view",
			want:            403,
			wantWorkspaceID: "019496b6-ff35-71a0-8bb4-ff7f0579a4c2",
			wantPrincipalID: "redhat/60ce65dc-4b5a-4812-8b65-b48178d92b12",
		},
		{
			description: "return 200 if kessel is enabled and kessel returns true a user",
			config:      configWithKesselEnabled(true),
			allowed:     kesselv2.Allowed_ALLOWED_TRUE,
			rbacClient: &mockRbacClient{
				id:  "019496b6-ff35-71a0-8bb4-ff7f0579a4c2",
				err: nil,
			},
			identity: identity.Identity{
				OrgID: "540155",
				User: &identity.User{
					Username: "user",
					UserID:   "1212",
				},
				Type: "User",
			},
			permission:      "config_manager_profile_view",
			want:            200,
			wantWorkspaceID: "019496b6-ff35-71a0-8bb4-ff7f0579a4c2",
			wantPrincipalID: "redhat/1212",
		},
		{
			description: "return 200 if kessel is enabled and kessel returns true a service account",
			config:      configWithKesselEnabled(true),
			allowed:     kesselv2.Allowed_ALLOWED_TRUE,
			rbacClient: &mockRbacClient{
				id:  "019496b6-ff35-71a0-8bb4-ff7f0579a4c2",
				err: nil,
			},
			identity: identity.Identity{
				OrgID: "540155",
				ServiceAccount: &identity.ServiceAccount{
					Username: "service-account-b69eaf9e-e6a6-4f9e-805e-02987daddfbd",
					UserId:   "60ce65dc-4b5a-4812-8b65-b48178d92b12",
				},
				Type: "ServiceAccount",
			},
			permission:      "config_manager_profile_view",
			want:            200,
			wantWorkspaceID: "019496b6-ff35-71a0-8bb4-ff7f0579a4c2",
			wantPrincipalID: "redhat/60ce65dc-4b5a-4812-8b65-b48178d92b12",
		},
		{
			description: "return 200 if kessel is disabled",
			config:      configWithKesselEnabled(false),
			allowed:     kesselv2.Allowed_ALLOWED_FALSE,
			rbacClient: &mockRbacClient{
				id:  "019496b6-ff35-71a0-8bb4-ff7f0579a4c2",
				err: nil,
			},
			identity: identity.Identity{
				OrgID: "540155",
				User: &identity.User{
					Username: "user",
					UserID:   "1212",
				},
				Type: "User",
			},
			permission:      "config_manager_profile_view",
			want:            200,
			wantWorkspaceID: "019496b6-ff35-71a0-8bb4-ff7f0579a4c2",
			wantPrincipalID: "redhat/1212",
		},
		{
			description: "return 500 on kessel error",
			config:      configWithKesselEnabled(true),
			allowed:     kesselv2.Allowed_ALLOWED_FALSE,
			rbacClient: &mockRbacClient{
				id:  "019496b6-ff35-71a0-8bb4-ff7f0579a4c2",
				err: nil,
			},
			err: context.Canceled,
			identity: identity.Identity{
				OrgID: "540155",
				User: &identity.User{
					Username: "user",
					UserID:   "1212",
				},
				Type: "User",
			},
			permission:      "config_manager_profile_view",
			want:            500,
			wantWorkspaceID: "019496b6-ff35-71a0-8bb4-ff7f0579a4c2",
			wantPrincipalID: "redhat/1212",
		},
		{
			description: "returns 500 on rbac error",
			config:      configWithKesselEnabled(true),
			allowed:     kesselv2.Allowed_ALLOWED_TRUE,
			rbacClient: &mockRbacClient{
				id:  "",
				err: context.Canceled,
			},
			identity: identity.Identity{
				OrgID: "540155",
				User: &identity.User{
					Username: "user",
					UserID:   "1212",
				},
				Type: "User",
			},
			permission: "config_manager_profile_view",
			want:       500,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &mockKesselInventoryServiceClient{
				response: &kesselv2.CheckResponse{
					Allowed: test.allowed,
				},
				responseError: test.err,
			}

			middlewareBuilder := &kesselMiddlewareBuilderImpl{
				config: test.config,
				client: &v1beta1.InventoryClient{
					KesselInventoryService: client,
				},
				rbacClient: test.rbacClient,
			}

			sampleHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("OK"))
			})

			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/profiles", nil)
			req = req.WithContext(identity.WithIdentity(req.Context(), identity.XRHID{Identity: test.identity}))

			middlewareBuilder.EnforceDefaultWorkspacePermission(test.permission)(sampleHandler).ServeHTTP(rr, req)

			if test.config.KesselEnabled && test.wantWorkspaceID != "" {
				assertEquals(t, "response status code", test.want, rr.Code)
				assertEquals(t, "workspace id", test.wantWorkspaceID, client.request.Object.ResourceId)
				assertEquals(t, "workspace resource type", "workspace", client.request.Object.ResourceType)
				assertEquals(t, "workspace reporter type", "rbac", client.request.Object.Reporter.Type)
				assertEquals(t, "relation", test.permission, client.request.Relation)
				assertEquals(t, "principal id", test.wantPrincipalID, client.request.Subject.Resource.ResourceId)
				assertEquals(t, "principal resource type", "principal", client.request.Subject.Resource.ResourceType)
				assertEquals(t, "principal reporter type", "rbac", client.request.Subject.Resource.Reporter.Type)
			}
		})

		t.Run(test.description+" (for update)", func(t *testing.T) {
			client := &mockKesselInventoryServiceClient{
				forUpdateResponse: &kesselv2.CheckForUpdateResponse{
					Allowed: test.allowed,
				},
				forUpdateResponseError: test.err,
			}

			middlewareBuilder := &kesselMiddlewareBuilderImpl{
				config: test.config,
				client: &v1beta1.InventoryClient{
					KesselInventoryService: client,
				},
				rbacClient: test.rbacClient,
			}

			sampleHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("OK"))
			})

			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/profiles", nil)
			req = req.WithContext(identity.WithIdentity(req.Context(), identity.XRHID{Identity: test.identity}))

			middlewareBuilder.EnforceDefaultWorkspacePermissionForUpdate(test.permission)(sampleHandler).ServeHTTP(rr, req)

			if test.config.KesselEnabled && test.wantWorkspaceID != "" {
				assertEquals(t, "response status code", test.want, rr.Code)
				assertEquals(t, "workspace id", test.wantWorkspaceID, client.forUpdateRequest.Object.ResourceId)
				assertEquals(t, "workspace resource type", "workspace", client.forUpdateRequest.Object.ResourceType)
				assertEquals(t, "workspace reporter type", "rbac", client.forUpdateRequest.Object.Reporter.Type)
				assertEquals(t, "relation", test.permission, client.forUpdateRequest.Relation)
				assertEquals(t, "principal id", test.wantPrincipalID, client.forUpdateRequest.Subject.Resource.ResourceId)
				assertEquals(t, "principal resource type", "principal", client.forUpdateRequest.Subject.Resource.ResourceType)
				assertEquals(t, "principal reporter type", "rbac", client.forUpdateRequest.Subject.Resource.Reporter.Type)
			}
		})
	}
}

func assertEquals[T comparable](t *testing.T, field string, want, got T) {
	if got != want {
		t.Errorf("expected %s %v, got %v", field, want, got)
	}
}
