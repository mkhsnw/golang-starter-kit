package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func NewConfig() *Config {
	config := viper.New()
	config.SetConfigName("env")
	config.SetConfigType("json")
	config.AddConfigPath("./")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AutomaticEnv()

	err := config.ReadInConfig()
	if err != nil {
		fmt.Printf("❌ FAIL-FAST ERROR: Failed to read env.json configuration file: %v\n", err)
		os.Exit(1)
	}

	var appConfig Config
	err = config.Unmarshal(&appConfig)
	if err != nil {
		fmt.Printf("❌ FAIL-FAST ERROR: Failed to parse configuration structure: %v\n", err)
		os.Exit(1)
	}

	if err := appConfig.Validate(); err != nil {
		fmt.Printf("❌ FAIL-FAST ERROR: Configuration validation failed:\n%v\n", err)
		os.Exit(1)
	}

	if appConfig.JWT.Secret == "your-secret-key-here" || len(appConfig.JWT.Secret) < 16 {
		fmt.Println("❌ FATAL SECURITY ERROR: JWT Secret in env.json must be changed and be at least 16 characters long!")
		os.Exit(1)
	}

	return &appConfig
}
