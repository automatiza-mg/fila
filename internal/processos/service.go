package processos

import (
	"context"
	"io"

	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type TextExtractor interface {
	ExtractText(ctx context.Context, r io.Reader, contentType string) (string, error)
}

type SeiClient interface {
	ConsultarProcedimento(ctx context.Context, protocolo string) (*sei.ConsultarProcedimentoResponse, error)
	ListarDocumentos(ctx context.Context, linkAcesso string) ([]sei.LinhaDocumento, error)
}

type TaskInserter interface {
	InsertTx(ctx context.Context, tx pgx.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error)
}

type Service struct {
	pool  *pgxpool.Pool
	store *database.Store
	sei   SeiClient
	cache cache.Cache
	queue TaskInserter
}

func New(pool *pgxpool.Pool, sei SeiClient, cache cache.Cache, queue TaskInserter) *Service {
	return &Service{
		pool:  pool,
		store: database.New(pool),
		sei:   sei,
		cache: cache,
		queue: queue,
	}
}
