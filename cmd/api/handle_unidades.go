package main

import (
	"net/http"
)

// Lista as unidades do SEI disponíveis para os analistas.
func (app *application) handleUnidadeList(w http.ResponseWriter, r *http.Request) {
	unidades, err := app.analistas.ListUnidades(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, unidades)
}
