package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

type errorPage struct {
	basePage
	Status      int
	Title       string
	Description string
}

// Renderiza a página de erro da aplicação com o status e descrição informadas.
func (app *application) serveErrorPage(w http.ResponseWriter, r *http.Request, status int, description string) {
	app.servePage(w, r, status, "erro.tmpl", errorPage{
		basePage:    app.newBasePage(r, fmt.Sprintf("Erro %d", status)),
		Status:      status,
		Title:       http.StatusText(status),
		Description: description,
	})
}

// Retorna responsta adequada para erros inesperados da aplicação (500).
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Error(
		"Erro interno do servidor",
		slog.String("err", err.Error()),
		slog.String("method", r.Method),
		slog.String("uri", r.URL.RequestURI()),
	)
	app.serveErrorPage(w, r, http.StatusInternalServerError, "Algo deu errado ao processar sua requisição")
}

// Retorna responsta adequada para requisições inválidas (400).
func (app *application) badRequest(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warn(
		"Requisição inválida",
		slog.String("err", err.Error()),
		slog.String("method", r.Method),
		slog.String("uri", r.URL.RequestURI()),
	)
	app.serveErrorPage(w, r, http.StatusBadRequest, "Requisição inválida ou malformada")
}
