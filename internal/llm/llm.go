package llm

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Client struct {
	openai openai.Client
}

func New(cfg *Config) *Client {
	return &Client{
		openai: openai.NewClient(
			option.WithBaseURL(cfg.AzureURL),
			option.WithAPIKey(cfg.AzureApiKey),
		),
	}
}
