package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error("Erro interno do servidor", slog.String("err", err.Error()), slog.String("uri", r.URL.RequestURI()))
	app.writeJSON(w, http.StatusInternalServerError, ErrorResponse{
		Message: "O servidor encontrou um erro inesperado ao processar sua requisição",
	})
}

func (app *application) decodeError(w http.ResponseWriter, _ *http.Request, _ error) {
	app.writeJSON(w, http.StatusBadRequest, ErrorResponse{
		Message: "A requisição é inválida ou malformada",
	})
}

func (app *application) notFound(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusNotFound, ErrorResponse{
		Message: "O recurso solicitado não foi encontrado",
	})
}

func (app *application) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
		Message: fmt.Sprintf("O método %q não é permitido para esse recurso", r.Method),
	})
}
