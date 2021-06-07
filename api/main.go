package api

import (
	"config-manager/api/controllers"
	"config-manager/api/instrumentation"
	"config-manager/config"
	"config-manager/infrastructure"
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"

	echoPrometheus "github.com/globocom/echo-prometheus"

	_ "github.com/lib/pq"
)

func Start(ctx context.Context, cfg *viper.Viper, errors chan<- error) {
	config := config.Get()

	container := infrastructure.Container{Config: config}
	instrumentation.Start()

	spec, err := controllers.GetSwagger()
	if err != nil {
		panic(err)
	}
	server := container.Server()
	server.Use(echoPrometheus.MetricsMiddleware())
	server.GET(config.GetString("URL_Base_Path")+"/openapi.json", func(ctx echo.Context) error {
		return ctx.JSON(http.StatusOK, spec)
	})

	configManager := container.CMController()
	configManager.Routes(spec)
	configManager.Server.HideBanner = true

	go func() {
		errors <- configManager.Server.Start(fmt.Sprintf("0.0.0.0:%d", config.GetInt("Web_Port")))
	}()
}
