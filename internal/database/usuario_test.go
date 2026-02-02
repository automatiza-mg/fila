package database

import (
	"crypto/rand"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Cria um novo usuário com dados aleatórios para fins de teste.
func seedUsuario(t *testing.T, store *Store) *Usuario {
	t.Helper()

	usuario := &Usuario{
		CPF:   rand.Text(),
		Email: rand.Text(),
	}
	err := store.SaveUsuario(t.Context(), usuario)
	if err != nil {
		t.Fatal(err)
	}

	return usuario
}

func TestUserLifecycle(t *testing.T) {
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

func TestUser_EmailTaken(t *testing.T) {
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

func TestUser_CPFTaken(t *testing.T) {
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
