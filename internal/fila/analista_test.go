package fila

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/automatiza-mg/fila/internal/auth"
	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/google/go-cmp/cmp"
)

func seedUsuario(t *testing.T, store *database.Store, papel string) int64 {
	t.Helper()

	u := &database.Usuario{
		Papel: sql.Null[string]{
			V:     papel,
			Valid: true,
		},
	}
	err := store.SaveUsuario(t.Context(), u)
	if err != nil {
		t.Fatal(err)
	}

	return u.ID
}

func TestAnalistaLifecycle(t *testing.T) {
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

	usuarioID := seedUsuario(t, fila.store, auth.PapelAnalista)

	analista, err := fila.CreateAnalista(t.Context(), CreateAnalistaParams{
		UsuarioID:    usuarioID,
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

	pool := ti.NewDatabase(t)
	fila := &Service{
		pool:       pool,
		store:      database.New(pool),
		sei:        &fakeSei{},
		cache:      cache.NewMemoryCache(),
		analyzer:   &fakeAnalyzer{},
		servidores: &fakeServidores{},
	}

	usuarioID := seedUsuario(t, fila.store, auth.PapelAnalista)

	_, err := fila.CreateAnalista(t.Context(), CreateAnalistaParams{
		UsuarioID:    usuarioID,
		SeiUnidadeID: "-1", // Unidade inválida.
		Orgao:        "SEPLAG",
	})
	if !errors.Is(err, ErrInvalidUnidade) {
		t.Fatalf("expected ErrInvalidUnidade, got: %v", err)
	}
}

func TestAnalista_InvalidPapel(t *testing.T) {
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

	usuarioID := seedUsuario(t, fila.store, auth.PapelGestor)

	_, err := fila.CreateAnalista(t.Context(), CreateAnalistaParams{
		UsuarioID:    usuarioID,
		SeiUnidadeID: "1",
		Orgao:        "SEPLAG",
	})
	if !errors.Is(err, ErrInvalidPapel) {
		t.Fatalf("expected ErrInvalidPapel, got: %v", err)
	}
}
