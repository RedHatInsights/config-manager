package config

import (
	"config-manager/internal/url"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sgreben/flagvar"

	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
)

// Config stores values that are used to configure the application.
type Config struct {
	APIVersion              string
	AppName                 string
	AWSAccessKeyId          string
	AWSRegion               string
	AWSSecretAccessKey      string
	CloudConnectorClientID  string
	CloudConnectorHost      flagvar.URL
	CloudConnectorImpl      flagvar.Enum
	CloudConnectorPSK       string
	CloudConnectorTimeout   int
	DBHost                  string
	DBName                  string
	DBPass                  string
	DBPort                  int
	DBUser                  string
	DispatcherBatchSize     int
	DispatcherHost          flagvar.URL
	DispatcherImpl          flagvar.Enum
	DispatcherPSK           string
	DispatcherTimeout       int
	InventoryHost           flagvar.URL
	InventoryImpl           flagvar.Enum
	InventoryTimeout        int
	KafkaBrokers            flagvar.Strings
	KafkaConsumerOffset     int64
	KafkaDispatcherTopic    string
	KafkaGroupID            string
	KafkaInventoryTopic     string
	KafkaSystemProfileTopic string
	KafkaUsername			string
	KafkaPassword			string
	KafkaSASLMechanism		string
	KafkaSecurityProtocol	string
	KafkaAuthType			string
	LogBatchFrequency       time.Duration
	LogFormat               flagvar.Enum
	LogGroup                string
	LogLevel                flagvar.Enum
	LogStream               string
	MetricsPath             string
	MetricsPort             int
	Modules                 flagvar.EnumSetCSV
	PlaybookFiles           string
	PlaybookHost            flagvar.URL
	PlaybookPath            string
	ServiceConfig           string
	URLPathPrefix           string
	WebPort                 int
}

func (c *Config) URLBasePath() string {
	return filepath.Join("/", c.URLPathPrefix, c.AppName, c.APIVersion)
}

// DefaultConfig is the default configuration variable, providing access to
// configuration values globally.
var DefaultConfig Config = Config{
	APIVersion:              "v1",
	AppName:                 "config-manager",
	AWSAccessKeyId:          os.Getenv("CW_AWS_ACCESS_KEY_ID"),
	AWSRegion:               "us-east-1",
	AWSSecretAccessKey:      os.Getenv("CW_AWS_SECRET_ACCESS_KEY"),
	CloudConnectorClientID:  "config-manager",
	CloudConnectorHost:      flagvar.URL{Value: url.MustParse("http://cloud-connector:8080")},
	CloudConnectorImpl:      flagvar.Enum{Choices: []string{"mock", "impl"}, Value: "impl"},
	CloudConnectorPSK:       "",
	CloudConnectorTimeout:   10,
	DBHost:                  "localhost",
	DBName:                  "insights",
	DBPass:                  "insights",
	DBPort:                  5432,
	DBUser:                  "insights",
	DispatcherBatchSize:     50,
	DispatcherHost:          flagvar.URL{Value: url.MustParse("http://playbook-dispatcher-api:8000")},
	DispatcherImpl:          flagvar.Enum{Choices: []string{"mock", "impl"}, Value: "impl"},
	DispatcherPSK:           "",
	DispatcherTimeout:       10,
	InventoryHost:           flagvar.URL{Value: url.MustParse("http://host-inventory-service:8000")},
	InventoryImpl:           flagvar.Enum{Choices: []string{"mock", "impl"}, Value: "impl"},
	InventoryTimeout:        10,
	KafkaBrokers:            flagvar.Strings{Values: []string{"localhost:9094"}},
	KafkaConsumerOffset:     0,
	KafkaDispatcherTopic:    "platform.playbook-dispatcher.runs",
	KafkaGroupID:            "config-manager",
	KafkaInventoryTopic:     "platform.inventory.events",
	KafkaSystemProfileTopic: "platform.inventory.system-profile",
	KafkaUsername: 			 os.Getenv("KAFKA_USER_NAME"),
	KafkaPassword:			 os.Getenv("KAFKA_PASSWORD"),
	KafkaSASLMechanism:		 os.Getenv("KAFKA_SASL_MECH"),
	KafkaSecurityProtocol: 	 os.Getenv("KAFKA_SECURITY_PROTOCOL"),
	KafkaAuthType:           "sasl",
	LogBatchFrequency:       10 * time.Second,
	LogFormat:               flagvar.Enum{Choices: []string{"json", "text"}, Value: "json"},
	LogGroup:                "platform-dev",
	LogLevel:                flagvar.Enum{Choices: []string{"panic", "fatal", "error", "warn", "info", "debug", "trace"}, Value: "info"},
	LogStream: func() string {
		hostname, err := os.Hostname()
		if err != nil {
			panic(err)
		}
		return hostname
	}(),
	MetricsPath:   "/metrics",
	MetricsPort:   9000,
	Modules:       flagvar.EnumSetCSV{Choices: []string{"api", "dispatcher-consumer", "inventory-consumer"}, Value: map[string]bool{}},
	PlaybookFiles: "./playbooks/",
	PlaybookHost:  flagvar.URL{Value: url.MustParse("https://cert.cloud.stage.redhat.com")},
	PlaybookPath:  "/api/config-manager/v1/states/%v/playbook",
	ServiceConfig: `{"insights":"enabled","compliance_openscap":"enabled","remediations":"enabled"}`,
	URLPathPrefix: "api",
	WebPort:       8081,
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
		DefaultConfig.KafkaBrokers.Values = clowder.LoadedConfig.Kafka.Brokers
		DefaultConfig.LogGroup = clowder.LoadedConfig.Logging.Cloudwatch.LogGroup
		DefaultConfig.MetricsPath = clowder.LoadedConfig.MetricsPath
		DefaultConfig.MetricsPort = clowder.LoadedConfig.MetricsPort
		DefaultConfig.WebPort = *clowder.LoadedConfig.PublicPort
		DefaultConfig.KafkaUsername = clowder.LoadedConfig.Kafka.Brokers[0].Sasl.Username
		DefaultConfig.KafkaPassword = clowder.LoadedConfig.Kafka.Brokers[0].Sasl.Password
		DefaultConfig.KafkaSASLMechanism = clowder.LoadedConfig.Kafka.Brokers[0].Sasl.saslMechanism
		DefaultConfig.KafkaSecurityProtocol = clowder.LoadedConfig.Kafka.Brokers[0].Sasl.securityProtocol
		DefaultConfig.KafkaAuthType = clowder.LoadedConfig.Kafka.Brokers[0].Authtype 
	}
}

