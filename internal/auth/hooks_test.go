package auth

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
)

var _ UsuarioHook = (*counterHook)(nil)

type counterHook struct {
	mu       sync.Mutex
	actions  int
	cleanups int
}

func (d *counterHook) Label() string {
	return "counter"
}

func (d *counterHook) GetActions(ctx context.Context, u *Usuario) ([]PendingAction, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.actions++
	return nil, nil
}

func (d *counterHook) Cleanup(ctx context.Context, tx pgx.Tx, trigger CleanupTrigger, usuario *Usuario) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.cleanups++
	return nil
}
