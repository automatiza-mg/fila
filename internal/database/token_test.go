package database

import (
	"crypto/rand"
	"encoding/base64"
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

	err = store.DeleteToken(t.Context(), token.Hash)
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

func TestToken_SaveToken(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	usuario := seedUsuario(t, store)

	b := make([]byte, 32)
	_, _ = rand.Read(b)
	plaintext := base64.RawURLEncoding.EncodeToString(b)

	token := &Token{
		Plaintext: plaintext,
		Hash:      hashToken(plaintext),
		UsuarioID: usuario.ID,
		Escopo:    "manual",
		ExpiraEm:  time.Now().Add(time.Hour),
	}
	err := store.SaveToken(t.Context(), token)
	if err != nil {
		t.Fatal(err)
	}

	read, err := store.GetUsuarioForToken(t.Context(), plaintext, "manual")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(usuario, read); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}

	err = store.DeleteToken(t.Context(), token.Hash)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.GetUsuarioForToken(t.Context(), plaintext, "manual")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestToken_DeleteTokensUsuario(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	usuario := seedUsuario(t, store)

	rs1, err := store.CreateToken(t.Context(), usuario.ID, "reset-senha", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	rs2, err := store.CreateToken(t.Context(), usuario.ID, "reset-senha", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	setup, err := store.CreateToken(t.Context(), usuario.ID, "setup", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	err = store.DeleteTokensUsuario(t.Context(), usuario.ID, "reset-senha")
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.GetUsuarioForToken(t.Context(), rs1.Plaintext, "reset-senha")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for rs1, got %v", err)
	}
	_, err = store.GetUsuarioForToken(t.Context(), rs2.Plaintext, "reset-senha")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for rs2, got %v", err)
	}

	read, err := store.GetUsuarioForToken(t.Context(), setup.Plaintext, "setup")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(usuario, read); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}
}

func TestToken_WrongEscopo(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	usuario := seedUsuario(t, store)

	token, err := store.CreateToken(t.Context(), usuario.ID, "setup", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.GetUsuarioForToken(t.Context(), token.Plaintext, "reset-senha")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	read, err := store.GetUsuarioForToken(t.Context(), token.Plaintext, "setup")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(usuario, read); diff != "" {
		t.Fatalf("mismatch:\n%s", diff)
	}
}
