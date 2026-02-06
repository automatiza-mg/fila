package logging

import (
	"context"
	"io"
	"log/slog"

	"github.com/lmittmann/tint"
)

type contextKey int

const (
	loggerContextKey contextKey = iota
)

// NewLogger cria uma nova instância de [slog.Logger].
func NewLogger(w io.Writer, dev bool) *slog.Logger {
	if dev {
		return slog.New(tint.NewHandler(w, &tint.Options{
			Level: slog.LevelDebug,
		}))
	}
	return slog.New(slog.NewJSONHandler(w, nil))
}

// WithLogger adiciona o logger como valor no contexto.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, logger)
}

// FromLogger retorna um [slog.Logger] de um contexto. Caso não seja
// encontrado, retorna [slog.Default].
func FromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerContextKey).(*slog.Logger)
	if !ok {
		return slog.Default()
	}
	return logger
}
