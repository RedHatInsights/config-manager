package config

import (
	"config-manager/internal/url"
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sgreben/flagvar"

	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
)

// Config stores values that are used to configure the application.
type Config struct {
	AppName                string
	AWSAccessKeyId         string
	AWSRegion              string
	AWSSecretAccessKey     string
	CloudConnectorClientID string
	CloudConnectorHost     flagvar.URL
	CloudConnectorPSK      string
	CloudConnectorTimeout  int
	DBHost                 string
	DBName                 string
	DBPass                 string
	DBPort                 int
	DBUser                 string
	DispatcherHost         flagvar.URL
	DispatcherPSK          string
	DispatcherTimeout      int
	InventoryHost          flagvar.URL
	InventoryTimeout       int
	KafkaBrokers           flagvar.Strings
	KafkaConsumerOffset    int64
	KafkaGroupID           string
	KafkaInventoryTopic    string
	KafkaPassword          string
	KafkaUsername          string
	KafkaCAPath            string
	KafkaSaslMechanism     string
	KafkaSecurityProtocol  string
	KesselEnabled          bool
	KesselURL              string
	KesselAuthEnabled      bool
	KesselAuthClientID     string
	KesselAuthClientSecret string
	KesselAuthOIDCIssuer   string
	KesselInsecure         bool
	LogBatchFrequency      time.Duration
	LogFormat              flagvar.Enum
	LogGroup               string
	LogLevel               flagvar.Enum
	LogStream              string
	MetricsPath            string
	MetricsPort            int
	Modules                flagvar.EnumSetCSV
	PlaybookFiles          string
	ServiceConfig          string
	StaleEventDuration     time.Duration
	TenantTranslatorHost   string
	URLPathPrefix          string
	WebPort                int
}

func (c *Config) URLBasePath(apiVersion string) string {
	return path.Join("/", c.URLPathPrefix, c.AppName, apiVersion)
}

// DefaultConfig is the default configuration variable, providing access to
// configuration values globally.
var DefaultConfig Config = Config{
	AppName:                "config-manager",
	AWSAccessKeyId:         os.Getenv("CW_AWS_ACCESS_KEY_ID"),
	AWSRegion:              "us-east-1",
	AWSSecretAccessKey:     os.Getenv("CW_AWS_SECRET_ACCESS_KEY"),
	CloudConnectorClientID: "config-manager",
	CloudConnectorHost:     flagvar.URL{Value: url.MustParse("http://cloud-connector:8080")},
	CloudConnectorPSK:      "",
	CloudConnectorTimeout:  10,
	DBHost:                 "localhost",
	DBName:                 "insights",
	DBPass:                 "insights",
	DBPort:                 5432,
	DBUser:                 "insights",
	DispatcherHost:         flagvar.URL{Value: url.MustParse("http://playbook-dispatcher-api:8000")},
	DispatcherPSK:          "",
	DispatcherTimeout:      10,
	InventoryHost:          flagvar.URL{Value: url.MustParse("http://host-inventory-service:8000")},
	InventoryTimeout:       10,
	KafkaBrokers:           flagvar.Strings{Values: []string{"localhost:9094"}},
	KafkaConsumerOffset:    0,
	KafkaGroupID:           "config-manager",
	KafkaInventoryTopic:    "platform.inventory.events",
	KafkaPassword:          "",
	KafkaUsername:          "",
	KafkaCAPath:            "",
	KafkaSaslMechanism:     "",
	KafkaSecurityProtocol:  "",
	KesselEnabled:          false,
	KesselURL:              "localhost:9091",
	KesselAuthEnabled:      false,
	KesselAuthClientID:     "",
	KesselAuthClientSecret: "",
	KesselAuthOIDCIssuer:   "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token",
	KesselInsecure:         true,
	LogBatchFrequency:      10 * time.Second,
	LogFormat:              flagvar.Enum{Choices: []string{"json", "text"}, Value: "json"},
	LogGroup:               "platform-dev",
	LogLevel:               flagvar.Enum{Choices: []string{"panic", "fatal", "error", "warn", "info", "debug", "trace"}, Value: "info"},
	LogStream: func() string {
		hostname, err := os.Hostname()
		if err != nil {
			panic(err)
		}
		return hostname
	}(),
	MetricsPath:          "/metrics",
	MetricsPort:          9000,
	Modules:              flagvar.EnumSetCSV{Choices: []string{"http-api", "dispatcher-consumer", "inventory-consumer"}, Value: map[string]bool{}},
	PlaybookFiles:        "./playbooks/",
	ServiceConfig:        `{"insights":"enabled","compliance_openscap":"enabled","remediations":"enabled"}`,
	StaleEventDuration:   24 * time.Hour,
	TenantTranslatorHost: "",
	URLPathPrefix:        "api",
	WebPort:              8081,
}

