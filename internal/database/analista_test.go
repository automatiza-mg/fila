package database

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestAnalistaLifecycle(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	usuario := seedUsuario(t, store)

	analista := &Analista{
		UsuarioID:    usuario.ID,
		Orgao:        "SEPLAG",
		SEIUnidadeID: "123123",
	}
	err := store.SaveAnalista(t.Context(), analista)
	if err != nil {
		t.Fatal(err)
	}

	read, err := store.GetAnalista(t.Context(), usuario.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(analista, read); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}

	analista.Afastado = true
	analista.UltimaAtribuicaoEm = sql.Null[time.Time]{
		V:     time.Now(),
		Valid: true,
	}

	err = store.UpdateAnalista(t.Context(), analista)
	if err != nil {
		t.Fatal(err)
	}

	read, err = store.GetAnalista(t.Context(), usuario.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(analista, read); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}

	err = store.DeleteAnalista(t.Context(), usuario.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.GetAnalista(t.Context(), usuario.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want: ErrNotFound, got: %v", err)
	}
}
