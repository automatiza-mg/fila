package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/diligencias"
	"github.com/automatiza-mg/fila/internal/validator"
)

// DiligenciaItemRequest representa um item individual de diligência enviado
// pelo cliente.
type DiligenciaItemRequest struct {
	Tipo          string   `json:"tipo"`
	Subcategorias []string `json:"subcategorias"`
	Detalhe       string   `json:"detalhe"`
}

// SalvarDiligenciaRequest representa o corpo da requisição para substituir
// os itens de um rascunho de diligência.
type SalvarDiligenciaRequest struct {
	Itens []DiligenciaItemRequest `json:"itens"`

	validator.Validator `json:"-"`
}

// handleDiligenciaRascunhoGet retorna o rascunho ativo de diligência para o
// analista no processo informado, criando um novo caso não exista.
func (app *application) handleDiligenciaRascunhoGet(w http.ResponseWriter, r *http.Request) {
	pa := app.getProcessoAposentadoriaFromRequest(w, r)
	if pa == nil {
		return
	}

	usuario := app.getAuth(r.Context())

	sd, err := app.diligencias.GetOrCreateRascunho(r.Context(), pa.ID, usuario.ID)
	if err != nil {
		app.handleDiligenciaError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, sd)
}

// handleDiligenciaRascunhoSalvar substitui o conjunto de itens do rascunho
// ativo. Aceita lista vazia, permitindo que o analista remova todos os itens.
// A validação de lista não vazia acontece apenas no envio da diligência.
func (app *application) handleDiligenciaRascunhoSalvar(w http.ResponseWriter, r *http.Request) {
	pa := app.getProcessoAposentadoriaFromRequest(w, r)
	if pa == nil {
		return
	}

	var input SalvarDiligenciaRequest
	if err := app.decodeJSON(w, r, &input); err != nil {
		app.decodeError(w, r, err)
		return
	}

	for i, it := range input.Itens {
		input.Check(validator.NotBlank(it.Tipo), fmt.Sprintf("itens[%d].tipo", i), "Campo obrigatório")
		input.Check(
			len(it.Subcategorias) > 0 || validator.NotBlank(it.Detalhe),
			fmt.Sprintf("itens[%d]", i),
			"Deve conter ao menos uma subcategoria ou detalhe",
		)
	}
	if !input.Valid() {
		app.validationFailed(w, r, input.FieldErrors)
		return
	}

	usuario := app.getAuth(r.Context())

	rascunho, err := app.diligencias.GetOrCreateRascunho(r.Context(), pa.ID, usuario.ID)
	if err != nil {
		app.handleDiligenciaError(w, r, err)
		return
	}

	itens := make([]diligencias.NovoItem, 0, len(input.Itens))
	for _, it := range input.Itens {
		itens = append(itens, diligencias.NovoItem{
			Tipo:          it.Tipo,
			Subcategorias: it.Subcategorias,
			Detalhe:       it.Detalhe,
		})
	}

	sd, err := app.diligencias.SalvarRascunho(r.Context(), diligencias.SalvarRascunhoParams{
		SolicitacaoID: rascunho.ID,
		AnalistaID:    usuario.ID,
		Itens:         itens,
	})
	if err != nil {
		app.handleDiligenciaError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, sd)
}

// handleDiligenciaRascunhoDescartar exclui o rascunho ativo de diligência.
func (app *application) handleDiligenciaRascunhoDescartar(w http.ResponseWriter, r *http.Request) {
	pa := app.getProcessoAposentadoriaFromRequest(w, r)
	if pa == nil {
		return
	}

	usuario := app.getAuth(r.Context())

	rascunho, err := app.diligencias.GetRascunho(r.Context(), pa.ID, usuario.ID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		app.serverError(w, r, err)
		return
	}

	if err := app.diligencias.DescartarRascunho(r.Context(), rascunho.ID, usuario.ID); err != nil {
		app.handleDiligenciaError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleDiligenciaRascunhoEnviar finaliza e envia o rascunho ativo de
// diligência, transitando o processo para EM_DILIGENCIA.
func (app *application) handleDiligenciaRascunhoEnviar(w http.ResponseWriter, r *http.Request) {
	pa := app.getProcessoAposentadoriaFromRequest(w, r)
	if pa == nil {
		return
	}

	usuario := app.getAuth(r.Context())

	rascunho, err := app.diligencias.GetRascunho(r.Context(), pa.ID, usuario.ID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			app.writeError(w, http.StatusNotFound, "Nenhum rascunho de diligência encontrado para este processo")
			return
		}
		app.serverError(w, r, err)
		return
	}

	sent, err := app.diligencias.EnviarDiligencia(r.Context(), rascunho.ID, usuario.ID)
	if err != nil {
		app.handleDiligenciaError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, sent)
}

// handleDiligenciaList retorna a lista de solicitações de diligência enviadas
// para o processo informado.
func (app *application) handleDiligenciaList(w http.ResponseWriter, r *http.Request) {
	pa := app.getProcessoAposentadoriaFromRequest(w, r)
	if pa == nil {
		return
	}

	list, err := app.diligencias.ListSolicitacoesEnviadas(r.Context(), pa.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, list)
}

// handleDiligenciaError traduz erros do service de diligências em respostas
// HTTP adequadas.
func (app *application) handleDiligenciaError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, database.ErrNotFound):
		app.notFound(w, r)
	case errors.Is(err, diligencias.ErrNotAssigned):
		app.writeError(w, http.StatusForbidden, "Você não possui permissão para alterar este processo")
	case errors.Is(err, diligencias.ErrInvalidStatus):
		app.writeError(w, http.StatusConflict, "O processo não está no status esperado para esta ação")
	case errors.Is(err, diligencias.ErrAlreadySent):
		app.writeError(w, http.StatusConflict, "A diligência já foi enviada e não pode ser modificada")
	case errors.Is(err, diligencias.ErrDraftEmpty):
		app.writeError(w, http.StatusConflict, "O rascunho não possui itens para envio")
	default:
		app.serverError(w, r, err)
	}
}
