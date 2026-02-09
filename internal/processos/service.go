package processos

import (
	"context"
	"io"

	"github.com/automatiza-mg/fila/internal/blob"
	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TextExtractor interface {
	// TODO: Implementar ExtractTextFromURL.
	ExtractText(ctx context.Context, r io.Reader, contentType string) (string, error)
}

type Service struct {
	pool    *pgxpool.Pool
	store   *database.Store
	sei     *sei.Client
	cache   cache.Cache
	storage blob.Storage
	ocr     TextExtractor
}

type ServiceOpts struct {
	Pool    *pgxpool.Pool
	Sei     *sei.Client
	Cache   cache.Cache
	Storage blob.Storage
	OCR     TextExtractor
}

func New(opts *ServiceOpts) *Service {
	return &Service{
		pool:    opts.Pool,
		store:   database.New(opts.Pool),
		sei:     opts.Sei,
		cache:   opts.Cache,
		storage: opts.Storage,
		ocr:     opts.OCR,
	}
}
