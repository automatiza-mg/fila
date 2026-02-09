package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/fila"
	"github.com/automatiza-mg/fila/internal/validator"
)

var orgaosAllowed = []string{
	"SEPLAG",
	"SEE",
}

type AnalistaCreateRequest struct {
	SEIUnidadeID string `json:"sei_unidade_id"`
	Orgao        string `json:"orgao"`

	validator.Validator `json:"-"`
}

func (app *application) handleAnalistaCreate(w http.ResponseWriter, r *http.Request) {
	usuario := app.getUsuario(r.Context())
	if !usuario.HasPapel(database.PapelAnalista) {
		app.writeError(w, http.StatusForbidden, "Apenas usuários com papel de analista podem ter dados complementares cadastrados")
		return
	}

	var input AnalistaCreateRequest
	err := app.decodeJSON(w, r, &input)
	if err != nil {
		app.decodeError(w, r, err)
		return
	}

	unidadesMap, err := app.fila.GetUnidadesMap(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	unidade, unidadeOk := unidadesMap[input.SEIUnidadeID]

	input.Check(validator.PermittedValue(input.Orgao, orgaosAllowed...), "orgao", fmt.Sprintf("Deve ser um dos valores: %s", strings.Join(orgaosAllowed, ", ")))
	input.Check(unidadeOk, "unidade_id", "A unidade informada deve ser válida")
	if !input.Valid() {
		app.validationFailed(w, r, input.FieldErrors)
		return
	}

	a, err := app.fila.CreateAnalista(r.Context(), fila.CreateAnalistaParams{
		UsuarioID:    usuario.ID,
		SeiUnidadeID: unidade.ID,
		Orgao:        input.Orgao,
	})
	if err != nil {
		switch {
		case errors.Is(err, database.ErrAnalistaExists):
			app.writeJSON(w, http.StatusConflict, ErrorResponse{
				Message: "O usuário já possui dados complementares de analista cadastrados.",
			})
		default:
			app.serverError(w, r, err)
		}
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/api/v1/usuarios/%d/analista", usuario.ID))
	app.writeJSON(w, http.StatusCreated, a)
}

func (app *application) handleAnalistaList(w http.ResponseWriter, r *http.Request) {
	analistas, err := app.fila.ListAnalistas(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, analistas)
}

func (app *application) handleAnalistaDetail(w http.ResponseWriter, r *http.Request) {
	usuario := app.getUsuario(r.Context())
	if !usuario.HasPapel(database.PapelAnalista) {
		app.writeError(w, http.StatusForbidden, "Apenas usuários com papel de analista podem ter dados complementares cadastrados")
		return
	}

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

func (app *application) handleAnalistaAfastar(w http.ResponseWriter, r *http.Request) {
	usuario := app.getUsuario(r.Context())
	if !usuario.HasPapel(database.PapelAnalista) {
		app.writeError(w, http.StatusForbidden, "Apenas usuários com papel de analista podem ter dados complementares cadastrados")
		return
	}

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

	err = app.fila.AfastarAnalista(r.Context(), analista.UsuarioID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) handleAnalistaRetornar(w http.ResponseWriter, r *http.Request) {
	usuario := app.getUsuario(r.Context())
	if !usuario.HasPapel(database.PapelAnalista) {
		app.writeError(w, http.StatusForbidden, "Apenas usuários com papel de analista podem ter dados complementares cadastrados")
		return
	}

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

	err = app.fila.RetornarAnalista(r.Context(), analista.UsuarioID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
