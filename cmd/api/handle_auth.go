package main

import (
	"errors"
	"net/http"
	"time"

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

	usuario, err := app.store.GetUsuarioByCPF(ctx, input.CPF)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.writeJSON(w, http.StatusUnauthorized, ErrorResponse{
				Message: "Credenciais inválidas",
			})
		default:
			app.serverError(w, r, err)
		}
		return
	}

	ok, err := usuario.CheckSenha(input.Senha)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if !ok {
		app.writeJSON(w, http.StatusUnauthorized, ErrorResponse{
			Message: "Credenciais inválidas",
		})
		return
	}

	token, err := app.store.CreateToken(ctx, usuario.ID, database.EscopoAuth, 24*time.Hour)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, token)
}

// Retorna os dados do usuário autenticado.
func (app *application) handleAuthUsuarioAtual(w http.ResponseWriter, r *http.Request) {
	usuario := app.getAuth(r.Context())
	app.writeJSON(w, http.StatusOK, usuario)
}
