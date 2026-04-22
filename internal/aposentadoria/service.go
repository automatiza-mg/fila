package aposentadoria

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/datalake"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/singleflight"
)

const (
	// Tempo de duração de uma chave/valor no cache.
	cacheTTL = 2 * time.Hour
)

var (
	// ErrNoProcesso é o erro retornado quando um CPF é consultado sem possuir
	// um processo de aposentadoria no banco de dados.
	ErrNoProcesso = errors.New("cpf does not have a processo")
)

type Service struct {
	pool     *pgxpool.Pool
	store    *database.Store
	datalake *datalake.DataLake
	cache    cache.Cache
	sg       singleflight.Group
}

func New(pool *pgxpool.Pool, dl *datalake.DataLake, cache cache.Cache) *Service {
	return &Service{
		pool:     pool,
		store:    database.New(pool),
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

// ttlUntilNextMidnight retorna a duração até a próxima meia-noite no fuso
// horário de São Paulo, alinhada com a atualização diária do datalake.
// Aplica um piso mínimo de 1 hora para evitar TTLs muito curtos próximos
// à virada do dia.
func ttlUntilNextMidnight() time.Duration {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	nextMidnight := time.Date(
		now.Year(), now.Month(), now.Day()+1,
		0, 0, 0, 0, loc,
	)
	ttl := time.Until(nextMidnight)
	if ttl < time.Hour {
		ttl = time.Hour
	}
	return ttl
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

// HasProcessoByCPF verifica se um determinado CPF possui um processo de
// aposentadoria cadastrado no sistema.
func (s *Service) HasProcessoByCPF(ctx context.Context, cpf string) (bool, error) {
	return s.store.HasProcessoAposentadoria(ctx, cpf)
}

// GetServidor retorna os dados de um servidor pelo CPF informado, utilizando
// cache para reduzir consultas ao datalake. O TTL do cache se alinha com a
// próxima meia-noite no fuso de São Paulo, refletindo a atualização diária
// do datalake. A verificação de existência de um processo de aposentadoria
// para o CPF é responsabilidade do chamador.
func (s *Service) GetServidor(ctx context.Context, cpf string) (*datalake.Servidor, error) {
	key := fmt.Sprintf("fila:datalake:servidor:%s", cpf)

	b, err := s.remember(ctx, key, ttlUntilNextMidnight(), func() ([]byte, error) {
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
