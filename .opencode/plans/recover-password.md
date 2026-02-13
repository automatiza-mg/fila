# Recover Password Feature - Implementation Plan

## Overview

Implement password recovery (forgot password) flow for the Fila application. The infrastructure is partially prepared — `EscopoResetSenha = "reset-senha"` is already defined in `internal/auth/token.go` but never used. The implementation follows the same pattern as the existing **setup flow** (`SendSetup` → token → email → `SetupUsuario`).

## Flow

1. **Request reset** (POST `/auth/recuperar-senha`): User submits their CPF. If a matching user exists (and has a verified email), a reset token is created and an email is sent with a reset link. Always returns 202 (no user enumeration).
2. **Validate token** (GET `/auth/recuperar-senha`): Frontend GETs with the token to verify it's valid and get minimal user info.
3. **Reset password** (POST `/auth/redefinir-senha`): User submits the token + new password + confirmation. Password is updated, all reset tokens for that user are invalidated.

## Changes by File

### 1. NEW: `internal/mail/templates/reset-senha.tmpl`

```tmpl
{{define "subject"}}
  Recuperação de Senha - Fila Aposentadoria
{{end}}

{{define "text"}}
Acesse o link abaixo para redefinir sua senha:
{{.ResetURL}}

Caso não tenha solicitado a recuperação, ignore este email.
{{end}}

{{define "html"}}{{end}}
```

### 2. EDIT: `internal/mail/templates.go`

Add after line 13 (after `setupTmpl`):
```go
resetSenhaTmpl = template.Must(template.ParseFS(fs, "templates/reset-senha.tmpl"))
```

Add after `NewSetupEmail` function (after line 51):
```go
type ResetSenhaEmailParams struct {
	// A URL para redefinição de senha do usuário.
	ResetURL string
}

// NewResetSenhaEmail retorna um novo [Email] para o template `reset-senha.tmpl`, possibilitando a redefinição de senha do usuário.
func NewResetSenhaEmail(to string, params ResetSenhaEmailParams) (*Email, error) {
	return executeTemplate(resetSenhaTmpl, []string{to}, params)
}
```

### 3. EDIT: `internal/database/usuario.go`

Add after `DeleteUsuario` function (after line 222):
```go
// UpdateUsuarioSenha atualiza apenas a senha (hash) de um usuário.
func (s *Store) UpdateUsuarioSenha(ctx context.Context, usuarioID int64, hashSenha string) error {
	q := `
	UPDATE usuarios SET
		hash_senha = $2,
		atualizado_em = CURRENT_TIMESTAMP
	WHERE id = $1`
	_, err := s.db.Exec(ctx, q, usuarioID, hashSenha)
	return err
}
```

### 4. EDIT: `internal/auth/service.go`

Add new params struct and two new methods after `SendSetup` (after line 208):

```go
// SendResetSenha envia um email de recuperação de senha para o usuário identificado pelo CPF.
// Retorna nil silenciosamente caso o usuário não exista ou não tenha email verificado,
// evitando enumeração de contas.
func (s *Service) SendResetSenha(ctx context.Context, cpf string, tokenFn func(token string) string) error {
	r, err := s.store.GetUsuarioByCPF(ctx, cpf)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil
		}
		return err
	}

	if !r.EmailVerificado {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	// Remove tokens de reset anteriores.
	err = store.DeleteTokensUsuario(ctx, r.ID, EscopoResetSenha.String())
	if err != nil {
		return err
	}

	token, err := s.createToken(ctx, store, r.ID, EscopoResetSenha, 1*time.Hour)
	if err != nil {
		return err
	}

	email, err := mail.NewResetSenhaEmail(r.Email, mail.ResetSenhaEmailParams{
		ResetURL: tokenFn(token.Token),
	})
	if err != nil {
		return err
	}

	_, err = s.queue.InsertTx(ctx, tx, tasks.SendEmailArgs{
		Email: email,
	}, nil)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

type ResetSenhaParams struct {
	Token string
	Senha string
}

// ResetSenha redefine a senha do usuário utilizando um token de recuperação válido.
// Retorna [ErrInvalidToken] se o token for inválido ou expirado.
func (s *Service) ResetSenha(ctx context.Context, params ResetSenhaParams) error {
	r, err := s.store.GetUsuarioForToken(ctx, params.Token, EscopoResetSenha.String())
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return ErrInvalidToken
		}
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(params.Senha), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	err = store.UpdateUsuarioSenha(ctx, r.ID, string(hash))
	if err != nil {
		return err
	}

	err = store.DeleteTokensUsuario(ctx, r.ID, EscopoResetSenha.String())
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
```

