package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(app.static)))

	mux.HandleFunc("GET /entrar", app.handleEntrar)
	mux.HandleFunc("POST /entrar", app.handleEntrarPost)

	mux.HandleFunc("GET /cadastrar", app.handleCadastrar)
	mux.HandleFunc("POST /cadastrar", app.handleCadastrarPost)

	mux.HandleFunc("GET /gestor/usuarios/criar", app.handleUsuarioCriar)
	mux.HandleFunc("POST /gestor/usuarios/criar", app.handleUsuarioCriarPost)

	return app.csrfProtection(mux)
}
