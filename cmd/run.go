package cmd

import (
	"config-manager/api"
	"config-manager/config"
	dispatcherConsumer "config-manager/dispatcher-consumer"
	inventoryConsumer "config-manager/inventory-consumer"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startModuleFn is a function definition that a module must implement in order
// be started by the run command.
type startModuleFn = func(
	ctx context.Context,
	cfg *viper.Viper,
	errors chan<- error,
)

func run(cmd *cobra.Command, args []string) error {
	modules, err := cmd.Flags().GetStringSlice("module")
	if err != nil {
		log.Error().Err(err).Msg("error getting modules")
		return err
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	errors := make(chan error, 1)

	cfg := config.Get()

	level, err := zerolog.ParseLevel(cfg.GetString("Log_Level"))
	if err != nil {
		log.Error().Err(err)
		return err
	}

	zerolog.SetGlobalLevel(level)

	switch cfg.GetString("Log_Format") {
	case "text":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	metricsServer := echo.New()
	metricsServer.HideBanner = true
	metricsServer.GET(cfg.GetString("Metrics_Path"), echo.WrapHandler(promhttp.Handler()))

	ctx := context.Background() //TODO context.WithCancel

	for _, module := range modules {
		log.Printf("Starting module %s\n", module)

		var startModule startModuleFn

		switch module {
		case moduleApi:
			startModule = api.Start
		case moduleDispatcherConsumer:
			startModule = dispatcherConsumer.Start
		case moduleInventoryConsumer:
			startModule = inventoryConsumer.Start
		default:
			return fmt.Errorf("unknown module %s", module)
		}

		startModule(ctx, cfg, errors)
	}

	log.Printf("Listening on service port %d\n", cfg.GetInt("Metrics_Port"))
	go func() {
		errors <- metricsServer.Start(fmt.Sprintf("0.0.0.0:%d", cfg.GetInt("Metrics_Port")))
	}()

	log.Info().Msg("Config Manager started")

	// stop on signal or error, whatever comes first
	select {
	case signal := <-signals:
		log.Info().Msgf("Shutting down due to signal: ", signal)
		return nil
	case error := <-errors:
		log.Info().Msgf("Shutting down due to error: ", error)
		return error
	}
}
