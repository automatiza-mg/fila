package main

import "net/http"

func (app *application) handleSeiListarUnidades(w http.ResponseWriter, r *http.Request) {
	resp, err := app.sei.ListarUnidades(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, resp.Parametros.Items)
}
