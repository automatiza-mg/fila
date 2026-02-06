package processos

import (
	"context"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Analyzer struct {
	pool  *pgxpool.Pool
	store *database.Store
}

func (a *Analyzer) ListAnalistas(ctx context.Context) ([]Analista, error) {
	usuarios, _, err := a.store.ListUsuarios(ctx, database.ListUsuariosParams{
		Papel: database.PapelAnalista,
	})
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(usuarios))
	for i, usuario := range usuarios {
		ids[i] = usuario.ID
	}

	analistasMap, err := a.store.GetAnalistasMap(ctx, ids)
	if err != nil {
		return nil, err
	}

	analistas := make([]Analista, 0)
	for _, usuario := range usuarios {
		_, ok := analistasMap[usuario.ID]
		if ok {
			analistas = append(analistas, Analista{
				ID:   usuario.ID,
				CPF:  usuario.CPF,
				Nome: usuario.Nome,
			})
		}
	}

	return analistas, nil
}
