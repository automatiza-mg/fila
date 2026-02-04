package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Define as rotas da API.
func (app *application) routes() http.Handler {
	r := chi.NewMux()

	r.Route("/api/v1", func(r chi.Router) {
		r.NotFound(app.notFound)
		r.MethodNotAllowed(app.methodNotAllowed)

		r.Use(app.authenticate)

		// Unidades SEI
		r.Get("/unidades", app.requireAuth(app.handleUnidadeList))

		// Usuários
		r.Route("/usuarios", func(r chi.Router) {
			r.Get("/", app.requireAuth(app.handleUsuarioList))
			r.Get("/{usuarioID}", app.requireAuth(app.handleUsuarioDetail))
		})

		// Auth
		r.Route("/auth", func(r chi.Router) {
			r.Post("/entrar", app.handleAuthEntrar)

			r.Get("/me", app.requireAuth(app.handleAuthUsuarioAtual))
		})
	})

	return r
}
