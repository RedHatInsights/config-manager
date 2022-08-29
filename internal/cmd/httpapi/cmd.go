package httpapi

import (
	"config-manager/internal/config"
	v1 "config-manager/internal/http/v1"
	v2 "config-manager/internal/http/v2"
	"context"
	"fmt"
	"net/http"
	"path"

	chiprometheus "github.com/766b/chi-prometheus"
	"github.com/go-chi/chi/v5"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/rs/zerolog/log"
)

var Command ffcli.Command = ffcli.Command{
	Name:      "http-api",
	ShortHelp: "Run the HTTP API server",
	Exec: func(ctx context.Context, args []string) error {
		log.Info().Str("command", "http-api").Msg("starting command")

		v1r, err := v1.NewMux()
		if err != nil {
			return fmt.Errorf("cannot create HTTP router: %w", err)
		}

		v2r, err := v2.NewMux()
		if err != nil {
			return fmt.Errorf("cannot create HTTP router: %w", err)
		}

		router := chi.NewMux()
		router.Use(chiprometheus.NewMiddleware(config.DefaultConfig.AppName))
		router.Mount(path.Join("/", config.DefaultConfig.URLPathPrefix, config.DefaultConfig.AppName, "v1"), v1r)
		router.Mount(path.Join("/", config.DefaultConfig.URLPathPrefix, config.DefaultConfig.AppName, "v2"), v2r)

		addr := fmt.Sprintf("0.0.0.0:%v", config.DefaultConfig.WebPort)
		log.Info().Str("addr", addr).Msg("listening and serving")
		if err := http.ListenAndServe(addr, router); err != nil {
			return fmt.Errorf("cannot listen on port %v: %w", config.DefaultConfig.WebPort, err)
		}

		return nil
	},
}
