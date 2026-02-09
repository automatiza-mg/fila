package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/automatiza-mg/fila/internal/auth"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/validator"
)

func (app *application) loadUsuario(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		usuarioID, err := app.intParam(r, "usuarioID")
		if err != nil || usuarioID < 1 {
			app.notFound(w, r)
			return
		}

		usuario, err := app.auth.GetUsuario(r.Context(), usuarioID)
		if err != nil {
			switch {
			case errors.Is(err, database.ErrNotFound):
				app.notFound(w, r)
			default:
				app.serverError(w, r, err)
			}
			return
		}

		ctx := app.setUsuario(r.Context(), usuario)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) handleUsuarioList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	papel := q.Get("papel")

	usuarios, err := app.auth.ListUsuarios(r.Context(), auth.ListUsuariosParams{
		Papel: papel,
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, usuarios)
}

// Lista de papeis permitidos
var papeisAllowed = []string{
	auth.PapelAnalista,
	auth.PapelGestor,
	auth.PapelSubsecretario,
}

type usuarioCreateRequest struct {
	Nome  string `json:"nome"`
	CPF   string `json:"cpf"`
	Email string `json:"email"`
	Papel string `json:"papel"`

	validator.Validator `json:"-"`
}

func (app *application) handleUsuarioCreate(w http.ResponseWriter, r *http.Request) {
	var input usuarioCreateRequest
	err := app.decodeJSON(w, r, &input)
	if err != nil {
		app.decodeError(w, r, err)
		return
	}

	input.Check(validator.NotBlank(input.Nome), "nome", "Campo obrigatório")
	input.Check(validator.MaxLength(input.Nome, 255), "nome", "Deve possuir até 255 caracteres")
	input.Check(validator.NotBlank(input.CPF), "cpf", "Campo obrigatório")
	input.Check(validator.Matches(input.CPF, validator.CpfRX), "cpf", "Deve ser um CPF válido")
	input.Check(validator.NotBlank(input.Email), "email", "Campo obrigatório")
	input.Check(validator.MaxLength(input.Email, 255), "email", "Deve possuir até 255 caracteres")
	input.Check(validator.Matches(input.Email, validator.EmailRX), "email", "Deve ser um email válido")
	input.Check(validator.PermittedValue(input.Papel, papeisAllowed...), "papel", fmt.Sprintf("Deve ser um dos valores: %s", strings.Join(papeisAllowed, ", ")))
	if !input.Valid() {
		app.validationFailed(w, r, input.FieldErrors)
		return
	}

	usuario, err := app.auth.CreateUsuario(r.Context(), auth.CreateUsuarioParams{
		Nome:  input.Nome,
		CPF:   input.CPF,
		Email: input.Email,
		Papel: input.Papel,
		TokenURL: func(token string) string {
			return fmt.Sprintf("%s/cadastro?token=%s", app.cfg.BaseURL, token)
		},
	})
	if err != nil {
		switch {
		case errors.Is(err, database.ErrUsuarioCPFTaken):
			input.SetFieldError("cpf", "Valor já está em uso")
			app.validationFailed(w, r, input.FieldErrors)
		case errors.Is(err, database.ErrUsuarioEmailTaken):
			input.SetFieldError("email", "Valor já está em uso")
			app.validationFailed(w, r, input.FieldErrors)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.writeJSON(w, http.StatusCreated, usuario)
}

func (app *application) handleUsuarioDetail(w http.ResponseWriter, r *http.Request) {
	usuario := app.getUsuario(r.Context())
	app.writeJSON(w, http.StatusOK, usuario)
}

func (app *application) handleUsuarioDelete(w http.ResponseWriter, r *http.Request) {
	usuario := app.getUsuario(r.Context())

	err := app.auth.DeleteUsuario(r.Context(), usuario)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) handleUsuarioEnviarCadastro(w http.ResponseWriter, r *http.Request) {
	usuario := app.getUsuario(r.Context())

	if usuario.EmailVerificado {
		app.writeError(w, http.StatusBadRequest, "Usuário já possui um cadastro ativo.")
		return
	}

	err := app.auth.SendSetup(r.Context(), usuario, func(token string) string {
		return fmt.Sprintf("%s/cadastrar?token=%s", app.cfg.BaseURL, token)
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
