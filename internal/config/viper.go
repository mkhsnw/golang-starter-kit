package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func NewConfig() *Config {
	config := viper.New()
	config.SetConfigName("env")
	config.SetConfigType("json")
	config.AddConfigPath("./../")
	config.AddConfigPath("./")
	config.AutomaticEnv()
	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Failed to read config %w \n", err))
	}

	var appConfig Config
	err = config.Unmarshal(&appConfig)
	if err != nil {
		panic(fmt.Errorf("Failed to unmarshal config %w \n", err))
	}
	
	return &appConfig
}
