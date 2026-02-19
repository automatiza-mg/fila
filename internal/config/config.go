package config

import (
	"net/url"

	"github.com/automatiza-mg/fila/internal/blob"
	"github.com/automatiza-mg/fila/internal/datalake"
	"github.com/automatiza-mg/fila/internal/docintel"
	"github.com/automatiza-mg/fila/internal/llm"
	"github.com/automatiza-mg/fila/internal/mail"
	"github.com/automatiza-mg/fila/internal/postgres"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	BaseURL   string   `env:"BASE_URL,notEmpty"`
	ClientURL *url.URL `env:"CLIENT_URL,notEmpty"`
	RedisURL  string   `env:"REDIS_URL,notEmpty" envDefault:"redis://localhost:6379"`

	Mail     mail.Config
	Postgres postgres.Config
	SEI      sei.Config
	DataLake datalake.Config
	Blob     blob.Config
	DocIntel docintel.Config
	LLM      llm.Config
}

func NewFromEnv() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
