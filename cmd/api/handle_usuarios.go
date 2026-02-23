package main

import (
	"errors"
	"fmt"
	"net/http"

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

type usuarioCreateRequest struct {
	Nome  string `json:"nome"`
	CPF   string `json:"cpf"`
	Email string `json:"email"`
	Papel string `json:"papel"`
}

func (app *application) handleUsuarioCreate(w http.ResponseWriter, r *http.Request) {
	var input usuarioCreateRequest
	err := app.decodeJSON(w, r, &input)
	if err != nil {
		app.decodeError(w, r, err)
		return
	}

	params := auth.CreateUsuarioParams{
		Nome:     input.Nome,
		CPF:      input.CPF,
		Email:    input.Email,
		Papel:    input.Papel,
		TokenURL: app.setupURL,
	}

	v := validator.New()
	auth.ValidateCreateUsuario(v, params)
	if !v.Valid() {
		app.validationFailed(w, r, v.FieldErrors)
		return
	}

	usuario, err := app.auth.CreateUsuario(r.Context(), params)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrUsuarioCPFTaken):
			v.SetFieldError("cpf", "Valor já está em uso")
			app.validationFailed(w, r, v.FieldErrors)
		case errors.Is(err, database.ErrUsuarioEmailTaken):
			v.SetFieldError("email", "Valor já está em uso")
			app.validationFailed(w, r, v.FieldErrors)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/usuarios/%d", usuario.ID))
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
		app.badRequest(w, r, "Usuário já possui um cadastro ativo.")
		return
	}

	err := app.auth.SendSetup(r.Context(), usuario, app.setupURL)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
