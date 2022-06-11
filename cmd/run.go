package cmd

import (
	"config-manager/api"
	dispatcherConsumer "config-manager/dispatcher-consumer"
	"config-manager/internal/config"
	inventoryConsumer "config-manager/inventory-consumer"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/peterbourgon/ff/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

// startModuleFn is a function definition that a module must implement in order
// be started by the run command.
type startModuleFn = func(
	ctx context.Context,
	errors chan<- error,
)

func run(cmd *cobra.Command, args []string) error {
	fs := config.FlagSet("config-manager", flag.ExitOnError)

	if err := ff.Parse(fs, args, ff.WithEnvVarPrefix("CM")); err != nil {
		return fmt.Errorf("cannot parse flags: %w", err)
	}

	modules, err := cmd.Flags().GetStringSlice("module")
	if err != nil {
		log.Error().Err(err).Msg("error getting modules")
		return err
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	errors := make(chan error, 1)

	level, err := zerolog.ParseLevel(config.DefaultConfig.LogLevel)
	if err != nil {
		log.Error().Err(err)
		return err
	}

	zerolog.SetGlobalLevel(level)

	switch config.DefaultConfig.LogFormat {
	case "text":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	metricsServer := echo.New()
	metricsServer.HideBanner = true
	metricsServer.GET(config.DefaultConfig.MetricsPath, echo.WrapHandler(promhttp.Handler()))

	ctx := context.Background() //TODO context.WithCancel

	for _, module := range modules {
		log.Info().Str("module", module).Msg("starting")

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

		startModule(ctx, errors)
	}

	log.Info().Int("port", config.DefaultConfig.MetricsPort).Str("service", "metrics").Msg("starting http server")
	go func() {
		errors <- metricsServer.Start(fmt.Sprintf("0.0.0.0:%d", config.DefaultConfig.MetricsPort))
	}()

	log.Debug().Msg("Config Manager started")

	// stop on signal or error, whatever comes first
	select {
	case signal := <-signals:
		log.Info().Msgf("Shutting down due to signal: %v", signal)
		return nil
	case err := <-errors:
		log.Error().Err(err).Msg("shutting down due to error")
		return err
	}
}