func init() {
	if clowder.IsClowderEnabled() {
		DefaultConfig.AWSAccessKeyId = clowder.LoadedConfig.Logging.Cloudwatch.AccessKeyId
		DefaultConfig.AWSRegion = clowder.LoadedConfig.Logging.Cloudwatch.Region
		DefaultConfig.AWSSecretAccessKey = clowder.LoadedConfig.Logging.Cloudwatch.SecretAccessKey
		DefaultConfig.DBHost = clowder.LoadedConfig.Database.Hostname
		DefaultConfig.DBName = clowder.LoadedConfig.Database.Name
		DefaultConfig.DBPass = clowder.LoadedConfig.Database.Password
		DefaultConfig.DBPort = clowder.LoadedConfig.Database.Port
		DefaultConfig.DBUser = clowder.LoadedConfig.Database.Username
		DefaultConfig.KafkaBrokers.Values = clowder.KafkaServers
		if clowder.LoadedConfig.Kafka != nil {
			if len(clowder.LoadedConfig.Kafka.Brokers) >= 1 {
				broker := clowder.LoadedConfig.Kafka.Brokers[0]
				if broker.Authtype != nil {

					DefaultConfig.KafkaUsername = *broker.Sasl.Username
					DefaultConfig.KafkaPassword = *broker.Sasl.Password
					DefaultConfig.KafkaSaslMechanism = *broker.Sasl.SaslMechanism
					DefaultConfig.KafkaSecurityProtocol = *broker.Sasl.SecurityProtocol

					if broker.Cacert != nil {
						caPath, err := clowder.LoadedConfig.KafkaCa(broker)
						if err != nil {
							log.Fatal().Err(err).Msg("Kafka CA cert failed to write")
						}

						DefaultConfig.KafkaCAPath = caPath
					}
				}
			}
			for requestedName, topicConfig := range clowder.KafkaTopics {
				switch requestedName {
				case "platform.inventory.events":
					DefaultConfig.KafkaInventoryTopic = topicConfig.Name
				}
			}
		}
		DefaultConfig.LogGroup = clowder.LoadedConfig.Logging.Cloudwatch.LogGroup
		DefaultConfig.MetricsPath = clowder.LoadedConfig.MetricsPath
		DefaultConfig.MetricsPort = clowder.LoadedConfig.MetricsPort
		DefaultConfig.WebPort = *clowder.LoadedConfig.PublicPort

		if DefaultConfig.KesselEnabled {
			for _, e := range clowder.LoadedConfig.Endpoints {
				if e.App == "kessel-inventory-api" {
					DefaultConfig.KesselURL = fmt.Sprintf("%s:%d", e.Hostname, e.Port)
				}
			}
		}
	}
}

