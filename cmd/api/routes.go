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

		// TODO: Adicionar verificação de admin.
		r.Route("/datalake", func(r chi.Router) {
			r.Get("/processos", app.handleDatalakeProcessos)
			r.Get("/processos/unidades", app.handleDatalakeUnidadesProcessos)

			r.Get("/servidores/{cpf}", app.handleDatalakeServidor)
		})

		r.Route("/usuarios", func(r chi.Router) {
			r.Get("/", app.handleUsuarioList)
			r.Post("/", app.handleUsuarioCreate)

			r.Route("/{usuarioID}", func(r chi.Router) {
				r.Use(app.loadUsuario)

				r.Get("/", app.handleUsuarioDetail)
				r.Delete("/", app.handleUsuarioDelete)

				r.Post("/analista", app.handleAnalistaCreate)
			})
		})

		r.Route("/unidades", func(r chi.Router) {
			r.Get("/", app.handleUnidadeList)
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/entrar", app.handleAuthEntrar)

			r.Group(func(r chi.Router) {
				r.Use(app.requireAuth)

				r.Get("/me", app.handleAuthUsuarioAtual)
			})
		})
	})

	return r
}
