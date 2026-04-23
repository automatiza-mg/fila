package llm

import (
	"log/slog"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

type Client struct {
	openai openai.Client
	logger *slog.Logger
}

func New(cfg *Config, logger *slog.Logger) *Client {
	return &Client{
		openai: openai.NewClient(
			option.WithBaseURL(cfg.AzureURL),
			option.WithAPIKey(cfg.AzureApiKey),
		),
		logger: logger.With(slog.String("service", "llm")),
	}
}

// logUsage registra métricas de uso de uma chamada à API de Responses no
// nível Info, anexando os campos padrão (tarefa, modelo, tokens, latência).
func (c *Client) logUsage(msg, tarefa string, resp *responses.Response, latencia time.Duration) {
	c.logger.Info(msg,
		slog.String("tarefa", tarefa),
		slog.String("response_id", resp.ID),
		slog.String("modelo", resp.Model),
		slog.String("status", string(resp.Status)),
		slog.Int64("input_tokens", resp.Usage.InputTokens),
		slog.Int64("output_tokens", resp.Usage.OutputTokens),
		slog.Int64("total_tokens", resp.Usage.TotalTokens),
		slog.Int64("cached_input_tokens", resp.Usage.InputTokensDetails.CachedTokens),
		slog.Int64("reasoning_tokens", resp.Usage.OutputTokensDetails.ReasoningTokens),
		slog.Duration("latencia", latencia),
	)
}
