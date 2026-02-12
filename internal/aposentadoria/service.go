package aposentadoria

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/datalake"
	"golang.org/x/sync/singleflight"
)

const (
	// Tempo de duração de uma chave/valor no cache.
	cacheTTL = 2 * time.Hour
)

type Service struct {
	datalake *datalake.DataLake
	cache    cache.Cache
	sg       singleflight.Group
}

func New(dl *datalake.DataLake, cache cache.Cache) *Service {
	return &Service{
		datalake: dl,
		cache:    cache,
	}
}

// Adiciona uma proteção contra Thundering Herd ao cache usando [singleflight.Group].
func (s *Service) remember(ctx context.Context, key string, ttl time.Duration, fn func() ([]byte, error)) ([]byte, error) {
	v, err, _ := s.sg.Do(key, func() (any, error) {
		return s.cache.Remember(ctx, key, ttl, fn)
	})
	if err != nil {
		return nil, err
	}
	return v.([]byte), nil
}

// ListProcessosAbertos retorna uma lista de processos abertos na unidade
// do SEI especificada.
func (s *Service) ListProcessosAbertos(ctx context.Context, unidade string) ([]datalake.Processo, error) {
	key := fmt.Sprintf("fila:datalake:processos:%s", unidade)

	b, err := s.remember(ctx, key, cacheTTL, func() ([]byte, error) {
		pp, _, err := s.datalake.ListProcessosAbertos(ctx, unidade)
		if err != nil {
			return nil, err
		}
		return json.Marshal(pp)
	})
	if err != nil {
		return nil, err
	}

	var pp []datalake.Processo
	err = json.Unmarshal(b, &pp)
	if err != nil {
		delErr := s.cache.Del(ctx, key)
		if delErr != nil {
			return nil, errors.Join(delErr, err)
		}
		return nil, err
	}

	return pp, nil
}

// ListUnidadesDisponiveis retorna a lista de unidades disponíveis para consultar
// os processos abertos.
func (s *Service) ListUnidadesDisponiveis(ctx context.Context) ([]string, error) {
	key := "fila:datalake:unidades"

	b, err := s.remember(ctx, key, cacheTTL, func() ([]byte, error) {
		uu, err := s.datalake.ListUnidadesDisponiveis(ctx)
		if err != nil {
			return nil, err
		}
		return json.Marshal(uu)
	})
	if err != nil {
		return nil, err
	}

	var uu []string
	err = json.Unmarshal(b, &uu)
	if err != nil {
		delErr := s.cache.Del(ctx, key)
		if delErr != nil {
			return nil, errors.Join(delErr, err)
		}
		return nil, err
	}

	return uu, nil
}

// GetServidorByCPF retorna os dados de um servidor pelo CPF informado.
func (s *Service) GetServidorByCPF(ctx context.Context, cpf string) (*datalake.Servidor, error) {
	key := fmt.Sprintf("fila:datalake:servidor:%s", cpf)

	b, err := s.remember(ctx, key, cacheTTL, func() ([]byte, error) {
		serv, err := s.datalake.GetServidorByCPF(ctx, cpf)
		if err != nil {
			return nil, err
		}
		return json.Marshal(serv)
	})
	if err != nil {
		return nil, err
	}

	var serv datalake.Servidor
	err = json.Unmarshal(b, &serv)
	if err != nil {
		delErr := s.cache.Del(ctx, key)
		if delErr != nil {
			return nil, errors.Join(delErr, err)
		}
		return nil, err
	}

	return &serv, nil
}
