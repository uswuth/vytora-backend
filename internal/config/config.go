package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL    string
	JWTSecret      string
	JWTExpiryHours int
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	viper.ReadInConfig()

	return &Config{
		DatabaseURL:    viper.GetString("DATABASE_URL"),
		JWTSecret:      viper.GetString("JWT_SECRET"),
		JWTExpiryHours: viper.GetInt("JWT_EXPIRY_HOURS"),
	}
}
