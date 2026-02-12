package main

import (
	"errors"
	"net/http"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/fila"
	"github.com/automatiza-mg/fila/internal/pagination"
	"github.com/google/uuid"
)

func (app *application) handleProcessoAposentadoriaList(w http.ResponseWriter, r *http.Request) {
	params := pagination.ParseQuery(r)

	result, err := app.fila.ListProcesso(r.Context(), fila.ListProcessoAposentadoriaParams{
		Status: r.URL.Query().Get("status"),
		Page:   params.Page,
		Limit:  params.Limit,
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, result)
}

func (app *application) handleProcessoDetailAposentadoria(w http.ResponseWriter, r *http.Request) {
	processoID, err := uuid.Parse(r.PathValue("processoID"))
	if err != nil {
		app.notFound(w, r)
		return
	}

	p, err := app.processos.GetProcesso(r.Context(), processoID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	pa, err := app.fila.GetProcessoByNumero(r.Context(), p.Numero)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, pa)
}
