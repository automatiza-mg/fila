package config

import (
	"github.com/automatiza-mg/fila/internal/mail"
	"github.com/automatiza-mg/fila/internal/postgres"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	BaseURL  string `env:"BASE_URL,notEmpty"`
	RedisURL string `env:"REDIS_URL,notEmpty" envDefault:"redis://localhost:6379"`
	Mail     mail.Config
	Postgres postgres.Config
	SEI      sei.Config
}

func NewFromEnv() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
