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

				r.Post("/enviar-cadastro", app.handleUsuarioEnviarCadastro)

				r.Get("/analista", app.handleAnalistaDetail)
				r.Post("/analista", app.handleAnalistaCreate)

				r.Post("/analista/afastar", app.handleAnalistaAfastar)
				r.Post("/analista/retornar", app.handleAnalistaRetornar)
			})
		})

		r.Route("/analistas", func(r chi.Router) {
			r.Get("/", app.handleAnalistaList)
		})

		r.Route("/unidades", func(r chi.Router) {
			r.Get("/", app.handleUnidadeList)
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/entrar", app.handleAuthEntrar)
			r.Get("/cadastrar", app.handleAuthCadastrarDetalhe)
			r.Post("/cadastrar", app.handleAuthCadastrar)

			r.Group(func(r chi.Router) {
				r.Use(app.requireAuth)

				r.Get("/me", app.handleAuthUsuarioAtual)
			})
		})
	})

	return r
}
