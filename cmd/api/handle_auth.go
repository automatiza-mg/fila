package main

import (
	"net/http"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
)

func (app *application) handleAuthEntrar(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CPF   string `json:"cpf"`
		Senha string `json:"senha"`
	}
	err := app.decodeJSON(w, r, &input)
	if err != nil {
		app.decodeError(w, r, err)
		return
	}

	record, err := app.store.GetUsuarioByCPF(r.Context(), input.CPF)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	ok, err := record.CheckSenha(input.Senha)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if !ok {
		app.writeJSON(w, http.StatusUnauthorized, ErrorResponse{
			Message: "Credenciais inválidas",
		})
		return
	}

	token, err := app.store.CreateToken(r.Context(), record.ID, database.EscopoAuth, 24*time.Hour)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, token)
}

// Retorna os dados do usuário autenticado.
func (app *application) handleAuthUsuarioAtual(w http.ResponseWriter, r *http.Request) {
	usuario := app.getUsuario(r.Context())
	if usuario.IsAnonymous() {
		app.writeJSON(w, http.StatusUnauthorized, ErrorResponse{
			Message: "Você deve estar autenticado para acessar esse recurso",
		})
		return
	}

	app.writeJSON(w, http.StatusOK, mapUsuario(usuario))
}
