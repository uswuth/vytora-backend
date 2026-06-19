package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL      string
	JWTSecret        string
	JWTExpiryHours   int
	AllowedOrigins   []string
	HealthAllowedIPs []string
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	viper.ReadInConfig()

	databaseURL := viper.GetString("DATABASE_URL")
	if databaseURL == "" {
		panic("DATABASE_URL is not set in .env")
	}

	jwtSecret := viper.GetString("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET is not set in .env")
	}

	jwtExpiryHours := viper.GetInt("JWT_EXPIRY_HOURS")
	if jwtExpiryHours == 0 {
		panic("JWT_EXPIRY_HOURS is not set in .env")
	}

	allowedOrigins := viper.GetStringSlice("ALLOWED_ORIGINS")
	if len(allowedOrigins) == 0 {
		panic("ALLOWED_ORIGINS is not set in .env")
	}

	healthAllowedIPs := viper.GetStringSlice("HEALTH_ALLOWED_IPS")
	if len(healthAllowedIPs) == 0 {
		panic("HEALTH_ALLOWED_IPS is not set in .env")
	}

	fmt.Println("Config loaded successfully from .env")
	return &Config{
		DatabaseURL:      databaseURL,
		JWTSecret:        jwtSecret,
		JWTExpiryHours:   jwtExpiryHours,
		AllowedOrigins:   allowedOrigins,
		HealthAllowedIPs: healthAllowedIPs,
	}
}