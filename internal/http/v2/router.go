package v2

import (
	"config-manager/internal/config"
	"config-manager/internal/http/middleware/authorization"
	"config-manager/internal/http/render"
	"fmt"
	"net/http"
	"path"

	oapimiddleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/rs/zerolog/log"
)

//go:generate oapi-codegen -config oapi-codegen.yml ./openapi.json

func NewMux() (*chi.Mux, error) {
	spec, err := GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("cannot get OpenAPI spec: %w", err)
	}

	router := chi.NewMux()

	router.Use(httplog.RequestLogger(httplog.NewLogger("v2", httplog.Options{
		LogLevel: config.DefaultConfig.LogLevel.Value,
		JSON:     config.DefaultConfig.LogFormat.Value == "json",
	})))
	router.Use(identity.EnforceIdentity)
	router.Use(middleware.RequestID)
	router.Get(path.Join("/", "openapi.json"), func(w http.ResponseWriter, r *http.Request) {
		render.RenderJSON(w, r, http.StatusOK, spec, log.Logger)
	})

	kessel := authorization.NewKesselClient(config.DefaultConfig)

	router.Route("/", func(r chi.Router) {
		r.Use(oapimiddleware.OapiRequestValidator(spec))

		r.Group(func(r chi.Router) {
			r.Use(kessel.EnforceOrgPermission("config_manager_profile_view"))
			r.Get("/profiles", getProfiles)
			r.Get("/profiles/{id}", getProfile)
			r.Get("/playbooks", getPlaybook)
		})

		r.Group(func(r chi.Router) {
			r.Use(kessel.EnforceOrgPermission("config_manager_profile_edit"))
			r.Post("/profiles", createProfile)
		})
	})

	return router, nil
}
