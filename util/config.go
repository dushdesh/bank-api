package util

import "github.com/spf13/viper"

// Stores all the configuration of the application
// Values are read by Viper from files or environment variables
type Config struct {
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

// Reads the configuration from file or environment
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env") // Can use json, xml or something else

	viper.AutomaticEnv() // read environment varibales and overwrite the variables read from the config file

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	viper.Unmarshal(&config)
	return
}
