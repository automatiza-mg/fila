package main

import (
	"errors"
	"net/http"

	"github.com/automatiza-mg/fila/internal/aposentadoria"
	"github.com/automatiza-mg/fila/internal/datalake"
	"github.com/automatiza-mg/fila/internal/validator"
)

func (app *application) handleServidoresDetail(w http.ResponseWriter, r *http.Request) {
	cpf := r.PathValue("cpf")

	if !validator.Matches(cpf, validator.CpfRX) {
		app.writeError(w, http.StatusBadRequest, "CPF inválido")
		return
	}

	servidor, err := app.apos.GetServidor(r.Context(), cpf)
	if err != nil {
		switch {
		case errors.Is(err, datalake.ErrNotFound):
			app.notFound(w, r)
		case errors.Is(err, aposentadoria.ErrInvalidCPF):
			app.writeError(w, http.StatusConflict, "O CPF informado não possui um processo de aposentadoria")
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.writeJSON(w, http.StatusOK, servidor)
}
