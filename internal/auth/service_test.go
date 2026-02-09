package auth

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/automatiza-mg/fila/internal/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

var ti *postgres.TestInstance

var _ TaskInserter = (*taskInserter)(nil)

type taskInserter struct {
	id atomic.Int64

	mu   sync.Mutex
	args []river.JobArgs
}

func (t *taskInserter) InsertTx(ctx context.Context, tx pgx.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	t.mu.Lock()
	t.args = append(t.args, args)
	t.mu.Unlock()

	id := t.id.Add(1)
	return &rivertype.JobInsertResult{
		Job: &rivertype.JobRow{
			ID: id,
		},
	}, nil
}

func (t *taskInserter) Args() []river.JobArgs {
	t.mu.Lock()
	defer t.mu.Unlock()

	res := make([]river.JobArgs, len(t.args))
	copy(res, t.args)
	return res
}

func newTestService(tb testing.TB) *Service {
	tb.Helper()

	pool := ti.NewDatabase(tb)
	return New(pool, slog.New(slog.DiscardHandler), &taskInserter{})
}

func TestMain(m *testing.M) {
	ti = postgres.MustTestInstance()
	defer ti.Close()

	code := m.Run()
	os.Exit(code)
}
