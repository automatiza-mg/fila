package main

import (
	"net/http"

	"github.com/automatiza-mg/fila/internal/fila"
	"github.com/automatiza-mg/fila/internal/pagination"
)

func (app *application) handleSolicitacoesPrioridadeList(w http.ResponseWriter, r *http.Request) {
	params := pagination.ParseQuery(r)

	ssp, err := app.fila.ListSolicitacoesPrioridade(r.Context(), fila.ListSolicitacoesPrioridadeParams{
		Page:  params.Page,
		Limit: params.Limit,
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, ssp)
}

func (app *application) handleSolicitacoesPrioridadeDetail(w http.ResponseWriter, r *http.Request) {
	spID, err := app.intParam(r, "spID")
	if err != nil || spID < 1 {
		app.notFound(w, r)
		return
	}

	sp, err := app.fila.GetSolicitacaoPrioridade(r.Context(), spID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, sp)
}

func (app *application) handleSolicitacoesPrioridadeAprovar(w http.ResponseWriter, r *http.Request) {
	spID, err := app.intParam(r, "spID")
	if err != nil || spID < 1 {
		app.notFound(w, r)
		return
	}

	err = app.fila.AprovarSolicitacaoPrioridade(r.Context(), spID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) handleSolicitacoesPrioridadeNegar(w http.ResponseWriter, r *http.Request) {
	spID, err := app.intParam(r, "spID")
	if err != nil || spID < 1 {
		app.notFound(w, r)
		return
	}

	err = app.fila.NegarSolicitacaoPrioridade(r.Context(), spID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
