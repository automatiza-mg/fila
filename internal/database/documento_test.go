package database

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDocumentoLifecycle(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	p := &Processo{
		Numero: "123123",
	}
	err := store.SaveProcesso(t.Context(), p)
	if err != nil {
		t.Fatal(err)
	}

	d := &Documento{
		Numero:       "123123",
		ProcessoID:   p.ID,
		MetadadosAPI: []byte("{}"),
	}
	err = store.SaveDocumento(t.Context(), d)
	if err != nil {
		t.Fatal(err)
	}

	d2, err := store.GetDocumento(t.Context(), d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(d, d2); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}

	d2, err = store.GetDocumentoByNumero(t.Context(), d.Numero)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(d, d2); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}
}
