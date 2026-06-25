package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// LoadWithDefaults returns *Config or a descriptive error instead of panic.
func LoadWithDefaults() (*Config, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		// Some environments may rely purely on AutomaticEnv.
		fmt.Fprintln(os.Stderr, "config: .env not loaded, using environment variables")
	}

	port := v.GetString("PORT")
	if port == "" {
		port = "8080"
	}
	host := v.GetString("HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	appEnv := v.GetString("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	connectionName := v.GetString("DB_CONNECTION_NAME")
	if connectionName == "" {
		connectionName = "VRMP Dev"
	}
	dbName := v.GetString("DB_NAME")
	dbHost := v.GetString("DB_HOST")
	dbPort := v.GetString("DB_PORT")
	dbUser := v.GetString("DB_USER")
	dbPassword := v.GetString("DB_PASSWORD")

	if dbHost == "" || dbPort == "" || dbUser == "" || dbName == "" {
		if u, err := url.Parse(v.GetString("DATABASE_URL")); err == nil {
			if dbHost == "" {
				dbHost = u.Hostname()
			}
			if dbPort == "" && u.Port() != "" {
				dbPort = u.Port()
			}
			if dbUser == "" {
				dbUser = u.User.Username()
			}
			if dbName == "" {
				dbName = strings.TrimPrefix(u.Path, "/")
			}
			if p, ok := u.User.Password(); ok && dbPassword == "" {
				dbPassword = p
			}
		}
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbUser == "" {
		dbUser = "postgres"
	}
	if dbName == "" {
		dbName = "vrmp_dev"
	}

	cfg := &Config{
		DatabaseURL:       v.GetString("DATABASE_URL"),
		JWTSecret:         v.GetString("JWT_SECRET"),
		JWTExpiryHours:    v.GetInt("JWT_EXPIRY_HOURS"),
		AllowedOrigins:    v.GetStringSlice("ALLOWED_ORIGINS"),
		HealthAllowedIPs:  v.GetStringSlice("HEALTH_ALLOWED_IPS"),
		RateLimitRequests: v.GetInt("RATE_LIMIT_REQUESTS"),
		RateLimitInterval: time.Duration(v.GetDuration("RATE_LIMIT_INTERVAL")),
		MaxBodySize:       v.GetInt64("MAX_BODY_SIZE"),
		Host:              host,
		Port:              port,
		AppEnv:            appEnv,
		ConnectionName:    connectionName,
		DBHost:            dbHost,
		DBPort:            dbPort,
		DBUser:            dbUser,
		DBPassword:        dbPassword,
		DBName:            dbName,
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is not set")
	}
	if cfg.JWTExpiryHours == 0 {
		return nil, fmt.Errorf("JWT_EXPIRY_HOURS is not set")
	}
	if len(cfg.AllowedOrigins) == 0 {
		return nil, fmt.Errorf("ALLOWED_ORIGINS is not set")
	}
	if len(cfg.HealthAllowedIPs) == 0 {
		return nil, fmt.Errorf("HEALTH_ALLOWED_IPS is not set")
	}
	if cfg.RateLimitRequests == 0 {
		cfg.RateLimitRequests = 100
	}
	if cfg.RateLimitInterval <= 0 {
		cfg.RateLimitInterval = time.Duration(v.GetDuration("RATE_LIMIT_INTERVAL"))
		if cfg.RateLimitInterval <= 0 {
			cfg.RateLimitInterval = time.Minute
		}
	}
	if cfg.MaxBodySize == 0 {
		cfg.MaxBodySize = 1048576
	}

	return cfg, nil
}
