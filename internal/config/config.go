package config

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	TgToken string `envconfig:"TG_TOKEN" required:"true"`

	DBHost string `envconfig:"DB_HOST" default:"db"`
	DBPort string `envconfig:"DB_PORT" default:"5432"`
	DBUser string `envconfig:"DB_USER" default:"postgres"`
	DBPass string `envconfig:"DB_PASS" default:"postgres"`
	DBName string `envconfig:"DB_NAME" default:"expensetracker"`

	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"15s"`
}

func Load() (Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	return cfg, err
}

func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	return cfg
}
