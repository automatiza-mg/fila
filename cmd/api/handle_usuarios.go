package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/validator"
)

func (app *application) loadUsuario(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		usuarioID, err := app.intParam(r, "usuarioID")
		if err != nil || usuarioID < 1 {
			app.notFound(w, r)
			return
		}

		usuario, err := app.store.GetUsuario(r.Context(), usuarioID)
		if err != nil {
			switch {
			case errors.Is(err, database.ErrNotFound):
				app.notFound(w, r)
			default:
				app.serverError(w, r, err)
			}
			return
		}

		// Carrega dados de analista
		if usuario.HasPapel(database.PapelAnalista) {
			usuario.Analista, err = app.store.GetAnalista(r.Context(), usuario.ID)
			if err != nil && !errors.Is(err, database.ErrNotFound) {
				app.serverError(w, r, err)
				return
			}
		}

		ctx := app.setUsuario(r.Context(), usuario)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) handleUsuarioList(w http.ResponseWriter, r *http.Request) {
	papel := r.URL.Query().Get("papel")

	usuarios, _, err := app.store.ListUsuarios(r.Context(), database.ListUsuariosParams{
		Papel: strings.ToUpper(papel),
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	analistaIDs := make([]int64, 0)
	for _, usuario := range usuarios {
		if usuario.HasPapel(database.PapelAnalista) {
			analistaIDs = append(analistaIDs, usuario.ID)
		}
	}

	if len(analistaIDs) > 0 {
		analistas, err := app.store.GetAnalistasByUsuarioIDs(r.Context(), analistaIDs)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		for _, u := range usuarios {
			if a, ok := analistas[u.ID]; ok {
				u.Analista = a
			}
		}
	}

	app.writeJSON(w, http.StatusOK, usuarios)
}

// Lista de papeis permitidos
var papeisAllowed = []string{
	database.PapelAnalista,
	database.PapelGestor,
	database.PapelSubsecretario,
}

type usuarioCreateRequest struct {
	Nome  string `json:"nome"`
	CPF   string `json:"cpf"`
	Email string `json:"email"`
	Papel string `json:"papel"`

	validator.Validator `json:"-"`
}

func (app *application) handleUsuarioCreate(w http.ResponseWriter, r *http.Request) {
	var input usuarioCreateRequest
	err := app.decodeJSON(w, r, &input)
	if err != nil {
		app.decodeError(w, r, err)
		return
	}

	input.Check(validator.NotBlank(input.Nome), "nome", "Campo obrigatório")
	input.Check(validator.MaxLength(input.Nome, 255), "nome", "Deve possuir até 255 caracteres")
	input.Check(validator.NotBlank(input.CPF), "cpf", "Campo obrigatório")
	input.Check(validator.Matches(input.CPF, validator.CpfRX), "cpf", "Deve ser um CPF válido")
	input.Check(validator.NotBlank(input.Email), "email", "Campo obrigatório")
	input.Check(validator.MaxLength(input.Email, 255), "email", "Deve possuir até 255 caracteres")
	input.Check(validator.Matches(input.Email, validator.EmailRX), "email", "Deve ser um email válido")
	input.Check(validator.PermittedValue(input.Papel, papeisAllowed...), "papel", fmt.Sprintf("Deve ser um dos valores: %s", strings.Join(papeisAllowed, ", ")))
	if !input.Valid() {
		app.validationFailed(w, r, input.FieldErrors)
		return
	}

	usuario := &database.Usuario{
		Nome:  input.Nome,
		CPF:   input.CPF,
		Email: input.Email,
	}
	usuario.SetPapel(input.Papel)

	err = app.store.SaveUsuario(r.Context(), usuario)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrUsuarioCPFTaken):
			input.SetFieldError("cpf", "Valor já está em uso")
			app.validationFailed(w, r, input.FieldErrors)
		case errors.Is(err, database.ErrUsuarioEmailTaken):
			input.SetFieldError("email", "Valor já está em uso")
			app.validationFailed(w, r, input.FieldErrors)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.writeJSON(w, http.StatusCreated, usuario)
}

func (app *application) handleUsuarioDetail(w http.ResponseWriter, r *http.Request) {
	usuario := app.getUsuario(r.Context())
	app.writeJSON(w, http.StatusOK, usuario)
}
