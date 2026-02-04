package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
)

type Usuario struct {
	ID              int64     `json:"id"`
	Nome            string    `json:"nome"`
	CPF             string    `json:"cpf"`
	Email           string    `json:"email"`
	EmailVerificado bool      `json:"email_verificado"`
	Papel           *string   `json:"papel"`
	CriadoEm        time.Time `json:"criado_em"`
	AtualizadoEm    time.Time `json:"atualizado_em"`
}

func mapUsuario(record *database.Usuario) *Usuario {
	return &Usuario{
		ID:              record.ID,
		Nome:            record.Nome,
		CPF:             record.CPF,
		Email:           record.Email,
		EmailVerificado: record.EmailVerificado,
		Papel:           database.Ptr(record.Papel),
		CriadoEm:        record.CriadoEm,
		AtualizadoEm:    record.AtualizadoEm,
	}
}

func (app *application) handleUsuarioList(w http.ResponseWriter, r *http.Request) {
	records, err := app.store.ListUsuarios(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	usuarios := make([]*Usuario, len(records))
	for i, record := range records {
		usuarios[i] = mapUsuario(record)
	}

	app.writeJSON(w, http.StatusOK, usuarios)
}

func (app *application) handleUsuarioDetail(w http.ResponseWriter, r *http.Request) {
	usuarioID, err := app.intParam(r, "usuarioID")
	if err != nil {
		app.notFound(w, r)
		return
	}

	record, err := app.store.GetUsuario(r.Context(), usuarioID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.writeJSON(w, http.StatusOK, mapUsuario(record))
}
