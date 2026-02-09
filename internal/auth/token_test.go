package auth

import (
	"errors"
	"testing"
	"time"
)

func TestTokenLifecycle(t *testing.T) {
	t.Parallel()

	auth := newTestService(t)

	u, err := auth.CreateUsuario(t.Context(), CreateUsuarioParams{
		Nome:  "Fulano da Silva",
		CPF:   "123.456.789-09",
		Email: "fulano@email.com",
		Papel: PapelAnalista,
	})
	if err != nil {
		t.Fatal(err)
	}

	token, err := auth.CreateToken(t.Context(), u.ID, EscopoAuth, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !token.Expira.After(time.Now()) {
		t.Fatal("expected token expiration to be in the future")
	}
	if len(token.Token) < tokenSize {
		t.Fatal("token smaller than expected")
	}

	u2, err := auth.GetTokenOwner(t.Context(), token.Token, EscopoAuth)
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != u2.ID {
		t.Fatal("expected users to match")
	}

	err = auth.DeleteToken(t.Context(), token.Token)
	if err != nil {
		t.Fatal(err)
	}

	_, err = auth.GetTokenOwner(t.Context(), token.Token, EscopoAuth)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