// FlagSet creates a new FlagSet, defined with flags for each struct field in
// the DefaultConfig variable.
func FlagSet(name string, errorHandling flag.ErrorHandling) *flag.FlagSet {
	fs := flag.NewFlagSet(name, errorHandling)

	fs.StringVar(&DefaultConfig.APIVersion, "api-version", DefaultConfig.APIVersion, "API version used in the URL path")
	fs.StringVar(&DefaultConfig.AppName, "app-name", DefaultConfig.AppName, "name of the application used in the URL path")
	fs.StringVar(&DefaultConfig.AWSAccessKeyId, "aws-access-key-id", DefaultConfig.AWSAccessKeyId, "CloudWatch access key ID")
	fs.StringVar(&DefaultConfig.AWSRegion, "aws-region", DefaultConfig.AWSRegion, "CloudWatch AWS region")
	fs.StringVar(&DefaultConfig.AWSSecretAccessKey, "aws-secret-access-key", DefaultConfig.AWSSecretAccessKey, "CloudWatch secret access key")
	fs.StringVar(&DefaultConfig.CloudConnectorClientID, "cloud-connector-client-id", DefaultConfig.CloudConnectorClientID, "client ID to use when authenticating to cloud-connector")
	fs.Var(&DefaultConfig.CloudConnectorHost, "cloud-connector-host", fmt.Sprintf("hostname for the cloud-connector service (%v)", DefaultConfig.CloudConnectorHost.Help()))
	fs.Var(&DefaultConfig.CloudConnectorImpl, "cloud-connector-impl", fmt.Sprintf("use either a mock or real implementation of cloud-connector (%v)", DefaultConfig.CloudConnectorImpl.Help()))
	fs.StringVar(&DefaultConfig.CloudConnectorPSK, "cloud-connector-psk", DefaultConfig.CloudConnectorPSK, "preshared key from config-manager")
	fs.IntVar(&DefaultConfig.CloudConnectorTimeout, "cloud-connector-timeout", DefaultConfig.CloudConnectorTimeout, "number of seconds before timing out HTTP requests to cloud-connector")
	fs.StringVar(&DefaultConfig.DBHost, "db-host", DefaultConfig.DBHost, "database host")
	fs.StringVar(&DefaultConfig.DBName, "db-name", DefaultConfig.DBName, "database name")
	fs.StringVar(&DefaultConfig.DBPass, "db-pass", DefaultConfig.DBPass, "database password")
	fs.IntVar(&DefaultConfig.DBPort, "db-port", DefaultConfig.DBPort, "database port")
	fs.StringVar(&DefaultConfig.DBUser, "db-user", DefaultConfig.DBUser, "database user")
	fs.IntVar(&DefaultConfig.DispatcherBatchSize, "dispatcher-batch-size", DefaultConfig.DispatcherBatchSize, "size of batches to transmit to playbook-dispatcher")
	fs.Var(&DefaultConfig.DispatcherHost, "dispatcher-host", fmt.Sprintf("hostname for the playbook-dispatcher service (%v)", DefaultConfig.DispatcherHost.Help()))
	fs.Var(&DefaultConfig.DispatcherImpl, "dispatcher-impl", fmt.Sprintf("use either a mock or real implementation of playbook-dispatcher (%v)", DefaultConfig.DispatcherImpl.Help()))
	fs.StringVar(&DefaultConfig.DispatcherPSK, "dispatcher-psk", DefaultConfig.DispatcherPSK, "preshared key from playbook-dispatcher")
	fs.IntVar(&DefaultConfig.DispatcherTimeout, "dispatcher-timeout", DefaultConfig.DispatcherTimeout, "number of seconds before timing out HTTP requests to playbook-dispatcher")
	fs.Var(&DefaultConfig.InventoryHost, "inventory-host", fmt.Sprintf("hostname for the host-inventory service (%v)", DefaultConfig.InventoryHost.Help()))
	fs.Var(&DefaultConfig.InventoryImpl, "inventory-impl", fmt.Sprintf("use either a mock or real implementation of host-inventory (%v)", DefaultConfig.InventoryImpl.Help()))
	fs.IntVar(&DefaultConfig.InventoryTimeout, "inventory-timeout", DefaultConfig.InventoryTimeout, "number of seconds before timing out HTTP requests to host-inventory")
	fs.Var(&DefaultConfig.KafkaBrokers, "kafka-brokers", "kafka bootstrap broker addresses")
	fs.Int64Var(&DefaultConfig.KafkaConsumerOffset, "kafka-consumer-offset", DefaultConfig.KafkaConsumerOffset, "kafka consumer offset")
	fs.StringVar(&DefaultConfig.KafkaDispatcherTopic, "kafka-dispatcher-topic", DefaultConfig.KafkaDispatcherTopic, "playbook-dispatcher runs topic name")
	fs.StringVar(&DefaultConfig.KafkaGroupID, "kafka-group-id", DefaultConfig.KafkaGroupID, "kafka group ID")
	fs.StringVar(&DefaultConfig.KafkaInventoryTopic, "kafka-inventory-topic", DefaultConfig.KafkaInventoryTopic, "host-inventory events topic name")
	fs.StringVar(&DefaultConfig.KafkaSystemProfileTopic, "kafka-system-profile-topic", DefaultConfig.KafkaSystemProfileTopic, "host-inventory system-profile topic name")
	fs.DurationVar(&DefaultConfig.LogBatchFrequency, "log-batch-frequency", DefaultConfig.LogBatchFrequency, "CloudWatch batch log frequency")
	fs.Var(&DefaultConfig.LogFormat, "log-format", fmt.Sprintf("structured logging output format (%v)", DefaultConfig.LogFormat.Help()))
	fs.StringVar(&DefaultConfig.LogGroup, "log-group", DefaultConfig.LogGroup, "CloudWatch log group")
	fs.Var(&DefaultConfig.LogLevel, "log-level", fmt.Sprintf("verbosity level for logging (%v)", DefaultConfig.LogLevel.Help()))
	fs.StringVar(&DefaultConfig.LogStream, "log-stream", DefaultConfig.LogStream, "CloudWatch log stream")
	fs.StringVar(&DefaultConfig.MetricsPath, "metrics-path", DefaultConfig.MetricsPath, "base path on which metrics HTTP server responds")
	fs.IntVar(&DefaultConfig.MetricsPort, "metrics-port", DefaultConfig.MetricsPort, "port on which metrics HTTP server listens")
	fs.Var(&DefaultConfig.Modules, "module", fmt.Sprintf("config-manager modules to execute (%v)", DefaultConfig.Modules.Help()))
	fs.StringVar(&DefaultConfig.PlaybookFiles, "playbook-files", DefaultConfig.PlaybookFiles, "path to playbook directory")
	fs.Var(&DefaultConfig.PlaybookHost, "playbook-host", fmt.Sprintf("default host from which to download playbooks (%v)", DefaultConfig.PlaybookHost.Help()))
	fs.StringVar(&DefaultConfig.PlaybookPath, "playbook-path", DefaultConfig.PlaybookPath, "path component for playbook downloads")
	fs.IntVar(&DefaultConfig.WebPort, "web-port", DefaultConfig.WebPort, "port on which HTTP API server listens")
	fs.StringVar(&DefaultConfig.URLPathPrefix, "url-path-prefix", DefaultConfig.URLPathPrefix, "generic prefix used in the URL path")
	fs.StringVar(&DefaultConfig.ServiceConfig, "service-config", DefaultConfig.ServiceConfig, "default state configuration")

	return fs
}
