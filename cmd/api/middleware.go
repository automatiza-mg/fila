package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/automatiza-mg/fila/internal/auth"
)

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			ctx := app.setAuth(r.Context(), auth.Anonymous)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			app.writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Message: "Header 'Authorization' é inválido",
			})
			return
		}

		token := parts[1]
		usuario, err := app.auth.GetTokenOwner(r.Context(), token, auth.EscopoAuth)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrInvalidToken):
				app.writeJSON(w, http.StatusUnauthorized, ErrorResponse{
					Message: "O token informado é inválido ou expirou",
				})
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
			app.writeJSON(w, http.StatusUnauthorized, ErrorResponse{
				Message: "Você deve estar autenticado para acessar esse recurso",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
