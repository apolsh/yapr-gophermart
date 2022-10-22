package config

import (
	"errors"
	"flag"

	"github.com/caarlos0/env"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8090"`
	//AuthSecretKey        string `env:"AUTH_SECRET_KEY" envDefault:"very_secret_key"`
	DatabaseType            string `env:"DATABASE_TYPE"  envDefault:"postgresql"`
	DatabaseURI             string `env:"DATABASE_URI"  envDefault:"postgresql://gophermartuser:gophermartpass@localhost:5432/gophermart"`
	TokenSecretKey          string `env:"TOKEN_SECRET_KEY"  envDefault:"very_secret_token_key"`
	LoyaltyServiceRateLimit int    `env:"LOYALTY_SERVICE_RATE_LIMIT" envDefault:"10"`
}

var availableDBTypes = map[string]bool{"postgresql": true}

func Load() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	var runAddressFlagValue string
	var accrualSystemAddressFlagValue string
	var databaseTypeFlagValue string
	var databaseURIFlagValue string
	var tokenSecretKeyFlagValue string
	var loyaltyServiceRateLimitFlagValue int

	flag.StringVar(&runAddressFlagValue, "a", "", "HTTP server address and port")
	flag.StringVar(&accrualSystemAddressFlagValue, "r", "", "address of the billing system")
	flag.StringVar(&databaseTypeFlagValue, "t", "", "database type, available: postgresql")
	flag.StringVar(&databaseURIFlagValue, "d", "", "database URI")
	flag.StringVar(&tokenSecretKeyFlagValue, "s", "", "token secret key")
	flag.IntVar(&loyaltyServiceRateLimitFlagValue, "l", -1, "loyalty service rate limit")

	flag.Parse()

	if runAddressFlagValue != "" {
		cfg.RunAddress = runAddressFlagValue
	}
	if accrualSystemAddressFlagValue != "" {
		cfg.AccrualSystemAddress = accrualSystemAddressFlagValue
	}
	if databaseURIFlagValue != "" {
		cfg.DatabaseURI = databaseURIFlagValue
	}
	if databaseTypeFlagValue != "" {
		cfg.DatabaseType = databaseTypeFlagValue
	}
	if tokenSecretKeyFlagValue != "" {
		cfg.TokenSecretKey = tokenSecretKeyFlagValue
	}
	if loyaltyServiceRateLimitFlagValue != -1 {
		cfg.LoyaltyServiceRateLimit = loyaltyServiceRateLimitFlagValue
	}

	if _, ok := availableDBTypes[cfg.DatabaseType]; !ok {
		return nil, errors.New("invalid database type")
	}

	return cfg, nil
}
