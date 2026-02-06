package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"slices"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
	"golang.org/x/crypto/bcrypt"
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

	hashSenha string
}

func mapUsuario(u *database.Usuario) *Usuario {
	return &Usuario{
		ID:              u.ID,
		Nome:            u.Nome,
		CPF:             u.CPF,
		Email:           u.Email,
		EmailVerificado: u.EmailVerificado,
		Papel:           u.Papel.V,

		hashSenha: u.HashSenha.V,
	}
}

// CheckSenha verifica se a senha informada equivale ao campo HashSenha do usuário.
// Caso o usuário não possua uma senha, retorna [ErrNoPassword].
func (u *Usuario) CheckSenha(senha string) (bool, error) {
	if !u.HasSenha() {
		return false, ErrNoPassword
	}

	err := bcrypt.CompareHashAndPassword([]byte(u.hashSenha), []byte(senha))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
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

// HasSenha reporta se o usuário possui uma senha cadastrada.
func (u *Usuario) HasSenha() bool {
	return u.hashSenha != ""
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

	u := mapUsuario(r)

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

	// Carrega pendências
	u.Pendencias = s.getPendingActions(ctx, u)

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
		u := mapUsuario(r)
		u.Pendencias = s.getPendingActions(ctx, u)

		usuarios[i] = u
	}
	return usuarios, nil
}

// GetUsuario retorna os dados de um usuário e carrega suas pendências, se houver.
func (s *Service) GetUsuario(ctx context.Context, usuarioID int64) (*Usuario, error) {
	r, err := s.store.GetUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}

	u := mapUsuario(r)
	u.Pendencias = s.getPendingActions(ctx, u)

	return u, nil
}

type UpdateUsuarioParams struct {
	// O identificador do usuário.
	UsuarioID int64
	// Nome completo do usuário.
	Nome string
	// O papel (role) do usuário. Deve ser um dos valores definidos em [AllowedPapeis].
	Papel string
}

// UpdateUsuario aplica as atualizações ao usuário. Caso ocorra mudança de papel,
// os métodos Cleanup dos providers registrados são chamados.
func (s *Service) UpdateUsuario(ctx context.Context, params UpdateUsuarioParams) error {
	if !slices.Contains(AllowedPapeis, params.Papel) {
		return ErrInvalidPapel
	}

	r, err := s.store.GetUsuario(ctx, params.UsuarioID)
	if err != nil {
		return err
	}

	u := mapUsuario(r)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	// Verifica se houve mudança de papel para limpeza em potencial de recursos.
	if r.Papel.Valid && r.Papel.V != params.Papel {
		if err := s.cleanupAll(ctx, tx, CleanupTriggerPapelUpdate, u); err != nil {
			return err
		}
	}

	// Atualiza dados do usuário.
	r.Nome = params.Nome
	r.Papel = sql.Null[string]{
		V:     params.Papel,
		Valid: true,
	}

	if err := store.UpdateUsuario(ctx, r); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// DeleteUsuario remove um usuário do banco de dados, executando as ações
// dos [CleanupProvider] registrados no serviço.
func (s *Service) DeleteUsuario(ctx context.Context, usuario *Usuario) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Executa a limpeza de recursos conforme necessário.
	if err := s.cleanupAll(ctx, tx, CleanupTriggerDelete, usuario); err != nil {
		return err
	}

	store := s.store.WithTx(tx)

	err = store.DeleteUsuario(ctx, usuario.ID)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
