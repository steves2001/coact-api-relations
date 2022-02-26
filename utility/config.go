package utility

import (
	"github.com/spf13/viper"
)

type Config struct {
	Neo4jUri      string `mapstructure:"NEO4J_URI"`
	Neo4jUser     string `mapstructure:"NEO4J_USER"`
	Neo4jPassword string `mapstructure:"NEO4J_PASSWORD"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
