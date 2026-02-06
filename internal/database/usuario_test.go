package database

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type seedUsuarioOpt func(u *Usuario)

func withPapel(papel string) seedUsuarioOpt {
	return func(u *Usuario) {
		u.Papel = sql.Null[string]{
			V:     papel,
			Valid: true,
		}
	}
}

// Cria um novo usuário com dados aleatórios para fins de teste.
func seedUsuario(t *testing.T, store *Store, opts ...seedUsuarioOpt) *Usuario {
	t.Helper()

	usuario := &Usuario{
		CPF:   rand.Text(),
		Email: rand.Text(),
	}

	for _, opt := range opts {
		opt(usuario)
	}

	err := store.SaveUsuario(t.Context(), usuario)
	if err != nil {
		t.Fatal(err)
	}

	return usuario
}

func TestUsuario_Lifecycle(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	usuario := &Usuario{
		Nome:            "Fulano da Silva",
		CPF:             "00000000000",
		Email:           "fulano@email.com",
		EmailVerificado: true,
	}
	err := store.SaveUsuario(t.Context(), usuario)
	if err != nil {
		t.Fatal(err)
	}

	empty, err := store.IsUsuariosEmpty(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if empty {
		t.Fatal("expected usuarios table to not be empty")
	}

	read, err := store.GetUsuario(t.Context(), usuario.ID)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(usuario, read); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}

	read, err = store.GetUsuarioByCPF(t.Context(), usuario.CPF)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(usuario, read); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}
}

func TestUsuario_EmailTaken(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	u1 := &Usuario{Email: "fulano@email.com", CPF: "0000000000"}
	err := store.SaveUsuario(t.Context(), u1)
	if err != nil {
		t.Fatal(err)
	}

	u2 := &Usuario{Email: "fulano@email.com", CPF: "0000000001"}
	err = store.SaveUsuario(t.Context(), u2)
	if !errors.Is(err, ErrUsuarioEmailTaken) {
		t.Fatalf("want ErrUsuarioEmailTaken, got: %v", err)
	}
}

func TestUsuario_CPFTaken(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	u1 := &Usuario{Email: "fulano@email.com", CPF: "0000000000"}
	err := store.SaveUsuario(t.Context(), u1)
	if err != nil {
		t.Fatal(err)
	}

	u2 := &Usuario{Email: "fulano1@email.com", CPF: "0000000000"}
	err = store.SaveUsuario(t.Context(), u2)
	if !errors.Is(err, ErrUsuarioCPFTaken) {
		t.Fatalf("want ErrUsuarioCPFTaken, got: %v", err)
	}
}

func TestUsuario_NotFound(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	_, err := store.GetUsuario(t.Context(), 1)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got: %v", err)
	}

	_, err = store.GetUsuarioByCPF(t.Context(), "")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got: %v", err)
	}
}

func TestUsuario_Empty(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	empty, err := store.IsUsuariosEmpty(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if !empty {
		t.Fatal("want: true, got: false")
	}
}

func TestUsuario_List(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	seed := map[string]int{
		PapelAdmin:         1,
		PapelAnalista:      4,
		PapelGestor:        2,
		PapelSubsecretario: 3,
	}

	total := 0

	var wg sync.WaitGroup
	for papel, qty := range seed {
		total += qty
		for range qty {
			wg.Go(func() {
				seedUsuario(t, store, withPapel(papel))
			})
		}
	}
	wg.Wait()

	_, count, err := store.ListUsuarios(t.Context(), ListUsuariosParams{})
	if err != nil {
		t.Fatal(err)
	}
	if count != total {
		t.Fatalf("expected len(usuario) to be %d", total)
	}

	for papel, qty := range seed {
		_, count, err := store.ListUsuarios(t.Context(), ListUsuariosParams{
			Papel: papel,
		})
		if err != nil {
			t.Fatal(err)
		}
		if count != qty {
			t.Fatalf("expected %d usuarios with papel %s, got: %d", qty, papel, count)
		}
	}
}
