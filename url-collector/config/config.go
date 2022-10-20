package config

import (
	"github.com/spf13/viper"
)

const (
	API_KEY             = "API_KEY"
	CONCURRENT_REQUESTS = "CONCURRENT_REQUESTS"
	PORT                = "PORT"

	DEFAULT_API_KEY            = "DEMO_KEY"
	DEFUALT_CONCURRENT_REQUEST = 5
	DEFAULT_PORT               = "8080"

	LOAD_WITH_VALUE = "%v load with value: %v"
)

var Config Configuration

type Configuration struct {
	ApiKey             string `mapstructure:"API_KEY"`
	ConcurrentRequests int    `mapstructure:"CONCURRENT_REQUESTS"`
	Port               string `mapstructure:"PORT"`
}

func LoadConfig() error {
	viper.SetDefault(API_KEY, DEFAULT_API_KEY)
	viper.SetDefault(CONCURRENT_REQUESTS, DEFUALT_CONCURRENT_REQUEST)
	viper.SetDefault(PORT, DEFAULT_PORT)
	viper.AutomaticEnv()
	err := viper.Unmarshal(&Config)
	return err
}
