package authorization

import (
	"config-manager/internal/config"
	"config-manager/internal/instrumentation"
	"errors"
	"net/http"

	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
	"github.com/project-kessel/inventory-client-go/common"
	v1beta2 "github.com/project-kessel/inventory-client-go/v1beta2"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
)

func NewKesselClient(config config.Config) KesselMiddlewareBuilder {

	client, _ := v1beta2.New(common.NewConfig(
		common.WithgRPCUrl("localhost:9091"), // TODO: more configuration options
		common.WithTLSInsecure(true),
	))

	return &kesselMiddlewareBuilderImpl{
		client: client,
		config: config,
	}
}

type KesselMiddlewareBuilder interface {
	EnforceOrgPermission(permission string) func(http.Handler) http.Handler
}

type kesselMiddlewareBuilderImpl struct {
	client *v1beta2.InventoryClient
	config config.Config
}

var _ KesselMiddlewareBuilder = &kesselMiddlewareBuilderImpl{}

func (a *kesselMiddlewareBuilderImpl) EnforceOrgPermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !a.config.KesselEnabled {
				next.ServeHTTP(w, r)
				return
			}

			id := identity.GetIdentity(r.Context())

			principalId, err := extractPrincipalId(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
			}

			kesselRequest := &kesselv2.CheckRequest{
				Object: &kesselv2.ResourceReference{
					ResourceType: "tenant",
					ResourceId:   id.Identity.OrgID,
					Reporter: &kesselv2.ReporterReference{
						Type: "rbac",
					},
				},
				Relation: permission,
				Subject: &kesselv2.SubjectReference{
					Resource: &kesselv2.ResourceReference{
						ResourceType: "principal",
						ResourceId:   principalId,
						Reporter: &kesselv2.ReporterReference{
							Type: "rbac",
						},
					},
				},
			}

			//opts, _ := a.client.GetTokenCallOption()

			res, err := a.client.KesselInventoryService.Check(r.Context(), kesselRequest)
			if err != nil {
				instrumentation.AuthorizationCheckError(err)
				http.Error(w, "Error performing authorization check", http.StatusInternalServerError)
				return
			}

			if res.GetAllowed() != kesselv2.Allowed_ALLOWED_TRUE {
				instrumentation.AuthorizationCheckFailed(kesselRequest.Subject.Resource.ResourceId, kesselRequest.Object.ResourceId, kesselRequest.Relation)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			instrumentation.AuthorizationCheckPassed(kesselRequest.Subject.Resource.ResourceId, kesselRequest.Object.ResourceId, kesselRequest.Relation)
			next.ServeHTTP(w, r)
		})
	}
}

func extractPrincipalId(identity identity.XRHID) (string, error) {
	switch identity.Identity.Type {
	case "User":
		return identity.Identity.User.Username, nil
	case "ServiceAccount":
		return identity.Identity.ServiceAccount.Username, nil
	default:
		return "", errors.New("unsupported identity type")
	}
}
