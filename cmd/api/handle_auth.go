package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/automatiza-mg/fila/internal/auth"
	"github.com/automatiza-mg/fila/internal/validator"
)

type EntrarRequest struct {
	CPF   string `json:"cpf"`
	Senha string `json:"senha"`

	validator.Validator `json:"-"`
}

func (app *application) handleAuthEntrar(w http.ResponseWriter, r *http.Request) {
	var input EntrarRequest
	err := app.decodeJSON(w, r, &input)
	if err != nil {
		app.decodeError(w, r, err)
		return
	}

	ctx := r.Context()

	input.Check(validator.NotBlank(input.CPF), "cpf", "Deve ser informado")
	input.Check(validator.Matches(input.CPF, validator.CpfRX), "cpf", "Deve ser um CPF válido")
	input.Check(validator.NotBlank(input.Senha), "senha", "Deve ser informado")
	input.Check(validator.MaxLength(input.Senha, 60), "senha", "Deve possuir até 60 caracteres")
	if !input.Valid() {
		app.validationFailed(w, r, input.FieldErrors)
		return
	}

	usuario, err := app.auth.Authenticate(ctx, input.CPF, input.Senha)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidCredentials):
			app.writeJSON(w, http.StatusUnauthorized, ErrorResponse{
				Message: "Credenciais inválidas",
			})
		default:
			app.serverError(w, r, err)
		}
		return
	}

	token, err := app.auth.CreateToken(r.Context(), usuario.ID, auth.EscopoAuth, 24*time.Hour)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, token)
}

func (app *application) handleAuthCadastrarDetalhe(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		app.tokenError(w, r)
		return
	}

	usuario, err := app.auth.GetTokenOwner(r.Context(), token, auth.EscopoSetup)
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

type CadastrarRequest struct {
	Token          string `json:"token"`
	Senha          string `json:"senha"`
	ConfirmarSenha string `json:"confirmar_senha"`

	validator.Validator `json:"-"`
}

func (app *application) handleAuthCadastrar(w http.ResponseWriter, r *http.Request) {
	var input CadastrarRequest
	err := app.decodeJSON(w, r, &input)
	if err != nil {
		app.decodeError(w, r, err)
		return
	}

	input.Check(validator.NotBlank(input.Senha), "senha", "Deve ser informado")
	input.Check(validator.StrongPassword(input.Senha), "senha", "Deve possuir pelo menos 8 caracteres, um dígito e um caractere especial")
	input.Check(validator.MaxLength(input.Senha, 60), "senha", "Deve possuir no máximo 60 caracteres")
	input.Check(validator.NotBlank(input.ConfirmarSenha), "confirmar_senha", "Deve ser informado")
	input.Check(input.Senha == input.ConfirmarSenha, "confirmar_senha", "Senhas devem ser idênticas")
	input.Check(validator.NotBlank(input.Token), "token", "Deve ser informado")
	if !input.Valid() {
		app.validationFailed(w, r, input.FieldErrors)
		return
	}

	err = app.auth.SetupUsuario(r.Context(), auth.SetupUsuarioParams{
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

// Retorna os dados do usuário autenticado.
func (app *application) handleAuthUsuarioAtual(w http.ResponseWriter, r *http.Request) {
	usuario := app.getAuth(r.Context())
	app.writeJSON(w, http.StatusOK, usuario)
}
