package config

import (
	"errors"
	"flag"

	"github.com/caarlos0/env"
)

const (
	PostgresStorageType = "postgresql"
)

type Config struct {
	RunAddress              string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	AccrualSystemAddress    string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8282"`
	DatabaseType            string `env:"DATABASE_TYPE"  envDefault:"postgresql"`
	DatabaseURI             string `env:"DATABASE_URI"  envDefault:"postgresql://gophermartuser:gophermartpass@localhost/gophermart?sslmode=disable"`
	TokenSecretKey          string `env:"TOKEN_SECRET_KEY"  envDefault:"very_secret_token_key"`
	LoyaltyServiceRateLimit int    `env:"LOYALTY_SERVICE_RATE_LIMIT" envDefault:"10"`
	LoyaltyServiceMaxTries  int    `env:"LOYALTY_SERVICE_MAX_TRIES" envDefault:"10"`
	LogLevel                string `env:"LOG_LEVEL" envDefault:"info"`
	HTTPSEnabled            bool   `env:"ENABLE_HTTPS" json:"enable_https"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	cmdLineCfg := readCmdLineArgs()

	cfg.overrideValues(cmdLineCfg)

	if _, ok := availableDBTypes[cfg.DatabaseType]; !ok {
		return nil, errors.New("invalid database type")
	}

	return cfg, nil
}

func readCmdLineArgs() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.RunAddress, "a", "", "HTTP server address and port")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "address of the billing system")
	flag.StringVar(&cfg.DatabaseType, "t", "", "database type, available: postgresql")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "database URI")
	flag.StringVar(&cfg.TokenSecretKey, "s", "", "token secret key")
	flag.IntVar(&cfg.LoyaltyServiceRateLimit, "l", -1, "loyalty service rate limit")

	flag.Parse()

	return cfg
}

func (c *Config) overrideValues(another *Config) {
	if another.RunAddress != "" {
		c.RunAddress = another.RunAddress
	}
	if another.AccrualSystemAddress != "" {
		c.AccrualSystemAddress = another.AccrualSystemAddress
	}
	if another.DatabaseURI != "" {
		c.DatabaseURI = another.DatabaseURI
	}
	if another.DatabaseType != "" {
		c.DatabaseType = another.DatabaseType
	}
	if another.TokenSecretKey != "" {
		c.TokenSecretKey = another.TokenSecretKey
	}
	if another.LoyaltyServiceRateLimit != -1 {
		c.LoyaltyServiceRateLimit = another.LoyaltyServiceRateLimit
	}
}

var availableDBTypes = map[string]bool{PostgresStorageType: true}
