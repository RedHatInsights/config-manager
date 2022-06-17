package main

import (
	"config-manager/api"
	dispatcherconsumer "config-manager/dispatcher-consumer"
	"config-manager/internal/config"
	"config-manager/internal/logging/cloudwatch"
	inventoryconsumer "config-manager/inventory-consumer"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func main() {
	root := ffcli.Command{
		FlagSet: config.FlagSet("config-manager", flag.ExitOnError),
		Options: []ff.Option{
			ff.WithEnvVarPrefix("CM"),
		},
		Subcommands: []*ffcli.Command{
			{
				Name: "run",
				FlagSet: func() *flag.FlagSet {
					fs := flag.NewFlagSet("run", flag.ExitOnError)
					fs.Var(&config.DefaultConfig.Modules, "m", fmt.Sprintf("config-manager modules to execute (%v)", config.DefaultConfig.Modules.Help()))
					return fs
				}(),
				Exec: func(ctx context.Context, args []string) error {
					signals := make(chan os.Signal, 1)
					signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
					errors := make(chan error, 1)

					level, err := zerolog.ParseLevel(config.DefaultConfig.LogLevel.Value)
					if err != nil {
						log.Fatal().Err(err).Msgf("cannot parse log level: %v", config.DefaultConfig.LogLevel)
						return err
					}

					zerolog.SetGlobalLevel(level)

					log.Debug().Interface("config", config.DefaultConfig).Send()

					writers := make([]io.Writer, 0)
					switch config.DefaultConfig.LogFormat.Value {
					case "json":
						writers = append(writers, os.Stderr)
					case "text":
						writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
					}

					if clowder.IsClowderEnabled() {
						cred := credentials.NewStaticCredentials(config.DefaultConfig.AWSAccessKeyId, config.DefaultConfig.AWSSecretAccessKey, "")
						awsCfg := aws.NewConfig().WithRegion(config.DefaultConfig.AWSRegion).WithCredentials(cred)
						batchWriter, err := cloudwatch.NewBatchWriter(config.DefaultConfig.LogGroup, config.DefaultConfig.LogStream, awsCfg, config.DefaultConfig.LogBatchFrequency)
						if err != nil {
							log.Error().Err(err).Msg("cannot create CloudWatch batch writer")
						}
						if batchWriter != nil {
							writers = append(writers, batchWriter)
						}
					}

					log.Logger = log.Output(zerolog.MultiLevelWriter(writers...))

					metricsServer := echo.New()
					metricsServer.HideBanner = true
					metricsServer.GET(config.DefaultConfig.MetricsPath, echo.WrapHandler(promhttp.Handler()))

					for _, module := range config.DefaultConfig.Modules.Values() {
						log.Info().Str("module", module).Msg("starting")

						var startModule func(
							ctx context.Context,
							errors chan<- error,
						)

						switch module {
						case "api":
							startModule = api.Start
						case "dispatcher-consumer":
							startModule = dispatcherconsumer.Start
						case "inventory-consumer":
							startModule = inventoryconsumer.Start
						default:
							return fmt.Errorf("unknown module %s", module)
						}

						startModule(context.Background(), errors)
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
				},
			},
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		log.Fatal().Err(err).Msg("cannot execute command")
	}
}
