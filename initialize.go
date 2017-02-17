package main

import (
	"encoding/json"
	"log"

	env "github.com/joeshaw/envdecode"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pkg/errors"
)

var (
	cfg Config // global configuration
)

// Config contains the configuration from environment variables
type Config struct {
	Debug bool   `env:"DEBUG,default=true"`
	Port  string `env:"PORT,default=8000"`

	SQL struct {
		// Host     string `env:"MSSQL_HOST,default=localhost"`
		// Port     string `env:"MSSQL_PORT,default=1433"`
		User     string `env:"SQL_USER,default=postgres"`
		Password string `env:"SQL_PASSWORD,default=mysecretpassword"`
		Database string `env:"SQL_DATABASE,default=products"`
	}
}

// initialize our configuration from environment variables.
func initialize() error {

	// For development, github.com/joho/godotenv/autoload
	// loads env variables from .env file for you.

	// Read configuration from env variables
	err := env.Decode(&cfg)
	if err != nil {
		return errors.Wrap(err, "configuration decode failed")
	}

	// log configuration for debugging
	if cfg.Debug {
		prettyCfg, _ := json.MarshalIndent(cfg, "", "  ")
		log.Printf("Configuration: \n%v", string(prettyCfg))
	}

	return nil
}
