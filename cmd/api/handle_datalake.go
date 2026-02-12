package main

import (
	"errors"
	"net/http"

	"github.com/automatiza-mg/fila/internal/datalake"
)

func (app *application) handleDatalakeProcessos(w http.ResponseWriter, r *http.Request) {
	unidade := r.URL.Query().Get("unidade")
	if unidade == "" {
		app.writeError(w, http.StatusBadRequest, "O parâmetro 'unidade' deve ser informado")
		return
	}

	processos, err := app.apos.ListProcessosAbertos(r.Context(), unidade)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, processos)
}

func (app *application) handleDatalakeUnidadesProcessos(w http.ResponseWriter, r *http.Request) {
	unidades, err := app.apos.ListUnidadesDisponiveis(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, unidades)
}

func (app *application) handleDatalakeServidor(w http.ResponseWriter, r *http.Request) {
	cpf := r.PathValue("cpf")

	servidor, err := app.apos.GetServidorByCPF(r.Context(), cpf)
	if err != nil {
		switch {
		case errors.Is(err, datalake.ErrNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.writeJSON(w, http.StatusOK, servidor)
}
