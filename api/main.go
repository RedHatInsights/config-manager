package api

import (
	"config-manager/api/controllers"
	"config-manager/api/instrumentation"
	"config-manager/infrastructure"
	"config-manager/internal/config"
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	echoPrometheus "github.com/globocom/echo-prometheus"

	_ "github.com/lib/pq"
)

// Start creates a new Echo HTTP server, sets up route handlers, and starts
// listening for HTTP requests. It is the module entrypoint for the REST API,
// conforming to the startModuleFn type definition in config-manager/cmd.
func Start(ctx context.Context, errors chan<- error) {

	container := infrastructure.Container{}
	instrumentation.Start()

	spec, err := controllers.GetSwagger()
	if err != nil {
		panic(err)
	}
	server := container.Server()
	server.Use(echoPrometheus.MetricsMiddleware())
	server.GET(config.DefaultConfig.URLBasePath()+"/openapi.json", func(ctx echo.Context) error {
		return ctx.JSON(http.StatusOK, spec)
	})

	configManager := container.CMController()
	configManager.Routes(spec)
	configManager.Server.HideBanner = true

	go func() {
		errors <- configManager.Server.Start(fmt.Sprintf("0.0.0.0:%d", config.DefaultConfig.WebPort))
	}()
}
