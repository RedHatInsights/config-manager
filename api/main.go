package api

import (
	"config-manager/api/instrumentation"
	"config-manager/infrastructure"
	"config-manager/internal/config"
	v1 "config-manager/internal/http/v1"
	"context"
	"fmt"
	"net/http"
)

// Start creates a new HTTP server, sets up route handlers, and starts
// listening for HTTP requests. It is the module entrypoint for the REST API,
// conforming to the startModuleFn type definition in config-manager/cmd.
func Start(ctx context.Context, errors chan<- error) {
	container := infrastructure.Container{}
	instrumentation.Start()
	container.Database()

	router, err := v1.NewMux()
	if err != nil {
		errors <- err
	}

	go func() {
		errors <- http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", config.DefaultConfig.WebPort), router)
	}()
}
