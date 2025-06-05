package authorization

import (
	"config-manager/internal/config"
	"config-manager/internal/instrumentation"
	"context"
	"errors"
	"fmt"
	"net/http"

	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
	"github.com/project-kessel/inventory-client-go/common"
	v1beta2 "github.com/project-kessel/inventory-client-go/v1beta2"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"google.golang.org/grpc"
)

func NewKesselClient(config config.Config) KesselMiddlewareBuilder {
	options := []func(*common.Config){
		common.WithgRPCUrl(config.KesselURL),
		common.WithTLSInsecure(config.KesselInsecure),
	}

	if config.KesselAuthEnabled {
		options = append(options, common.WithAuthEnabled(config.KesselAuthClientID, config.KesselAuthClientSecret, config.KesselAuthOIDCIssuer))
	}

	client, err := v1beta2.New(common.NewConfig(options...))
	if err != nil {
		panic(fmt.Errorf("failed to configure Kessel client: %w", err))
	}

	return &kesselMiddlewareBuilderImpl{
		client:     client,
		config:     config,
		rbacClient: newRbacClient(config.RbacURL),
	}
}

type KesselMiddlewareBuilder interface {
	EnforceDefaultWorkspacePermission(permission string) func(http.Handler) http.Handler
	EnforceDefaultWorkspacePermissionForUpdate(permission string) func(http.Handler) http.Handler
}

type kesselMiddlewareBuilderImpl struct {
	client     *v1beta2.InventoryClient
	config     config.Config
	rbacClient RbacClient
}

var _ KesselMiddlewareBuilder = &kesselMiddlewareBuilderImpl{}

type AllowedResponse interface {
	GetAllowed() kesselv2.Allowed
}

type kesselCheckFn func(context.Context, *kesselv2.ResourceReference, string, *kesselv2.SubjectReference, ...grpc.CallOption) (AllowedResponse, error)

func (a *kesselMiddlewareBuilderImpl) callCheck(ctx context.Context, object *kesselv2.ResourceReference, relation string, subject *kesselv2.SubjectReference, opts ...grpc.CallOption) (AllowedResponse, error) {
	request := &kesselv2.CheckRequest{
		Object:   object,
		Relation: relation,
		Subject:  subject,
	}

	return a.client.KesselInventoryService.Check(ctx, request, opts...)
}

func (a *kesselMiddlewareBuilderImpl) callCheckForUpdate(ctx context.Context, object *kesselv2.ResourceReference, relation string, subject *kesselv2.SubjectReference, opts ...grpc.CallOption) (AllowedResponse, error) {
	request := &kesselv2.CheckForUpdateRequest{
		Object:   object,
		Relation: relation,
		Subject:  subject,
	}

	return a.client.KesselInventoryService.CheckForUpdate(ctx, request, opts...)
}

func (a *kesselMiddlewareBuilderImpl) EnforceDefaultWorkspacePermission(permission string) func(http.Handler) http.Handler {
	return a.enforceOrgPermission(permission, a.callCheck)
}

func (a *kesselMiddlewareBuilderImpl) EnforceDefaultWorkspacePermissionForUpdate(permission string) func(http.Handler) http.Handler {
	return a.enforceOrgPermission(permission, a.callCheckForUpdate)
}

func (a *kesselMiddlewareBuilderImpl) enforceOrgPermission(permission string, checkFn kesselCheckFn) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !a.config.KesselEnabled {
				next.ServeHTTP(w, r)
				return
			}

			id := identity.GetIdentity(r.Context())

			defaultWorkspaceID, err := a.rbacClient.GetDefaultWorkspaceID(r.Context(), id.Identity.OrgID)
			if err != nil {
				instrumentation.WorkspaceLookupError(err, id.Identity.OrgID)
				http.Error(w, "Error performing authorization check", http.StatusInternalServerError)
				return
			}

			instrumentation.WorkspaceLookupOK(id.Identity.OrgID, defaultWorkspaceID)

			userID, err := extractUserID(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
			}

			principalID := fmt.Sprintf("redhat/%s", userID)

			object := &kesselv2.ResourceReference{
				ResourceType: "workspace",
				ResourceId:   defaultWorkspaceID,
				Reporter: &kesselv2.ReporterReference{
					Type: "rbac",
				},
			}

			subject := &kesselv2.SubjectReference{
				Resource: &kesselv2.ResourceReference{
					ResourceType: "principal",
					ResourceId:   principalID,
					Reporter: &kesselv2.ReporterReference{
						Type: "rbac",
					},
				},
			}

			var opts []grpc.CallOption
			if a.config.KesselAuthEnabled {
				opts, err = a.client.GetTokenCallOption()
				if err != nil {
					instrumentation.AuthorizationCheckError(err)
					http.Error(w, "Error performing authorization check", http.StatusInternalServerError)
					return
				}
			}

			res, err := checkFn(r.Context(), object, permission, subject, opts...)
			if err != nil {
				instrumentation.AuthorizationCheckError(err)
				http.Error(w, "Error performing authorization check", http.StatusInternalServerError)
				return
			}

			if res.GetAllowed() != kesselv2.Allowed_ALLOWED_TRUE {
				instrumentation.AuthorizationCheckFailed(subject.Resource.ResourceId, object.ResourceId, permission)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			instrumentation.AuthorizationCheckPassed(subject.Resource.ResourceId, object.ResourceId, permission)
			next.ServeHTTP(w, r)
		})
	}
}

func extractUserID(identity identity.XRHID) (string, error) {
	switch identity.Identity.Type {
	case "User":
		return identity.Identity.User.UserID, nil
	case "ServiceAccount":
		return identity.Identity.ServiceAccount.UserId, nil
	default:
		return "", errors.New("unsupported identity type")
	}
}