// FlagSet creates a new FlagSet, defined with flags for each struct field in
// the DefaultConfig variable.
func FlagSet(name string, errorHandling flag.ErrorHandling) *flag.FlagSet {
	fs := flag.NewFlagSet(name, errorHandling)

	fs.StringVar(&DefaultConfig.AppName, "app-name", DefaultConfig.AppName, "name of the application used in the URL path")
	fs.StringVar(&DefaultConfig.AWSAccessKeyId, "aws-access-key-id", DefaultConfig.AWSAccessKeyId, "CloudWatch access key ID")
	fs.StringVar(&DefaultConfig.AWSRegion, "aws-region", DefaultConfig.AWSRegion, "CloudWatch AWS region")
	fs.StringVar(&DefaultConfig.AWSSecretAccessKey, "aws-secret-access-key", DefaultConfig.AWSSecretAccessKey, "CloudWatch secret access key")
	fs.StringVar(&DefaultConfig.CloudConnectorClientID, "cloud-connector-client-id", DefaultConfig.CloudConnectorClientID, "client ID to use when authenticating to cloud-connector")
	fs.Var(&DefaultConfig.CloudConnectorHost, "cloud-connector-host", fmt.Sprintf("hostname for the cloud-connector service (%v)", DefaultConfig.CloudConnectorHost.Help()))
	fs.StringVar(&DefaultConfig.CloudConnectorPSK, "cloud-connector-psk", DefaultConfig.CloudConnectorPSK, "preshared key from config-manager")
	fs.IntVar(&DefaultConfig.CloudConnectorTimeout, "cloud-connector-timeout", DefaultConfig.CloudConnectorTimeout, "number of seconds before timing out HTTP requests to cloud-connector")
	fs.StringVar(&DefaultConfig.DBHost, "db-host", DefaultConfig.DBHost, "database host")
	fs.StringVar(&DefaultConfig.DBName, "db-name", DefaultConfig.DBName, "database name")
	fs.StringVar(&DefaultConfig.DBPass, "db-pass", DefaultConfig.DBPass, "database password")
	fs.IntVar(&DefaultConfig.DBPort, "db-port", DefaultConfig.DBPort, "database port")
	fs.StringVar(&DefaultConfig.DBUser, "db-user", DefaultConfig.DBUser, "database user")
	fs.Var(&DefaultConfig.DispatcherHost, "dispatcher-host", fmt.Sprintf("hostname for the playbook-dispatcher service (%v)", DefaultConfig.DispatcherHost.Help()))
	fs.StringVar(&DefaultConfig.DispatcherPSK, "dispatcher-psk", DefaultConfig.DispatcherPSK, "preshared key from playbook-dispatcher")
	fs.IntVar(&DefaultConfig.DispatcherTimeout, "dispatcher-timeout", DefaultConfig.DispatcherTimeout, "number of seconds before timing out HTTP requests to playbook-dispatcher")
	fs.Var(&DefaultConfig.InventoryHost, "inventory-host", fmt.Sprintf("hostname for the host-inventory service (%v)", DefaultConfig.InventoryHost.Help()))
	fs.IntVar(&DefaultConfig.InventoryTimeout, "inventory-timeout", DefaultConfig.InventoryTimeout, "number of seconds before timing out HTTP requests to host-inventory")
	fs.Var(&DefaultConfig.KafkaBrokers, "kafka-brokers", "kafka bootstrap broker addresses")
	fs.Int64Var(&DefaultConfig.KafkaConsumerOffset, "kafka-consumer-offset", DefaultConfig.KafkaConsumerOffset, "kafka consumer offset")
	fs.StringVar(&DefaultConfig.KafkaGroupID, "kafka-group-id", DefaultConfig.KafkaGroupID, "kafka group ID")
	fs.StringVar(&DefaultConfig.KafkaInventoryTopic, "kafka-inventory-topic", DefaultConfig.KafkaInventoryTopic, "host-inventory events topic name")
	fs.StringVar(&DefaultConfig.KafkaPassword, "kafka-password", DefaultConfig.KafkaPassword, "managed kafka auth password")
	fs.StringVar(&DefaultConfig.KafkaUsername, "kafka-username", DefaultConfig.KafkaUsername, "managed kafka auth username")
	fs.StringVar(&DefaultConfig.KafkaCAPath, "kafka-cacert-path", DefaultConfig.KafkaCAPath, "managed kafka cacert path")
	fs.StringVar(&DefaultConfig.KafkaSaslMechanism, "kafka-sasl-mechanism", DefaultConfig.KafkaSaslMechanism, "managed kafka sasl mechanism")
	fs.StringVar(&DefaultConfig.KafkaSecurityProtocol, "kafka-security-protocol", DefaultConfig.KafkaSecurityProtocol, "managed kafka security protocol")
	fs.BoolVar(&DefaultConfig.KesselEnabled, "kessel-enabled", DefaultConfig.KesselEnabled, "enable authorization using Kessel")
	fs.StringVar(&DefaultConfig.KesselURL, "kessel-url", DefaultConfig.KesselURL, "Kessel API URL")
	fs.BoolVar(&DefaultConfig.KesselAuthEnabled, "kessel-auth-enabled", DefaultConfig.KesselAuthEnabled, "enable Kessel client authentication")
	fs.StringVar(&DefaultConfig.KesselAuthClientID, "kessel-auth-client-id", DefaultConfig.KesselAuthClientID, "Kessel authentication client id")
	fs.StringVar(&DefaultConfig.KesselAuthClientSecret, "kessel-auth-client-secret", DefaultConfig.KesselAuthClientSecret, "Kessel authentication client secret")
	fs.StringVar(&DefaultConfig.KesselAuthOIDCIssuer, "kessel-auth-oidc-issuer", DefaultConfig.KesselAuthOIDCIssuer, "Kessel authentication OIDC issuer")
	fs.BoolVar(&DefaultConfig.KesselInsecure, "kessel-insecure", DefaultConfig.KesselInsecure, "disable TLS for the Kessel client")
	fs.DurationVar(&DefaultConfig.LogBatchFrequency, "log-batch-frequency", DefaultConfig.LogBatchFrequency, "CloudWatch batch log frequency")
	fs.Var(&DefaultConfig.LogFormat, "log-format", fmt.Sprintf("structured logging output format (%v)", DefaultConfig.LogFormat.Help()))
	fs.StringVar(&DefaultConfig.LogGroup, "log-group", DefaultConfig.LogGroup, "CloudWatch log group")
	fs.Var(&DefaultConfig.LogLevel, "log-level", fmt.Sprintf("verbosity level for logging (%v)", DefaultConfig.LogLevel.Help()))
	fs.StringVar(&DefaultConfig.LogStream, "log-stream", DefaultConfig.LogStream, "CloudWatch log stream")
	fs.StringVar(&DefaultConfig.MetricsPath, "metrics-path", DefaultConfig.MetricsPath, "base path on which metrics HTTP server responds")
	fs.IntVar(&DefaultConfig.MetricsPort, "metrics-port", DefaultConfig.MetricsPort, "port on which metrics HTTP server listens")
	fs.Var(&DefaultConfig.Modules, "module", fmt.Sprintf("config-manager modules to execute (%v)", DefaultConfig.Modules.Help()))
	fs.StringVar(&DefaultConfig.PlaybookFiles, "playbook-files", DefaultConfig.PlaybookFiles, "path to playbook directory")
	fs.StringVar(&DefaultConfig.ServiceConfig, "service-config", DefaultConfig.ServiceConfig, "default state configuration")
	fs.DurationVar(&DefaultConfig.StaleEventDuration, "stale-event-duration", DefaultConfig.StaleEventDuration, "duration of time after which inventory events are discarded")
	fs.StringVar(&DefaultConfig.TenantTranslatorHost, "tenant-translator-host", DefaultConfig.TenantTranslatorHost, "tenant translator service host")
	fs.IntVar(&DefaultConfig.WebPort, "web-port", DefaultConfig.WebPort, "port on which HTTP API server listens")
	fs.StringVar(&DefaultConfig.URLPathPrefix, "url-path-prefix", DefaultConfig.URLPathPrefix, "generic prefix used in the URL path")

	return fs
}
