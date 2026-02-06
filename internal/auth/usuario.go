package auth

import (
	"context"
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
	ID              int64     `json:"id"`
	Nome            string    `json:"nome"`
	CPF             string    `json:"cpf"`
	Email           string    `json:"email"`
	EmailVerificado bool      `json:"email_verificado"`
	Papel           *string   `json:"papel"`
	CriadoEm        time.Time `json:"criado_em"`
	AtualizadoEm    time.Time `json:"atualizado_em"`
}

// IsAnonymous reporta se o usuário é anônimo (não autenticado).
func (u *Usuario) IsAnonymous() bool {
	return u == Anonymous
}

// HasPapel reporta se o usuário possui determinado papel.
func (u *Usuario) HasPapel(papel string) bool {
	if u.Papel == nil {
		return false
	}
	return *u.Papel == papel
}

// SetPapel define o papel do usuário.
func (u *Usuario) SetPapel(papel string) {
	u.Papel = &papel
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
func (s *Service) CreateUsuario(ctx context.Context, params CreateUsuarioParams) (*database.Usuario, error) {
	if !slices.Contains(AllowedPapeis, params.Papel) {
		return nil, ErrInvalidPapel
	}

	// Salva usuário no banco de dados.
	record := &database.Usuario{
		Nome:  params.Nome,
		CPF:   params.CPF,
		Email: params.Email,
	}
	record.SetPapel(params.Papel)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	err = store.SaveUsuario(ctx, record)
	if err != nil {
		return nil, err
	}

	// Enviar email de confirmação.
	if params.TokenURL != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			err := s.SendSetup(ctx, record, params.TokenURL)
			if err != nil {
				log.Printf("Não foi possível enviar email de cadastro")
			}
		}()
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return record, nil
}
