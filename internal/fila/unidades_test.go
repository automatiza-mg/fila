package fila

import (
	"strings"
	"testing"

	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
)

func TestListarUnidadesAnalista(t *testing.T) {
	t.Parallel()

	pool := ti.NewDatabase(t)
	fila := &Service{
		pool:       pool,
		store:      database.New(pool),
		sei:        &fakeSei{},
		cache:      cache.NewMemoryCache(),
		analyzer:   &fakeAnalyzer{},
		servidores: &fakeServidores{},
	}

	unidades, err := fila.ListUnidadesAnalistas(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	for _, unidade := range unidades {
		if !strings.HasPrefix(unidade.Sigla, "SEPLAG/AP") {
			t.Fatalf("expected %q to have SEPLAG/AP prefix", unidade.Sigla)
		}
	}
}

func TestGetUnidadesMap(t *testing.T) {
	t.Parallel()

	pool := ti.NewDatabase(t)
	fila := &Service{
		pool:     pool,
		store:    database.New(pool),
		sei:      &fakeSei{},
		cache:    cache.NewMemoryCache(),
		analyzer: &fakeAnalyzer{},
	}

	m, err := fila.GetUnidadesMap(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for id, unidade := range m {
		if id != unidade.ID {
			t.Fatalf("key/id mismatch %s = %s", id, unidade.ID)
		}
	}
}
