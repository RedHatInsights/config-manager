package httpapi

import (
	"config-manager/internal/config"
	v1 "config-manager/internal/http/v1"
	"context"
	"fmt"
	"net/http"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/rs/zerolog/log"
)

var Command ffcli.Command = ffcli.Command{
	Name:      "http-api",
	ShortHelp: "Run the HTTP API server",
	Exec: func(ctx context.Context, args []string) error {
		log.Info().Str("command", "http-api").Msg("starting command")

		router, err := v1.NewMux()
		if err != nil {
			return fmt.Errorf("cannot create HTTP router: %w", err)
		}

		if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", config.DefaultConfig.WebPort), router); err != nil {
			return fmt.Errorf("cannot listen on port %v: %w", config.DefaultConfig.WebPort, err)
		}

		return nil
	},
}
