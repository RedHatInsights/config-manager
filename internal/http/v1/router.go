package v1

//go:generate oapi-codegen -config oapi-codegen.yml ./openapi.yaml

import (
	"config-manager/internal/config"
	"config-manager/internal/http/render"
	"fmt"
	"net/http"
	"path"

	chiprometheus "github.com/766b/chi-prometheus"
	oapimiddleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/redhatinsights/platform-go-middlewares/identity"
	"github.com/rs/zerolog/log"
)

func NewMux() (*chi.Mux, error) {
	spec, err := GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("cannot get OpenAPI spec: %w", err)
	}

	router := chi.NewMux()

	router.Use(httplog.RequestLogger(httplog.NewLogger("v1", httplog.Options{
		LogLevel: config.DefaultConfig.LogLevel.Value,
		JSON:     config.DefaultConfig.LogFormat.Value == "json",
	})))
	router.Use(identity.EnforceIdentity)
	router.Use(chiprometheus.NewMiddleware("config-manager"))
	router.Use(middleware.RequestID)
	router.Get(path.Join(config.DefaultConfig.URLBasePath(), "openapi.json"), func(w http.ResponseWriter, r *http.Request) {
		render.RenderJSON(w, r, http.StatusOK, spec, log.Logger)
	})

	router.Route(config.DefaultConfig.URLBasePath(), func(r chi.Router) {
		r.Use(oapimiddleware.OapiRequestValidator(spec))
		r.Post("/states", postStates)
		r.Get("/states", getStates)
		r.Get("/states/current", getCurrentState)
		r.Get("/states/{id}", getStateByID)
		r.Get("/states/{id}/playbook", getStatesIDPlaybook)
		r.Post("/states/preview", postStatesPreview)
		r.Post("/manage", postManage)
	})

	return router, nil
}
