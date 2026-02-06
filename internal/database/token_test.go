package database

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestToken_Lifecycle(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	usuario := seedUsuario(t, store)

	token, err := store.CreateToken(t.Context(), usuario.ID, "setup", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if token.Plaintext == "" {
		t.Fatal("expected token to return plaintext on creation")
	}

	read, err := store.GetUsuarioForToken(t.Context(), token.Plaintext, token.Escopo)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(usuario, read); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}

	err = store.DeleteTokensUsuario(t.Context(), usuario.ID, token.Escopo)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.GetUsuarioForToken(t.Context(), token.Plaintext, token.Escopo)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestToken_Invalid(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	usuario := seedUsuario(t, store)

	_, err := store.GetUsuarioForToken(t.Context(), "foobar", "reset-senha")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	// Cria um token expirado
	token, err := store.CreateToken(t.Context(), usuario.ID, "reset-senha", -time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.GetUsuarioForToken(t.Context(), token.Plaintext, token.Escopo)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
