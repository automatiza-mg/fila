package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/fila"
	"github.com/automatiza-mg/fila/internal/pagination"
	"github.com/automatiza-mg/fila/internal/validator"
)

func (app *application) handleProcessoAposentadoriaList(w http.ResponseWriter, r *http.Request) {
	params := pagination.ParseQuery(r)

	result, err := app.fila.ListProcesso(r.Context(), fila.ListProcessoAposentadoriaParams{
		Numero: r.URL.Query().Get("numero"),
		Status: r.URL.Query().Get("status"),
		Page:   params.Page,
		Limit:  params.Limit,
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, result)
}

func (app *application) handleProcessoAposentadoriaDetail(w http.ResponseWriter, r *http.Request) {
	paID, err := app.intParam(r, "paID")
	if err != nil || paID < 1 {
		app.notFound(w, r)
		return
	}

	pa, err := app.fila.GetProcessoAposentadoria(r.Context(), paID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.writeJSON(w, http.StatusOK, pa)
}

type SolicitarPrioridadeRequest struct {
	Justificativa string `json:"justificativa"`

	validator.Validator `json:"-"`
}

func (app *application) handleProcessoAposentadoriaSolicitarPrioridade(w http.ResponseWriter, r *http.Request) {
	paID, err := app.intParam(r, "paID")
	if err != nil || paID < 1 {
		app.notFound(w, r)
		return
	}

	pa, err := app.fila.GetProcessoAposentadoria(r.Context(), paID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	var input SolicitarPrioridadeRequest
	err = app.decodeJSON(w, r, &input)
	if err != nil {
		app.decodeError(w, r, err)
		return
	}

	input.Check(validator.NotBlank(input.Justificativa), "justificativa", "Deve ser preenchido")
	if !input.Valid() {
		app.validationFailed(w, r, input.FieldErrors)
		return
	}

	sp, err := app.fila.CreateSolicitacaoPrioridade(r.Context(), fila.SolicitarPrioridadeParams{
		ProcessoAposentadoriaID: pa.ID,
		UsuarioID:               app.getAuth(r.Context()).ID,
		Justificativa:           input.Justificativa,
		SolicitacaoURL: func(numero string) string {
			q := make(url.Values)
			q.Set("numero", numero)
			return fmt.Sprintf("%s/processos/prioridades?%s", app.cfg.ClientURL, q.Encode())
		},
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, sp)
}

func (app *application) handleProcessoAposentadoriaHistorico(w http.ResponseWriter, r *http.Request) {
	paID, err := app.intParam(r, "paID")
	if err != nil || paID < 1 {
		app.notFound(w, r)
		return
	}

	historico, err := app.fila.ListHistorico(r.Context(), paID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if historico == nil {
		historico = make([]*fila.HistoricoStatusProcesso, 0)
	}

	app.writeJSON(w, http.StatusOK, historico)
}

func (app *application) handleAnalistaProcessoAtribuido(w http.ResponseWriter, r *http.Request) {
	usuario := app.getUsuario(r.Context())

	pa, err := app.fila.GetProcessoAtribuido(r.Context(), usuario.ID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.writeJSON(w, http.StatusOK, pa)
}

func (app *application) handleMeuProcessoAtribuido(w http.ResponseWriter, r *http.Request) {
	usuario := app.getAuth(r.Context())

	pa, err := app.fila.GetProcessoAtribuido(r.Context(), usuario.ID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.writeJSON(w, http.StatusOK, pa)
}
