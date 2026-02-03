package config

import (
	"github.com/automatiza-mg/fila/internal/mail"
	"github.com/automatiza-mg/fila/internal/postgres"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Mail     mail.Config
	Postgres postgres.Config
}

func NewFromEnv() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
