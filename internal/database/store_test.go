package database

import (
	"testing"

	"github.com/automatiza-mg/fila/internal/postgres"
)

var ti *postgres.TestInstance

// Cria uma nova instância de [Store] usando a instância de teste do Postgres.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	pool := ti.NewDatabase(t)
	return &Store{db: pool}
}

func TestMain(m *testing.M) {
	ti = postgres.MustTestInstance()
	defer ti.Close()

	m.Run()
}
