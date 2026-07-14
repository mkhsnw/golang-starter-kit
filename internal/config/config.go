package config

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
	JWT      JwtConfig      `mapstructure:"jwt"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
	Port        int    `mapstructure:"port"`
	Url         string `mapstructure:"url"`
}

type DatabasePoolConfig struct {
	MaxIdle     int    `mapstructure:"maxIdle"`
	MaxOpen     int    `mapstructure:"maxOpen"`
	MaxLifetime string `mapstructure:"maxLifetime"`
}

type DatabaseConfig struct {
	Port     int                `mapstructure:"port"`
	Username string             `mapstructure:"username"`
	Password string             `mapstructure:"password"`
	Host     string             `mapstructure:"host"`
	Name     string             `mapstructure:"name"`
	Pool     DatabasePoolConfig `mapstructure:"pool"`
}

type LogConfig struct {
	Level int `mapstructure:"level"`
}

type JwtConfig struct {
	Secret          string `mapstructure:"secret"`
	ExpirationHours int    `mapstructure:"expiration_hours"`
}
