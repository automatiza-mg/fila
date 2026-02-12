package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/pagination"
	"github.com/automatiza-mg/fila/internal/processos"
	"github.com/google/uuid"
)

type ProcessoCreateRequest struct {
	Numero string `json:"numero"`
}

func (app *application) handleProcessoCreate(w http.ResponseWriter, r *http.Request) {
	var input ProcessoCreateRequest
	err := app.decodeJSON(w, r, &input)
	if err != nil {
		app.decodeError(w, r, err)
		return
	}

	p, err := app.processos.CreateProcesso(r.Context(), input.Numero)
	if err != nil {
		switch {
		case errors.Is(err, processos.ErrProcessoExists):
			app.writeError(w, http.StatusConflict, "O processo informado já existe")
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.writeJSON(w, http.StatusCreated, p)
}

func (app *application) handleProcessoList(w http.ResponseWriter, r *http.Request) {
	params := pagination.ParseQuery(r)

	result, err := app.processos.ListProcessos(r.Context(), processos.ListProcessosParams{
		Numero: r.URL.Query().Get("numero"),
		Page:   params.Page,
		Limit:  params.Limit,
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, result)
}

func (app *application) handleProcessoDetail(w http.ResponseWriter, r *http.Request) {
	processoID, err := uuid.Parse(r.PathValue("processoID"))
	if err != nil {
		app.notFound(w, r)
		return
	}

	p, err := app.processos.GetProcesso(r.Context(), processoID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.writeJSON(w, http.StatusOK, p)
}

func (app *application) handleProcessoDetailDocumentos(w http.ResponseWriter, r *http.Request) {
	processoID, err := uuid.Parse(r.PathValue("processoID"))
	if err != nil {
		app.notFound(w, r)
		return
	}

	p, err := app.processos.GetProcesso(r.Context(), processoID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	dd, err := app.processos.ListDocumentos(r.Context(), p.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, dd)
}

func (app *application) handleProcessoAnalyze(w http.ResponseWriter, r *http.Request) {
	processoID, err := uuid.Parse(r.PathValue("processoID"))
	if err != nil {
		app.notFound(w, r)
		return
	}

	p, err := app.processos.GetProcesso(r.Context(), processoID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	go func() {
		// TODO: Realizar a análise no River.
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if err := app.processos.Analyze(ctx, p.ID); err != nil {
			app.logger.Error("Análise de processo falhou", slog.Any("err", err))
		}
	}()

	w.WriteHeader(http.StatusAccepted)
}
