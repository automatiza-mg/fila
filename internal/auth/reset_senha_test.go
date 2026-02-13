package auth

import (
	"errors"
	"testing"
	"time"
)

func TestSendResetSenha(t *testing.T) {
	t.Parallel()

	svc, queue := newTestServiceWithQueue(t)

	u, err := svc.CreateAdmin(t.Context(), CreateAdminParams{
		Nome:  "Test User",
		CPF:   "111.111.111-11",
		Email: "test@example.com",
		Senha: "Senha@123",
	})
	if err != nil {
		t.Fatal(err)
	}

	tokenFn := func(token string) string {
		return "https://example.com/reset?token=" + token
	}

	err = svc.SendResetSenha(t.Context(), u.CPF, tokenFn)
	if err != nil {
		t.Fatal(err)
	}

	// Verifica se o email foi enfileirado.
	args := queue.Args()
	if len(args) != 1 {
		t.Fatalf("expected 1 task, got: %d", len(args))
	}
}

func TestSendResetSenha_UserNotFound(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)

	tokenFn := func(token string) string {
		return "https://example.com/reset?token=" + token
	}

	err := svc.SendResetSenha(t.Context(), "999.999.999-99", tokenFn)
	if err != nil {
		t.Fatalf("expected nil error for non-existent user, got: %v", err)
	}
}

func TestSendResetSenha_EmailNotVerified(t *testing.T) {
	t.Parallel()

	svc, queue := newTestServiceWithQueue(t)

	// Cria um usuário sem email verificado.
	_, err := svc.CreateUsuario(t.Context(), CreateUsuarioParams{
		Nome:  "Unverified User",
		CPF:   "222.222.222-22",
		Email: "unverified@example.com",
		Papel: PapelAnalista,
	})
	if err != nil {
		t.Fatal(err)
	}

	tokenFn := func(token string) string {
		return "https://example.com/reset?token=" + token
	}

	err = svc.SendResetSenha(t.Context(), "222.222.222-22", tokenFn)
	if err != nil {
		t.Fatalf("expected nil error for unverified email, got: %v", err)
	}

	// Nenhum email deve ter sido enfileirado.
	if len(queue.Args()) != 0 {
		t.Fatalf("expected 0 tasks, got: %d", len(queue.Args()))
	}
}

func TestResetSenha(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)

	u, err := svc.CreateAdmin(t.Context(), CreateAdminParams{
		Nome:  "Test User",
		CPF:   "333.333.333-33",
		Email: "reset@example.com",
		Senha: "Senha@123",
	})
	if err != nil {
		t.Fatal(err)
	}

	token, err := svc.CreateToken(t.Context(), u.ID, EscopoResetSenha, 1*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	err = svc.ResetSenha(t.Context(), ResetSenhaParams{
		Token: token.Token,
		Senha: "NovaSenha@456",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Deve autenticar com a nova senha.
	_, err = svc.Authenticate(t.Context(), u.CPF, "NovaSenha@456")
	if err != nil {
		t.Fatalf("expected authentication with new password to succeed, got: %v", err)
	}

	// Não deve autenticar com a senha antiga.
	_, err = svc.Authenticate(t.Context(), u.CPF, "Senha@123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials with old password, got: %v", err)
	}

	// Token deve ser invalidado após uso.
	_, err = svc.GetTokenOwner(t.Context(), token.Token, EscopoResetSenha)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken after reset, got: %v", err)
	}
}

func TestResetSenha_InvalidToken(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)

	err := svc.ResetSenha(t.Context(), ResetSenhaParams{
		Token: "invalid-token",
		Senha: "NovaSenha@456",
	})
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got: %v", err)
	}
}
