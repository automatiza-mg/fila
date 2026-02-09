package main

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/automatiza-mg/fila/internal/auth"
)

type loggerWriter struct {
	status int
	http.ResponseWriter
}

func (w *loggerWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggerWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (app *application) reqLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := &loggerWriter{ResponseWriter: w, status: http.StatusOK}

		t := time.Now()
		defer func() {
			app.logger.Info(
				"Requisição HTTP",
				slog.Int("status", lw.status),
				slog.String("proto", r.Proto),
				slog.String("method", r.Method),
				slog.String("uri", r.URL.RequestURI()),
				slog.Duration("duration", time.Since(t)),
			)
		}()

		next.ServeHTTP(lw, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			ctx := app.setAuth(r.Context(), auth.Anonymous)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			app.writeError(w, http.StatusUnauthorized, "Header 'Authorization' é inválido")
			return
		}

		token := strings.TrimSpace(parts[1])
		if token == "" {
			app.writeError(w, http.StatusUnauthorized, "Token ausente")
			return
		}

		usuario, err := app.auth.GetTokenOwner(r.Context(), token, auth.EscopoAuth)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrInvalidToken):
				app.tokenError(w, r)
			default:
				app.serverError(w, r, err)
			}
			return
		}

		ctx := app.setAuth(r.Context(), usuario)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		usuario := app.getAuth(r.Context())
		if usuario.IsAnonymous() {
			w.Header().Set("WWW-Authenticate", "Bearer")
			app.writeError(w, http.StatusUnauthorized, "Você deve estar autenticado para acessar esse recurso")
			return
		}

		// Remove a possibilidade de caching dos dados servidos pela API.
		// Rotas protegidas tem alta probabilidade de retornas dados sensíveis (PII, Processos SEI, etc).
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		next.ServeHTTP(w, r)
	})
}
