package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/automatiza-mg/fila/internal/auth"
	"github.com/automatiza-mg/fila/internal/database"
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
		case errors.Is(err, auth.ErrNoPassword):
			app.badRequest(w, r, "O usuário não possui uma senha cadastrada")
		case errors.Is(err, auth.ErrInvalidCredentials):
			app.writeError(w, http.StatusUnauthorized, "Credenciais inválidas")
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

// Retorna os dados do usuário dono de um token. Requer os query params `token` e `escopo`.
// Escopos permitidos: "setup" e "reset-senha".
func (app *application) handleAuthTokenInfo(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		app.tokenError(w, r)
		return
	}

	escopoParam := r.URL.Query().Get("escopo")
	var escopo auth.Escopo
	switch escopoParam {
	case auth.EscopoSetup.String():
		escopo = auth.EscopoSetup
	case auth.EscopoResetSenha.String():
		escopo = auth.EscopoResetSenha
	default:
		app.badRequest(w, r, "Escopo inválido")
		return
	}

	usuario, err := app.auth.GetTokenOwner(r.Context(), token, escopo)
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

// Retorna os dados de analista do usuário autenticado.
func (app *application) handleAuthAnalistaAtual(w http.ResponseWriter, r *http.Request) {
	usuario := app.getAuth(r.Context())

	analista, err := app.fila.GetAnalista(r.Context(), usuario.ID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.writeJSON(w, http.StatusOK, analista)
}

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
