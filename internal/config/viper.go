package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func NewConfig() *Config {
	config := viper.New()
	config.SetConfigName("env")
	config.SetConfigType("json")
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

	if appConfig.JWT.Secret == "your-secret-key-here" || len(appConfig.JWT.Secret) < 32 {
		panic(fmt.Errorf("FATAL SECURITY ERROR: JWT Secret in env.json must be changed and be at least 32 characters long!"))
	}

	return &appConfig
}
