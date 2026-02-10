package database

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHistoricoStatusProcessoLifecycle(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	p := &Processo{
		Numero: "12512312",
	}

	err := store.SaveProcesso(t.Context(), p)
	if err != nil {
		t.Fatal(err)
	}

	pa := &ProcessoAposentadoria{
		ProcessoID: p.ID,
		Status:     StatusProcessoAnalisePendente,
	}
	err = store.SaveProcessoAposentadoria(t.Context(), pa)
	if err != nil {
		t.Fatal(err)
	}

	h := &HistoricoStatusProcesso{
		ProcessoAposentadoriaID: pa.ID,
		StatusNovo:              StatusProcessoEmAnalise,
	}
	err = store.SaveHistoricoStatusProcesso(t.Context(), h)
	if err != nil {
		t.Fatal(err)
	}

	h2, err := store.GetHistoricoStatusProcesso(t.Context(), h.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(h, h2); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}

	hh, err := store.ListHistoricoStatusProcesso(t.Context(), pa.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(hh) != 1 {
		t.Fatalf("expected len=1, got len=%d", len(hh))
	}
	if diff := cmp.Diff(h, hh[0]); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}
}
