package main

import (
	"context"
	"net/http"

	"github.com/automatiza-mg/fila/internal/sei"
)

// Retorna um mapa de unidades SEI onde o campo IdUnidade é a chave.
func (app *application) getUnidadesMap(ctx context.Context) (map[string]sei.Unidade, error) {
	unidades, err := app.fila.ListUnidadesAnalistas(ctx)
	if err != nil {
		return nil, err
	}

	unidadesMap := make(map[string]sei.Unidade, len(unidades))
	for _, unidade := range unidades {
		unidadesMap[unidade.IdUnidade] = unidade
	}
	return unidadesMap, nil
}

// Lista as unidades do SEI disponíveis para os analistas.
func (app *application) handleUnidadeList(w http.ResponseWriter, r *http.Request) {
	unidades, err := app.fila.ListUnidadesAnalistas(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, unidades)
}
