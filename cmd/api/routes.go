package main

import (
	"context"
	"net/http"

	"github.com/automatiza-mg/fila/internal/auth"
	"github.com/go-chi/chi/v5"
	"riverqueue.com/riverui"
)

// Define as rotas da API.
func (app *application) routes() http.Handler {
	r := chi.NewMux()

	if app.dev {
		r.Group(func(r chi.Router) {
			endpoints := riverui.NewEndpoints(app.queue, nil)
			opts := &riverui.HandlerOpts{
				Endpoints: endpoints,
				Logger:    app.logger,
				Prefix:    "/riverui",
			}

			h, err := riverui.NewHandler(opts)
			if err != nil {
				panic(err)
			}

			h.Start(context.Background())
			r.Handle("/riverui/*", h)
		})
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.NotFound(app.notFound)
		r.MethodNotAllowed(app.methodNotAllowed)

		r.Use(app.authenticate, app.reqLogger)

		if app.dev {
			r.Route("/datalake", func(r chi.Router) {
				r.Get("/processos", app.handleDatalakeProcessos)
				r.Get("/processos/unidades", app.handleDatalakeUnidadesProcessos)

				r.Get("/servidores/{cpf}", app.handleDatalakeServidor)
			})
		}

		r.Route("/usuarios", func(r chi.Router) {
			r.Use(
				app.requireAuth,
				app.requirePapel(auth.PapelGestor, auth.PapelSubsecretario),
			)

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
				r.Get("/analista/processo", app.handleAnalistaProcessoAtribuido)
			})
		})

		r.Route("/processos", func(r chi.Router) {
			r.Use(
				app.requireAuth,
				app.requirePapel(auth.PapelGestor, auth.PapelSubsecretario),
			)

			r.Get("/", app.handleProcessoList)
			r.Post("/", app.handleProcessoCreate)
			r.Get("/{processoID}", app.handleProcessoDetail)
			r.Get("/{processoID}/documentos", app.handleProcessoDetailDocumentos)
		})

		r.Route("/aposentadoria", func(r chi.Router) {
			r.Use(
				app.requireAuth,
				app.requirePapel(auth.PapelGestor, auth.PapelSubsecretario, auth.PapelAnalista),
			)

			r.Get("/", app.handleProcessoAposentadoriaList)
			r.Get("/{paID}", app.handleProcessoAposentadoriaDetail)
			r.Get("/{paID}/historico", app.handleProcessoAposentadoriaHistorico)
			r.Post("/{paID}/prioridade", app.handleProcessoAposentadoriaSolicitarPrioridade)
			r.Get("/{paID}/preview", app.handleAposentadoriaPreview)
			r.Post("/{paID}/leitura-invalida", app.handleProcessoAposentadoriaLeituraInvalida)

			r.Group(func(r chi.Router) {
				r.Use(app.requirePapel(auth.PapelGestor, auth.PapelSubsecretario))
				r.Post("/recalcular-scores", app.handleRecalcularScores)
			})
		})

		r.Route("/analistas", func(r chi.Router) {
			r.Use(
				app.requireAuth,
				app.requirePapel(auth.PapelGestor, auth.PapelSubsecretario),
			)

			r.Get("/", app.handleAnalistaList)
		})

		r.Route("/unidades", func(r chi.Router) {
			r.Use(
				app.requireAuth,
				app.requirePapel(auth.PapelGestor, auth.PapelSubsecretario),
			)

			r.Get("/", app.handleUnidadeList)
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/entrar", app.handleAuthEntrar)
			r.Get("/token", app.handleAuthTokenInfo)
			r.Post("/cadastrar", app.handleAuthCadastrar)

			r.Post("/recuperar-senha", app.handleAuthRecuperarSenha)
			r.Post("/redefinir-senha", app.handleAuthRedefinirSenha)

			r.Group(func(r chi.Router) {
				r.Use(app.requireAuth)

				r.Get("/me", app.handleAuthUsuarioAtual)
				r.Get("/me/analista", app.handleAuthAnalistaAtual)
				r.Post("/alterar-senha", app.handleAuthAlterarSenha)
			})
		})

		r.Route("/solicitacoes-prioridade", func(r chi.Router) {
			r.Use(
				app.requireAuth,
				app.requirePapel(auth.PapelSubsecretario),
			)

			r.Get("/", app.handleSolicitacoesPrioridadeList)
			r.Get("/{spID}", app.handleSolicitacoesPrioridadeDetail)
			r.Post("/{spID}/aprovar", app.handleSolicitacoesPrioridadeAprovar)
			r.Post("/{spID}/negar", app.handleSolicitacoesPrioridadeNegar)
		})

		r.Group(func(r chi.Router) {
			r.Use(app.requireAuth)
			r.Get("/meu-processo", app.handleMeuProcessoAtribuido)
		})
	})

	return r
}
