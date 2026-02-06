package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/automatiza-mg/fila/internal/datalake"
)

func (app *application) handleDatalakeProcessos(w http.ResponseWriter, r *http.Request) {
	unidade := r.URL.Query().Get("unidade")
	if unidade == "" {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "O parâmetro 'unidade' deve ser informado",
		})
		return
	}

	key := fmt.Sprintf("fila:datalake:processos:%s", unidade)
	b, err := app.cache.Remember(r.Context(), key, 2*time.Hour, func() ([]byte, error) {
		processos, _, err := app.datalake.ListProcessosAbertos(r.Context(), unidade)
		if err != nil {
			return nil, err
		}
		return json.Marshal(processos)
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	var processos []datalake.Processo
	err = json.Unmarshal(b, &processos)
	if err != nil {
		_ = app.cache.Del(r.Context(), key)
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, processos)
}

func (app *application) handleDatalakeUnidadesProcessos(w http.ResponseWriter, r *http.Request) {
	unidades, err := app.datalake.ListUnidadesDisponiveis(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, unidades)
}

func (app *application) handleDatalakeServidor(w http.ResponseWriter, r *http.Request) {
	cpf := r.PathValue("cpf")

	key := fmt.Sprintf("fila:datalake:servidores:%s", cpf)
	b, err := app.cache.Remember(r.Context(), key, 24*time.Hour, func() ([]byte, error) {
		servidor, err := app.datalake.GetServidorByCPF(r.Context(), cpf)
		if err != nil {
			return nil, err
		}
		return json.Marshal(servidor)
	})
	if err != nil {
		switch {
		case errors.Is(err, datalake.ErrNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	var servidor datalake.Servidor
	err = json.Unmarshal(b, &servidor)
	if err != nil {
		_ = app.cache.Del(context.Background(), key)
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, servidor)
}
