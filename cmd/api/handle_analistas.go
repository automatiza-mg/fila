package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/validator"
)

var orgaosAllowed = []string{
	"SEPLAG",
	"SEE",
}

type analistaCreateRequest struct {
	UnidadeID string `json:"unidade_id"`
	Orgao     string `json:"orgao"`

	validator.Validator `json:"-"`
}

func (app *application) handleAnalistaCreate(w http.ResponseWriter, r *http.Request) {
	usuario := app.getUsuario(r.Context())
	if !usuario.HasPapel(database.PapelAnalista) {
		app.writeJSON(w, http.StatusForbidden, ErrorResponse{
			Message: "Apenas usuários com papel de analista podem ter dados complementares cadastrados.",
		})
		return
	}

	var input analistaCreateRequest
	err := app.decodeJSON(w, r, &input)
	if err != nil {
		app.decodeError(w, r, err)
		return
	}

	unidadesMap, err := app.getUnidadesMap(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	unidade, unidadeOk := unidadesMap[input.UnidadeID]

	input.Check(validator.PermittedValue(input.Orgao, orgaosAllowed...), "orgao", fmt.Sprintf("Deve ser um dos valores: %s", strings.Join(orgaosAllowed, ", ")))
	input.Check(unidadeOk, "unidade_id", "A unidade informada deve ser válida")
	if !input.Valid() {
		app.validationFailed(w, r, input.FieldErrors)
		return
	}

	analista := &database.Analista{
		UsuarioID:       usuario.ID,
		Orgao:           input.Orgao,
		SEIUnidadeID:    unidade.IdUnidade,
		SEIUnidadeSigla: unidade.Sigla,
	}
	err = app.store.SaveAnalista(r.Context(), analista)
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

	w.Header().Set("Location", fmt.Sprintf("/api/v1/usuarios/%d", usuario.ID))
	app.writeJSON(w, http.StatusCreated, analista)
}
