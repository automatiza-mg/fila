package fila

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/automatiza-mg/fila/internal/sei"
)

type UnidadeSei struct {
	ID        string `json:"id"`
	Sigla     string `json:"sigla"`
	Descricao string `json:"descricao"`
}

// ListUnidadesAnalistas retorna as unidades do SEI reservadas aos analistas de processos de aposentadoria (SEPLAG/APXX).
func (s *Service) ListUnidadesAnalistas(ctx context.Context) ([]UnidadeSei, error) {
	key := "fila:sei:unidades"

	b, err := s.cache.Remember(ctx, key, 24*time.Hour, func() ([]byte, error) {
		resp, err := s.sei.ListarUnidades(ctx)
		if err != nil {
			return nil, err
		}

		// Filtra as unidades de aposentadoria.
		unidades := make([]sei.Unidade, 0)
		for _, unidade := range resp.Parametros.Items {
			if len(unidade.Sigla) == 11 && strings.HasPrefix(unidade.Sigla, "SEPLAG/AP") {
				unidades = append(unidades, unidade)
			}
		}

		return json.Marshal(unidades)
	})

	if err != nil {
		return nil, err
	}

	var items []sei.Unidade
	if err := json.Unmarshal(b, &items); err != nil {
		return nil, err
	}

	unidades := make([]UnidadeSei, len(items))
	for i, item := range items {
		unidades[i] = UnidadeSei{
			ID:        item.IdUnidade,
			Sigla:     item.Sigla,
			Descricao: item.Descricao,
		}
	}
	return unidades, nil
}
