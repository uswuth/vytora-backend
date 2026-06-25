package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL       string
	JWTSecret         string
	JWTExpiryHours    int
	AllowedOrigins    []string
	HealthAllowedIPs  []string
	RateLimitRequests int
	RateLimitInterval time.Duration
	MaxBodySize       int64
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

	rateLimitRequests := viper.GetInt("RATE_LIMIT_REQUESTS")
	if rateLimitRequests == 0 {
		rateLimitRequests = 100
	}

	rateLimitIntervalStr := viper.GetString("RATE_LIMIT_INTERVAL")
	rateLimitInterval, err := time.ParseDuration(rateLimitIntervalStr)
	if err != nil || rateLimitInterval <= 0 {
		rateLimitInterval = time.Minute
	}

	maxBodySize := viper.GetInt64("MAX_BODY_SIZE")
	if maxBodySize == 0 {
		maxBodySize = 1048576 // 1MB default
	}

	fmt.Println("Config loaded successfully from .env")
	return &Config{
		DatabaseURL:       databaseURL,
		JWTSecret:         jwtSecret,
		JWTExpiryHours:    jwtExpiryHours,
		AllowedOrigins:    allowedOrigins,
		HealthAllowedIPs:  healthAllowedIPs,
		RateLimitRequests: rateLimitRequests,
		RateLimitInterval: rateLimitInterval,
		MaxBodySize:       maxBodySize,
	}
}