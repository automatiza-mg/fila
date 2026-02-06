package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"slices"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
)

const (
	PapelAdmin         = "ADMIN"
	PapelSubsecretario = "SUBSECRETARIO"
	PapelGestor        = "GESTOR"
	PapelAnalista      = "ANALISTA"
)

var (
	// ErrInvalidPapel é o erro retornado durante a criação de um usuário com papel inválido. Ver [AllowedPapeis].
	ErrInvalidPapel = errors.New("invalid papel")

	// AllowedPapeis são os papeis permitidos para criação de novos usuários.
	AllowedPapeis = []string{
		PapelAnalista,
		PapelGestor,
		PapelSubsecretario,
	}

	// Anonymous representa um usuário não autenticado
	Anonymous = &Usuario{}
)

type Usuario struct {
	ID              int64           `json:"id"`
	Nome            string          `json:"nome"`
	CPF             string          `json:"cpf"`
	Email           string          `json:"email"`
	EmailVerificado bool            `json:"email_verificado"`
	Papel           string          `json:"papel,omitempty"`
	Pendencias      []PendingAction `json:"pendencias"`
}

// IsAnonymous reporta se o usuário é anônimo (não autenticado).
func (u *Usuario) IsAnonymous() bool {
	return u == Anonymous
}

// HasPapel reporta se o usuário possui determinado papel.
func (u *Usuario) HasPapel(papel string) bool {
	return u.Papel == papel
}

// IsAnalista verifica se o usuário é um analista.
func (u *Usuario) IsAnalista() bool {
	return u.Papel == PapelAnalista
}

type CreateUsuarioParams struct {
	// Nome completo do novo usuário.
	Nome string
	// O cadastro de pessoa física do usuário. Deve ser estar formatado (com pontos e traços).
	CPF string
	// Email de contato do novo usuário. As notificações serão enviadas para esse email.
	Email string
	// O papel (role) do novo usuário. Deve ser um dos valores definidos em [AllowedPapeis].
	Papel string
	// TokenURL é a função que gera um URL de destino para o token de cadastro.
	// Se o valor não for nil, um email de cadastro será enviado ao novo usuário.
	TokenURL func(token string) string
}

// CreateUsuario cria um novo usuário no sistema para os dados informados. A criação de usuários admin
// não é permitida por esse método.
func (s *Service) CreateUsuario(ctx context.Context, params CreateUsuarioParams) (*Usuario, error) {
	if !slices.Contains(AllowedPapeis, params.Papel) {
		return nil, ErrInvalidPapel
	}

	// Salva usuário no banco de dados.
	r := &database.Usuario{
		Nome:  params.Nome,
		CPF:   params.CPF,
		Email: params.Email,
		Papel: sql.Null[string]{
			V:     params.Papel,
			Valid: true,
		},
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	err = store.SaveUsuario(ctx, r)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Mapeia registro no banco para um Usuario.
	u := &Usuario{
		ID:              r.ID,
		Nome:            r.Nome,
		CPF:             r.CPF,
		Email:           r.Email,
		EmailVerificado: r.EmailVerificado,
		Papel:           r.Papel.V,
	}

	// Enviar email de confirmação.
	if params.TokenURL != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			err := s.SendSetup(ctx, u, params.TokenURL)
			if err != nil {
				log.Printf("Não foi possível enviar email de cadastro")
			}
		}()
	}

	// Carregar pendências.
	u.Pendencias, err = s.GetActions(ctx, u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

type ListUsuariosParams struct {
	Papel string
}

// ListUsuarios retorna a lista de usuários com os dados de pendências carregados.
func (s *Service) ListUsuarios(ctx context.Context, params ListUsuariosParams) ([]*Usuario, error) {
	records, _, err := s.store.ListUsuarios(ctx, database.ListUsuariosParams{
		Papel: params.Papel,
	})
	if err != nil {
		return nil, err
	}

	usuarios := make([]*Usuario, len(records))
	for i, r := range records {
		u := &Usuario{
			ID:              r.ID,
			Nome:            r.Nome,
			CPF:             r.CPF,
			Email:           r.Email,
			EmailVerificado: r.EmailVerificado,
			Papel:           r.Papel.V,
		}

		u.Pendencias, err = s.GetActions(ctx, u)
		if err != nil {
			return nil, err
		}

		usuarios[i] = u
	}
	return usuarios, nil
}

func (s *Service) GetUsuario(ctx context.Context, usuarioID int64) (*Usuario, error) {
	r, err := s.store.GetUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}

	u := &Usuario{
		ID:              r.ID,
		Nome:            r.Nome,
		CPF:             r.CPF,
		Email:           r.Email,
		EmailVerificado: r.EmailVerificado,
		Papel:           r.Papel.V,
	}

	u.Pendencias, err = s.GetActions(ctx, u)
	if err != nil {
		return nil, err
	}

	return u, nil
}
