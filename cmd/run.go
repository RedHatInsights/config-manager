package cmd

import (
	"config-manager/api"
	"config-manager/config"
	dispatcherConsumer "config-manager/dispatcher-consumer"
	inventoryConsumer "config-manager/inventory-consumer"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	echoPrometheus "github.com/globocom/echo-prometheus"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type startModuleFn = func(
	ctx context.Context,
	cfg *viper.Viper,
	errors chan<- error,
)

func run(cmd *cobra.Command, args []string) error {
	modules, err := cmd.Flags().GetStringSlice("module")
	if err != nil {
		log.Println("Error getting modules: ", err)
		return err
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	errors := make(chan error, 1)

	cfg := config.Get()

	metricsServer := echo.New()
	metricsServer.HideBanner = true
	metricsServer.Use(echoPrometheus.MetricsMiddleware())
	metricsServer.GET(cfg.GetString("Metrics_Path"), echo.WrapHandler(promhttp.Handler()))

	ctx := context.Background() //TODO context.WithCancel

	for _, module := range modules {
		fmt.Printf("Starting module %s\n", module)

		var startModule startModuleFn

		switch module {
		case moduleApi:
			startModule = api.Start
		case moduleDispatcherConsumer:
			startModule = dispatcherConsumer.Start
		case moduleInventoryConsumer:
			startModule = inventoryConsumer.Start
		default:
			return fmt.Errorf("Unknown module %s", module)
		}

		startModule(ctx, cfg, errors)
	}

	fmt.Printf("Listening on service port %d\n", cfg.GetInt("Metrics_Port"))
	go func() {
		errors <- metricsServer.Start(fmt.Sprintf("0.0.0.0:%d", cfg.GetInt("Metrics_Port")))
	}()

	log.Println("Config Manager started")

	// stop on signal or error, whatever comes first
	select {
	case signal := <-signals:
		log.Println("Shutting down due to signal: ", signal)
		return nil
	case error := <-errors:
		log.Println("Shutting down due to error: ", error)
		return error
	}
}
