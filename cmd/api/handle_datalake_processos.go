package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/automatiza-mg/fila/internal/datalake"
)

func (app *application) handleDatalakeProcessos(w http.ResponseWriter, r *http.Request) {
	unidade := r.URL.Query().Get("unidade")
	if unidade == "" {
		app.writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Message: "O parâmtro 'unidade' deve ser informado",
		})
		return
	}

	key := fmt.Sprintf("fila:datalake:processos:%s", unidade)
	b, err := app.cache.Remember(r.Context(), key, 2*time.Hour, func() ([]byte, error) {
		processos, _, err := app.dl.ListProcessosAbertos(r.Context(), unidade)
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
	unidades, err := app.dl.ListUnidadesDisponiveis(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, unidades)
}
