package database

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
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

func TestProcesso_NotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	_, err := store.GetProcesso(t.Context(), uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got: %v", err)
	}

	_, err = store.GetProcessoByNumero(t.Context(), "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got: %v", err)
	}
}

func TestProcesso_ListProcessos(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	for i := range 3 {
		p := &Processo{
			Numero:              fmt.Sprintf("list-proc-%03d", i+1),
			StatusProcessamento: "PENDENTE",
		}
		err := store.SaveProcesso(t.Context(), p)
		if err != nil {
			t.Fatal(err)
		}
	}

	// No filter returns all
	pp, count, err := store.ListProcessos(t.Context(), ListProcessosParams{})
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("expected count=3, got count=%d", count)
	}
	if len(pp) != 3 {
		t.Fatalf("expected len=3, got len=%d", len(pp))
	}

	// Filter by partial numero
	pp, count, err = store.ListProcessos(t.Context(), ListProcessosParams{Numero: "002"})
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected count=1, got count=%d", count)
	}
	if len(pp) != 1 {
		t.Fatalf("expected len=1, got len=%d", len(pp))
	}
	if pp[0].Numero != "list-proc-002" {
		t.Fatalf("expected numero=list-proc-002, got %s", pp[0].Numero)
	}

	// Non-matching filter
	pp, count, err = store.ListProcessos(t.Context(), ListProcessosParams{Numero: "nonexistent"})
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected count=0, got count=%d", count)
	}
	if len(pp) != 0 {
		t.Fatalf("expected len=0, got len=%d", len(pp))
	}
}

func TestProcesso_GetProcessosMap(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	p1 := &Processo{Numero: "map-proc-1", StatusProcessamento: "PENDENTE"}
	p2 := &Processo{Numero: "map-proc-2", StatusProcessamento: "PENDENTE"}
	err := store.SaveProcesso(t.Context(), p1)
	if err != nil {
		t.Fatal(err)
	}
	err = store.SaveProcesso(t.Context(), p2)
	if err != nil {
		t.Fatal(err)
	}

	processoMap, err := store.GetProcessosMap(t.Context(), []uuid.UUID{p1.ID, p2.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(processoMap) != 2 {
		t.Fatalf("expected 2 keys in map, got %d", len(processoMap))
	}
	if diff := cmp.Diff(p1, processoMap[p1.ID]); diff != "" {
		t.Fatalf("p1 mismatch:\n%s", diff)
	}
	if diff := cmp.Diff(p2, processoMap[p2.ID]); diff != "" {
		t.Fatalf("p2 mismatch:\n%s", diff)
	}

	// Empty input returns empty map
	processoMap, err = store.GetProcessosMap(t.Context(), []uuid.UUID{})
	if err != nil {
		t.Fatal(err)
	}
	if len(processoMap) != 0 {
		t.Fatalf("expected empty map, got %d keys", len(processoMap))
	}

	// Non-existent IDs return empty map
	processoMap, err = store.GetProcessosMap(t.Context(), []uuid.UUID{uuid.New()})
	if err != nil {
		t.Fatal(err)
	}
	if len(processoMap) != 0 {
		t.Fatalf("expected empty map, got %d keys", len(processoMap))
	}
}
