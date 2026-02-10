package database

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestProcessoAposentadoriaLifecycle(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	p := &Processo{
		Numero: "123123",
	}
	err := store.SaveProcesso(t.Context(), p)
	if err != nil {
		t.Fatal(err)
	}

	pa := &ProcessoAposentadoria{
		ProcessoID:               p.ID,
		DataRequerimento:         time.Now(),
		Status:                   StatusProcessoAnalisePendente,
		CPFRequerente:            "123.456.789-09",
		DataNascimentoRequerente: time.Date(1950, time.April, 2, 0, 0, 0, 0, time.Local),
	}
	err = store.SaveProcessoAposentadoria(t.Context(), pa)
	if err != nil {
		t.Fatal(err)
	}

	pa2, err := store.GetProcessoAposentadoria(t.Context(), pa.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(pa, pa2); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}

	pa2, err = store.GetProcessoAposentadoriaByNumero(t.Context(), p.Numero)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(pa, pa2); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}

	pa.Status = StatusProcessoEmAnalise
	pa.Score = 10
	err = store.UpdateProcessoAposentadoria(t.Context(), pa)
	if err != nil {
		t.Fatal(err)
	}

	pa2, err = store.GetProcessoAposentadoria(t.Context(), pa.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(pa, pa2); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}
}
