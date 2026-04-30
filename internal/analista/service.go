package analista

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/automatiza-mg/fila/internal/auth"
	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrInvalidUnidade é o erro retornado quando a unidade SEI informada não é válida.
	ErrInvalidUnidade = errors.New("invalid unidade id")
	// ErrInvalidPapel é o erro retornado quando o usuário não possui papel de analista.
	ErrInvalidPapel = errors.New("invalid usuario papel: expected ANALISTA")

	// AllowedOrgaos são os órgãos permitidos para o cadastro de analistas.
	AllowedOrgaos = []string{
		"SEPLAG",
		"SEE",
	}
)

// SeiClient define a interface para comunicação com o SEI.
type SeiClient interface {
	ListarUnidades(ctx context.Context) (*sei.ListarUnidadesResponse, error)
}

// Analista representa os dados complementares de um usuário analista.
type Analista struct {
	UsuarioID          int64      `json:"usuario_id"`
	Orgao              string     `json:"orgao"`
	SeiUnidadeID       string     `json:"sei_unidade_id"`
	SeiUnidadeSigla    string     `json:"sei_unidade_sigla"`
	Afastado           bool       `json:"afastado"`
	UltimaAtribuicaoEm *time.Time `json:"ultima_atribuicao_em"`
}

func mapAnalista(a *database.Analista) *Analista {
	return &Analista{
		UsuarioID:          a.UsuarioID,
		Orgao:              a.Orgao,
		SeiUnidadeID:       a.SEIUnidadeID,
		SeiUnidadeSigla:    a.SEIUnidadeSigla,
		Afastado:           a.Afastado,
		UltimaAtribuicaoEm: database.Ptr(a.UltimaAtribuicaoEm),
	}
}

// Service gerencia operações de analistas.
type Service struct {
	pool   *pgxpool.Pool
	store  *database.Store
	logger *slog.Logger
	sei    SeiClient
	cache  cache.Cache
}

// New cria uma nova instância de [Service].
func New(pool *pgxpool.Pool, logger *slog.Logger, sei SeiClient, cache cache.Cache) *Service {
	return &Service{
		pool:   pool,
		store:  database.New(pool),
		logger: logger.With(slog.String("service", "analistas")),
		sei:    sei,
		cache:  cache,
	}
}

// CreateAnalistaParams são os parâmetros para criação de um analista.
type CreateAnalistaParams struct {
	UsuarioID    int64
	SeiUnidadeID string
	Orgao        string
}

// CreateAnalista cadastra os dados de analista para um determinado usuário.
func (s *Service) CreateAnalista(ctx context.Context, params CreateAnalistaParams) (*Analista, error) {
	if !slices.Contains(AllowedOrgaos, params.Orgao) {
		return nil, fmt.Errorf("invalid orgao: %q", params.Orgao)
	}

	u, err := s.store.GetUsuario(ctx, params.UsuarioID)
	if err != nil {
		return nil, err
	}
	if u.Papel.V != auth.PapelAnalista {
		return nil, ErrInvalidPapel
	}

	unidadesMap, err := s.GetUnidadesMap(ctx)
	if err != nil {
		return nil, err
	}

	unidade, ok := unidadesMap[params.SeiUnidadeID]
	if !ok {
		return nil, ErrInvalidUnidade
	}

	record := &database.Analista{
		UsuarioID:       params.UsuarioID,
		Orgao:           params.Orgao,
		SEIUnidadeID:    unidade.ID,
		SEIUnidadeSigla: unidade.Sigla,
	}
	err = s.store.SaveAnalista(ctx, record)
	if err != nil {
		return nil, err
	}

	return mapAnalista(record), nil
}

// GetAnalista retorna os dados básicos de Analista para o ID de usuário informado.
func (s *Service) GetAnalista(ctx context.Context, usuarioID int64) (*Analista, error) {
	r, err := s.store.GetAnalista(ctx, usuarioID)
	if err != nil {
		return nil, err
	}
	return mapAnalista(r), nil
}

// AfastarAnalista marca um analista como afastado, não podendo receber novos processos.
func (s *Service) AfastarAnalista(ctx context.Context, usuarioID int64) error {
	r, err := s.store.GetAnalista(ctx, usuarioID)
	if err != nil {
		return err
	}

	r.Afastado = true
	return s.store.UpdateAnalista(ctx, r)
}

// RetornarAnalista marca um analista como não afastado, podendo receber novos processos.
func (s *Service) RetornarAnalista(ctx context.Context, usuarioID int64) error {
	r, err := s.store.GetAnalista(ctx, usuarioID)
	if err != nil {
		return err
	}

	r.Afastado = false
	return s.store.UpdateAnalista(ctx, r)
}

// ListAnalistas retorna os dados dos analistas da aplicação.
func (s *Service) ListAnalistas(ctx context.Context) ([]*Analista, error) {
	rr, err := s.store.ListAnalistas(ctx)
	if err != nil {
		return nil, err
	}

	aa := make([]*Analista, len(rr))
	for i, r := range rr {
		aa[i] = mapAnalista(r)
	}
	return aa, nil
}