### 5. EDIT: `internal/auth/validation.go`

Add after `ValidateCreateAdmin` (after line 20):
```go
// ValidateResetSenha valida os parâmetros para redefinição de senha.
func ValidateResetSenha(v *validator.Validator, senha, confirmarSenha string) {
	v.Check(validator.NotBlank(senha), "senha", "Campo obrigatório")
	v.Check(validator.MinLength(senha, 8), "senha", "Deve possuir no mínimo 8 caracteres")
	v.Check(validator.MaxLength(senha, 60), "senha", "Deve possuir no máximo 60 caracteres")
	v.Check(validator.StrongPassword(senha), "senha", "Deve possuir pelo menos um número e um caractere especial")
	v.Check(validator.NotBlank(confirmarSenha), "confirmar_senha", "Campo obrigatório")
	v.Check(senha == confirmarSenha, "confirmar_senha", "Senhas devem ser idênticas")
}
```

### 6. EDIT: `cmd/api/handle_auth.go`

Add three new handlers after `handleAuthAnalistaAtual` (after line 148):

```go
type RecuperarSenhaRequest struct {
	CPF string `json:"cpf"`

	validator.Validator `json:"-"`
}

// Envia um email de recuperação de senha para o usuário com o CPF informado.
func (app *application) handleAuthRecuperarSenha(w http.ResponseWriter, r *http.Request) {
	var input RecuperarSenhaRequest
	err := app.decodeJSON(w, r, &input)
	if err != nil {
		app.decodeError(w, r, err)
		return
	}

	input.Check(validator.NotBlank(input.CPF), "cpf", "Deve ser informado")
	input.Check(validator.Matches(input.CPF, validator.CpfRX), "cpf", "Deve ser um CPF válido")
	if !input.Valid() {
		app.validationFailed(w, r, input.FieldErrors)
		return
	}

	tokenFn := func(token string) string {
		return app.cfg.BaseURL + "/recuperar-senha?token=" + token
	}

	err = app.auth.SendResetSenha(r.Context(), input.CPF, tokenFn)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// Retorna os dados do usuário dono de um token de recuperação de senha.
func (app *application) handleAuthRecuperarSenhaDetalhe(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		app.tokenError(w, r)
		return
	}

	usuario, err := app.auth.GetTokenOwner(r.Context(), token, auth.EscopoResetSenha)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidToken):
			app.tokenError(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.writeJSON(w, http.StatusOK, usuario)
}

type RedefinirSenhaRequest struct {
	Token          string `json:"token"`
	Senha          string `json:"senha"`
	ConfirmarSenha string `json:"confirmar_senha"`

	validator.Validator `json:"-"`
}

// Redefine a senha do usuário utilizando um token de recuperação.
func (app *application) handleAuthRedefinirSenha(w http.ResponseWriter, r *http.Request) {
	var input RedefinirSenhaRequest
	err := app.decodeJSON(w, r, &input)
	if err != nil {
		app.decodeError(w, r, err)
		return
	}

	input.Check(validator.NotBlank(input.Token), "token", "Deve ser informado")
	input.Check(validator.NotBlank(input.Senha), "senha", "Deve ser informado")
	input.Check(validator.StrongPassword(input.Senha), "senha", "Deve possuir pelo menos 8 caracteres, um dígito e um caractere especial")
	input.Check(validator.MaxLength(input.Senha, 60), "senha", "Deve possuir no máximo 60 caracteres")
	input.Check(validator.NotBlank(input.ConfirmarSenha), "confirmar_senha", "Deve ser informado")
	input.Check(input.Senha == input.ConfirmarSenha, "confirmar_senha", "Senhas devem ser idênticas")
	if !input.Valid() {
		app.validationFailed(w, r, input.FieldErrors)
		return
	}

	err = app.auth.ResetSenha(r.Context(), auth.ResetSenhaParams{
		Token: input.Token,
		Senha: input.Senha,
	})
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidToken):
			app.tokenError(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

### 7. EDIT: `cmd/api/routes.go`

Add 3 routes under the `/auth` group (public, before the `r.Group` with `requireAuth`). After line 88:

```go
r.Post("/recuperar-senha", app.handleAuthRecuperarSenha)
r.Get("/recuperar-senha", app.handleAuthRecuperarSenhaDetalhe)
r.Post("/redefinir-senha", app.handleAuthRedefinirSenha)
```

### 8. EDIT: `internal/database/usuario_test.go`

Add after `TestUsuario_List` (after line 204):

```go
func TestUsuario_UpdateSenha(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	usuario := seedUsuario(t, store)

	err := store.UpdateUsuarioSenha(t.Context(), usuario.ID, "new-hash-senha")
	if err != nil {
		t.Fatal(err)
	}

	read, err := store.GetUsuario(t.Context(), usuario.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !read.HashSenha.Valid || read.HashSenha.V != "new-hash-senha" {
		t.Fatalf("expected hash_senha to be 'new-hash-senha', got: %v", read.HashSenha)
	}
}
```

### 9. EDIT: `internal/auth/service_test.go`

Add tests after `TestMain` (after line 63). These will need helpers from `usuario_test.go` (which already exists with `seedUsuario`).

```go
func TestSendResetSenha(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)

	// Create a user with verified email and password
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

	// Should succeed for valid user
	err = svc.SendResetSenha(t.Context(), u.CPF, tokenFn)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSendResetSenha_UserNotFound(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)

	tokenFn := func(token string) string {
		return "https://example.com/reset?token=" + token
	}

	// Should return nil for non-existent user
	err := svc.SendResetSenha(t.Context(), "999.999.999-99", tokenFn)
	if err != nil {
		t.Fatalf("expected nil error for non-existent user, got: %v", err)
	}
}

