package database

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestProcessoLifecycle(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	p := &Processo{
		Numero:              "123456",
		StatusProcessamento: "PENDENTE",
		SeiUnidadeID:        "12124214",
		SeiUnidadeSigla:     "TESTE/TESTE",
	}
	err := store.SaveProcesso(t.Context(), p)
	if err != nil {
		t.Fatal(err)
	}

	p2, err := store.GetProcesso(t.Context(), p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(p, p2); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}

	p2, err = store.GetProcessoByNumero(t.Context(), p.Numero)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(p, p2); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}

	p.StatusProcessamento = "CONCLUIDO"
	p.Aposentadoria = sql.Null[bool]{
		V:     true,
		Valid: true,
	}
	p.AnalisadoEm = sql.Null[time.Time]{
		V:     time.Now(),
		Valid: true,
	}
	err = store.UpdateProcesso(t.Context(), p)
	if err != nil {
		t.Fatal(err)
	}

	p2, err = store.GetProcessoByNumero(t.Context(), p.Numero)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(p, p2); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}

	err = store.DeleteProcesso(t.Context(), p.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.GetProcesso(t.Context(), p.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}
