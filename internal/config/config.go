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

	jwtSecret := viper.GetString("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "vytora-backend-secret-change-me-in-production"
	}

	jwtExpiryHours := viper.GetInt("JWT_EXPIRY_HOURS")
	if jwtExpiryHours == 0 {
		jwtExpiryHours = 24
	}

	return &Config{
		DatabaseURL:    viper.GetString("DATABASE_URL"),
		JWTSecret:      jwtSecret,
		JWTExpiryHours: jwtExpiryHours,
	}
}