func TestResetSenha(t *testing.T) {
	t.Parallel()

	svc := newTestService(t)

	// Create a verified user
	u, err := svc.CreateAdmin(t.Context(), CreateAdminParams{
		Nome:  "Test User",
		CPF:   "222.222.222-22",
		Email: "reset@example.com",
		Senha: "Senha@123",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create a reset token directly
	token, err := svc.CreateToken(t.Context(), u.ID, EscopoResetSenha, 1*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// Reset the password
	err = svc.ResetSenha(t.Context(), ResetSenhaParams{
		Token: token.Token,
		Senha: "NovaSenha@456",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Should authenticate with new password
	_, err = svc.Authenticate(t.Context(), u.CPF, "NovaSenha@456")
	if err != nil {
		t.Fatalf("expected authentication with new password to succeed, got: %v", err)
	}

	// Should NOT authenticate with old password
	_, err = svc.Authenticate(t.Context(), u.CPF, "Senha@123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials with old password, got: %v", err)
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
```

Note: add `"time"` and `"errors"` to the import block in service_test.go.

## Files Changed Summary

| File | Action |
|------|--------|
| `internal/mail/templates/reset-senha.tmpl` | **New** |
| `internal/mail/templates.go` | Edit — add `resetSenhaTmpl`, `ResetSenhaEmailParams`, `NewResetSenhaEmail` |
| `internal/database/usuario.go` | Edit — add `UpdateUsuarioSenha` |
| `internal/auth/service.go` | Edit — add `SendResetSenha`, `ResetSenha`, `ResetSenhaParams` |
| `internal/auth/validation.go` | Edit — add `ValidateResetSenha` |
| `cmd/api/handle_auth.go` | Edit — add 3 handlers + request structs |
| `cmd/api/routes.go` | Edit — add 3 routes |
| `internal/database/usuario_test.go` | Edit — add `TestUsuario_UpdateSenha` |
| `internal/auth/service_test.go` | Edit — add 4 test functions |

## Design Decisions

- **Token TTL**: 1 hour (shorter than 72h setup tokens, since password reset is time-sensitive)
- **Identification by CPF**: Consistent with login flow (`handleAuthEntrar` uses CPF)
- **Silent failure on unknown CPF**: Prevents user enumeration attacks
- **No migration needed**: `tokens` table already supports arbitrary scopes
- **Separate `UpdateUsuarioSenha`**: Focused query avoids requiring all user fields like the general `UpdateUsuario`
