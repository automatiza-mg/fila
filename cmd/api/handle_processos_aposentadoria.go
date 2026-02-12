package main

import (
	"errors"
	"net/http"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/fila"
	"github.com/automatiza-mg/fila/internal/pagination"
)

func (app *application) handleProcessoAposentadoriaList(w http.ResponseWriter, r *http.Request) {
	params := pagination.ParseQuery(r)

	result, err := app.fila.ListProcesso(r.Context(), fila.ListProcessoAposentadoriaParams{
		Numero: r.URL.Query().Get("numero"),
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

func (app *application) handleProcessoAposentadoriaDetail(w http.ResponseWriter, r *http.Request) {
	paID, err := app.intParam(r, "paID")
	if err != nil || paID < 1 {
		app.notFound(w, r)
		return
	}

	pa, err := app.fila.GetProcessoAposentadoria(r.Context(), paID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.writeJSON(w, http.StatusOK, pa)
}

func (app *application) handleProcessoAposentadoriaHistorico(w http.ResponseWriter, r *http.Request) {
	paID, err := app.intParam(r, "paID")
	if err != nil || paID < 1 {
		app.notFound(w, r)
		return
	}

	historico, err := app.fila.ListHistorico(r.Context(), paID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if historico == nil {
		historico = make([]*fila.HistoricoStatusProcesso, 0)
	}

	app.writeJSON(w, http.StatusOK, historico)
}
