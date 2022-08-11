package main

import (
	"config-manager/internal/cmd/dispatcherconsumer"
	"config-manager/internal/cmd/httpapi"
	"config-manager/internal/cmd/inventoryconsumer"
	"config-manager/internal/config"
	"config-manager/internal/db"
	"config-manager/internal/logging/cloudwatch"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	root := ffcli.Command{
		FlagSet: config.FlagSet("config-manager", flag.ExitOnError),
		Options: []ff.Option{
			ff.WithEnvVarPrefix("CM"),
		},
		Subcommands: []*ffcli.Command{
			&httpapi.Command,
			&inventoryconsumer.Command,
			&dispatcherconsumer.Command,
		},
		Exec: func(ctx context.Context, args []string) error {
			modules := map[string]*ffcli.Command{
				"dispatcher-consumer": &dispatcherconsumer.Command,
				"http-api":            &httpapi.Command,
				"inventory-consumer":  &inventoryconsumer.Command,
			}

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

			for _, module := range config.DefaultConfig.Modules.Values() {
				subcommand := modules[module]
				go func() {
					if err := subcommand.Exec(ctx, args); err != nil {
						log.Fatal().Err(err).Msg("cannot run subcommand")
					}
				}()
			}

			<-quit

			return nil
		},
	}

	if err := root.Parse(os.Args[1:]); err != nil {
		log.Fatal().Err(err).Msg("unable to parse flags")
	}

	level, err := zerolog.ParseLevel(config.DefaultConfig.LogLevel.Value)
	if err != nil {
		log.Fatal().Err(err).Str("level", config.DefaultConfig.LogLevel.Value).Msgf("cannot parse log level")
	}

	zerolog.SetGlobalLevel(level)

	if level <= zerolog.DebugLevel {
		log.Logger = log.With().Caller().Logger()
	}

	log.Debug().Interface("config", config.DefaultConfig).Send()

	writers := make([]io.Writer, 0)
	switch config.DefaultConfig.LogFormat.Value {
	case "text":
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr})
	default:
		writers = append(writers, os.Stderr)
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

	go func() {
		mux := http.NewServeMux()
		mux.Handle(config.DefaultConfig.MetricsPath, promhttp.Handler())
		if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", config.DefaultConfig.MetricsPort), mux); err != nil {
			log.Fatal().Err(err).Int("metrics-port", config.DefaultConfig.MetricsPort).Msg("cannot listen on port")
		}
	}()

	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable",
		config.DefaultConfig.DBUser,
		config.DefaultConfig.DBPass,
		config.DefaultConfig.DBName,
		config.DefaultConfig.DBHost,
		config.DefaultConfig.DBPort)

	if err := db.Open("pgx", connectionString); err != nil {
		log.Fatal().Err(err).Msg("cannot open database")
	}

	if err := db.Migrate("file://./db/migrations", false); err != nil {
		log.Fatal().Err(err).Msg("cannot migrate database")
	}

	if err := root.Run(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("unable to run command")
	}
}
