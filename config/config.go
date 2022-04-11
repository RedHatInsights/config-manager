package config

import (
	"fmt"
	"os"

	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
	"github.com/spf13/viper"
)

// Get creates a Viper configuration registry with keys and default values,
// loads values from the environment automatically, and returns it.
func Get() *viper.Viper {
	options := viper.New()

	options.SetDefault("App_Name", "config-manager")
	options.SetDefault("API_Version", "v1")
	options.SetDefault("URL_Path_Prefix", "api")
	options.SetDefault("URL_Base_Path", buildURL(
		options.GetString("URL_Path_Prefix"),
		options.GetString("App_Name"),
		options.GetString("API_Version"),
	))

	if os.Getenv("CLOWDER_ENABLED") == "true" {
		cfg := clowder.LoadedConfig

		options.SetDefault("Web_Port", cfg.WebPort)
		options.SetDefault("Metrics_Port", cfg.MetricsPort)
		options.SetDefault("Metrics_Path", cfg.MetricsPath)

		options.SetDefault("Kafka_Brokers", clowder.KafkaServers)

		options.SetDefault("Log_Group", cfg.Logging.Cloudwatch.LogGroup)
		options.SetDefault("Aws_Region", cfg.Logging.Cloudwatch.Region)
		options.SetDefault("Aws_Access_Key_Id", cfg.Logging.Cloudwatch.AccessKeyId)
		options.SetDefault("Aws_Secret_Access_Key", cfg.Logging.Cloudwatch.SecretAccessKey)

		options.SetDefault("DB_Host", cfg.Database.Hostname)
		options.SetDefault("DB_Port", cfg.Database.Port)
		options.SetDefault("DB_Name", cfg.Database.Name)
		options.SetDefault("DB_User", cfg.Database.Username)
		options.SetDefault("DB_Pass", cfg.Database.Password)
	} else {
		options.SetDefault("Web_Port", 8081)
		options.SetDefault("Metrics_Port", 9000)
		options.SetDefault("Metrics_Path", "/metrics")

		options.SetDefault("Kafka_Brokers", []string{"localhost:29092"})

		options.SetDefault("Log_Group", "platform-dev")
		options.SetDefault("Aws_Region", "us-east-1")
		options.SetDefault("Aws_Access_Key_Id", os.Getenv("CW_AWS_ACCESS_KEY_ID"))
		options.SetDefault("Aws_Secret_Access_Key", os.Getenv("CW_AWS_SECRET_ACCESS_KEY"))

		options.SetDefault("DB_Host", "localhost")
		options.SetDefault("DB_Port", 5432)
		options.SetDefault("DB_Name", "insights")
		options.SetDefault("DB_User", "insights")
		options.SetDefault("DB_Pass", "insights")
	}

	options.SetDefault("Kafka_Group_ID", "config-manager")
	options.SetDefault("Kafka_Consumer_Offset", -1)
	options.SetDefault("Kafka_Dispatcher_Topic", "platform.playbook-dispatcher.runs")
	options.SetDefault("Kafka_Inventory_Topic", "platform.inventory.events")
	options.SetDefault("Kafka_System_Profile_Topic", "platform.inventory.system-profile")

	options.SetDefault("Dispatcher_Host", "http://playbook-dispatcher-api.playbook-dispatcher-ci.svc.cluster.local:8000")
	options.SetDefault("Dispatcher_PSK", "")
	options.SetDefault("Dispatcher_Batch_Size", 50)
	options.SetDefault("Dispatcher_Timeout", 10)
	options.SetDefault("Dispatcher_Impl", "impl")

	options.SetDefault("Playbook_Host", "https://ci.cloud.redhat.com")
	options.SetDefault("Playbook_Path", "/api/config-manager/v1/states/%s/playbook")

	options.SetDefault("Cloud_Connector_Host", "http://cloud-connector.connector-ci.svc.cluster.local:8080")
	options.SetDefault("Cloud_Connector_Client_ID", "config-manager")
	options.SetDefault("Cloud_Connector_PSK", "")
	options.SetDefault("Cloud_Connector_Timeout", 10)
	options.SetDefault("Cloud_Connector_Impl", "impl")

	options.SetDefault("Inventory_Host", "http://insights-inventory.platform-ci.svc.cluster.local:8080")
	options.SetDefault("Inventory_Timeout", 10)
	options.SetDefault("Inventory_Impl", "impl")

	options.SetDefault("Playbook_Files", "./playbooks/")

	options.SetDefault("Service_Config", `{
		"insights": "enabled",
		"compliance_openscap": "enabled",
		"remediations": "enabled"
	}`)

	options.SetEnvPrefix("CM")
	options.AutomaticEnv()

	return options
}

func buildURL(prefix, appName, version string) string {
	return fmt.Sprintf("/%s/%s/%s", prefix, appName, version)
}
