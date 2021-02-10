package config

import (
	"github.com/spf13/viper"
)

func Get() *viper.Viper {
	options := viper.New()

	options.SetDefault("ApiSpecFile", "api/openapi.json")

	options.SetDefault("DBUser", "insights")
	options.SetDefault("DBPass", "insights")
	options.SetDefault("DBName", "config-manager")
	options.SetDefault("DBHost", "localhost")

	options.SetDefault("KafkaGroupID", "config-manager")
	options.SetDefault("KafkaBrokers", []string{"localhost:9092"})
	options.SetDefault("KafkaConsumerOffset", -1)
	options.SetDefault("KafkaResultsTopic", "platform.playbook-dispatcher.results")
	options.SetDefault("KafkaConnectionsTopic", "platform.inventory.connections")

	options.SetEnvPrefix("CM")
	options.AutomaticEnv()

	return options
}
