package fila

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAnalistaLifecycle(t *testing.T) {
	t.Parallel()

	fila := newTestService(t)

	analista, err := fila.CreateAnalista(t.Context(), CreateAnalistaParams{
		UsuarioID:    1,
		SeiUnidadeID: "1",
		Orgao:        "SEPLAG",
	})
	if err != nil {
		t.Fatal(err)
	}

	read, err := fila.GetAnalista(t.Context(), analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(analista, read); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}

	err = fila.AfastarAnalista(t.Context(), analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}

	read, err = fila.GetAnalista(t.Context(), analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}
	if !read.Afastado {
		t.Fatal("expected Afatastado to be true")
	}

	err = fila.RetornarAnalista(t.Context(), analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}

	read, err = fila.GetAnalista(t.Context(), analista.UsuarioID)
	if err != nil {
		t.Fatal(err)
	}
	if read.Afastado {
		t.Fatal("expected Afatastado to be false")
	}
}

func TestAnalista_InvalidUnidade(t *testing.T) {
	t.Parallel()

	fila := newTestService(t)

	_, err := fila.CreateAnalista(t.Context(), CreateAnalistaParams{
		UsuarioID:    1,
		SeiUnidadeID: "-1", // Unidade inválida.
		Orgao:        "SEPLAG",
	})
	if !errors.Is(err, ErrInvalidUnidade) {
		t.Fatalf("expected ErrInvalidUnidade, got: %v", err)
	}
}
