package database

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func seedAnalista(t *testing.T, store *Store) (*Usuario, *Analista) {
	t.Helper()

	usuario := seedUsuario(t, store)
	analista := &Analista{
		UsuarioID:       usuario.ID,
		Orgao:           "SEPLAG",
		SEIUnidadeID:    rand.Text(),
		SEIUnidadeSigla: "SEPLAG/AP00",
	}
	err := store.SaveAnalista(t.Context(), analista)
	if err != nil {
		t.Fatal(err)
	}
	return usuario, analista
}

func TestAnalistaLifecycle(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	usuario := seedUsuario(t, store)

	analista := &Analista{
		UsuarioID:       usuario.ID,
		Orgao:           "SEPLAG",
		SEIUnidadeID:    "123123",
		SEIUnidadeSigla: "SEPLAG/AP00",
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

func TestAnalista_NotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	_, err := store.GetAnalista(t.Context(), 999999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got: %v", err)
	}
}

func TestAnalista_Duplicate(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	usuario, _ := seedAnalista(t, store)

	dup := &Analista{
		UsuarioID:       usuario.ID,
		Orgao:           "OTHER",
		SEIUnidadeID:    "999",
		SEIUnidadeSigla: "OTHER/XX00",
	}
	err := store.SaveAnalista(t.Context(), dup)
	if !errors.Is(err, ErrAnalistaExists) {
		t.Fatalf("want ErrAnalistaExists, got: %v", err)
	}
}

func TestAnalista_ListAnalistas(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	analistas, err := store.ListAnalistas(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(analistas) != 0 {
		t.Fatalf("expected empty list, got %d", len(analistas))
	}

	_, a1 := seedAnalista(t, store)
	_, a2 := seedAnalista(t, store)
	_, a3 := seedAnalista(t, store)

	analistas, err = store.ListAnalistas(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(analistas) != 3 {
		t.Fatalf("expected 3 analistas, got %d", len(analistas))
	}

	got := make(map[int64]*Analista)
	for _, a := range analistas {
		got[a.UsuarioID] = a
	}
	for _, want := range []*Analista{a1, a2, a3} {
		g, ok := got[want.UsuarioID]
		if !ok {
			t.Fatalf("analista with usuario_id %d not found in list", want.UsuarioID)
		}
		if diff := cmp.Diff(want, g); diff != "" {
			t.Fatalf("mismatch for usuario_id %d:\n%s", want.UsuarioID, diff)
		}
	}
}

func TestAnalista_GetAnalistasMap(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	_, a1 := seedAnalista(t, store)
	_, a2 := seedAnalista(t, store)

	analistaMap, err := store.GetAnalistasMap(t.Context(), []int64{a1.UsuarioID, a2.UsuarioID})
	if err != nil {
		t.Fatal(err)
	}
	if len(analistaMap) != 2 {
		t.Fatalf("expected 2 keys in map, got %d", len(analistaMap))
	}
	if diff := cmp.Diff(a1, analistaMap[a1.UsuarioID]); diff != "" {
		t.Fatalf("a1 mismatch:\n%s", diff)
	}
	if diff := cmp.Diff(a2, analistaMap[a2.UsuarioID]); diff != "" {
		t.Fatalf("a2 mismatch:\n%s", diff)
	}

	analistaMap, err = store.GetAnalistasMap(t.Context(), []int64{})
	if err != nil {
		t.Fatal(err)
	}
	if len(analistaMap) != 0 {
		t.Fatalf("expected empty map, got %d keys", len(analistaMap))
	}

	analistaMap, err = store.GetAnalistasMap(t.Context(), []int64{999998, 999999})
	if err != nil {
		t.Fatal(err)
	}
	if len(analistaMap) != 0 {
		t.Fatalf("expected empty map, got %d keys", len(analistaMap))
	}
}
