package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/fila"
	"github.com/automatiza-mg/fila/internal/pagination"
	"github.com/automatiza-mg/fila/internal/processos"
	"github.com/automatiza-mg/fila/internal/validator"
)

func (app *application) handleRecalcularScores(w http.ResponseWriter, r *http.Request) {
	err := app.fila.EnqueueRecalcularScores(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, map[string]string{
		"mensagem": "Recálculo de scores enfileirado com sucesso",
	})
}

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

func (app *application) handleMeuHistorico(w http.ResponseWriter, r *http.Request) {
	usuario := app.getAuth(r.Context())
	params := pagination.ParseQuery(r)
	numero := r.URL.Query().Get("numero")

	result, err := app.fila.ListHistoricoAnalista(r.Context(), usuario.ID, numero, params.Page, params.Limit)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, result)
}

// getProcessoAposentadoriaFromRequest carrega o ProcessoAposentadoria pelo paID
// da rota e verifica se o usuário autenticado tem acesso. Gestores e
// subsecretários acessam qualquer processo; analistas só o que estiver
// atribuído a eles. Retorna nil quando a resposta já foi escrita.
func (app *application) getProcessoAposentadoriaFromRequest(w http.ResponseWriter, r *http.Request) *fila.ProcessoAposentadoria {
	paID, err := app.intParam(r, "paID")
	if err != nil || paID < 1 {
		app.notFound(w, r)
		return nil
	}

	pa, err := app.fila.GetProcessoAposentadoria(r.Context(), paID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return nil
	}

	usuario := app.getAuth(r.Context())
	if usuario.IsAnalista() {
		if pa.AnalistaID == nil || *pa.AnalistaID != usuario.ID {
			app.writeError(w, http.StatusForbidden, "Você não possui permissão para acessar este processo")
			return nil
		}
	}

	return pa
}

func (app *application) handleAposentadoriaPreview(w http.ResponseWriter, r *http.Request) {
	pa := app.getProcessoAposentadoriaFromRequest(w, r)
	if pa == nil {
		return
	}

	preview, err := app.processos.GetPreview(r.Context(), pa.ProcessoID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		case errors.Is(err, processos.ErrPreviewUnavailable):
			app.writeError(w, http.StatusNotFound, "Preview ainda não disponível para este processo")
		default:
			app.serverError(w, r, err)
		}
		return
	}
	defer preview.Body.Close()

	w.Header().Set("Content-Type", preview.ContentType)
	io.Copy(w, preview.Body)
}

type LeituraInvalidaRequest struct {
	Motivo string `json:"motivo"`

	validator.Validator `json:"-"`
}

func (app *application) handleProcessoAposentadoriaLeituraInvalida(w http.ResponseWriter, r *http.Request) {
	paID, err := app.intParam(r, "paID")
	if err != nil || paID < 1 {
		app.notFound(w, r)
		return
	}

	var input LeituraInvalidaRequest
	err = app.decodeJSON(w, r, &input)
	if err != nil {
		app.decodeError(w, r, err)
		return
	}

	input.Check(validator.NotBlank(input.Motivo), "motivo", "Campo obrigatório")
	if !input.Valid() {
		app.validationFailed(w, r, input.FieldErrors)
		return
	}

	usuario := app.getAuth(r.Context())

	err = app.fila.MarcarLeituraInvalida(r.Context(), fila.MarcarLeituraInvalidaParams{
		AnalistaID: usuario.ID,
		ProcessoID: paID,
		Motivo:     input.Motivo,
	})
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		case errors.Is(err, fila.ErrNotAssigned):
			app.writeError(w, http.StatusForbidden, "Você não possui permissão para alterar este processo")
		case errors.Is(err, fila.ErrInvalidStatus):
			app.writeError(w, http.StatusConflict, "O processo não está no status esperado para esta ação")
		default:
			app.serverError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) handleProcessoAposentadoriaRegistrarPublicacao(w http.ResponseWriter, r *http.Request) {
	paID, err := app.intParam(r, "paID")
	if err != nil || paID < 1 {
		app.notFound(w, r)
		return
	}

	usuario := app.getAuth(r.Context())

	err = app.fila.RegistrarPublicacao(r.Context(), paID, usuario.ID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		case errors.Is(err, fila.ErrNotAssigned):
			app.writeError(w, http.StatusForbidden, "Você não possui permissão para alterar este processo")
		case errors.Is(err, fila.ErrInvalidStatus):
			app.writeError(w, http.StatusConflict, "O processo não está no status esperado para esta ação")
		default:
			app.serverError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
