package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	// ErrUsuarioCPFTaken é o erro retornado ao tentar salvar um usuário com CPF duplicado.
	ErrUsuarioCPFTaken = errors.New("duplicate user cpf")
	// ErrUsuarioEmailTaken é o erro retornado ao tentar salvar um usuário com Email duplicado.
	ErrUsuarioEmailTaken = errors.New("duplicate user email")
	// ErrNoPassword é o erro retornado quando não é possível verificar a senha de um usuário que não possui uma.
	ErrNoPassword = errors.New("user has no password")
)

const (
	PapelAdmin         = "ADMIN"
	PapelSubsecretario = "SUBSECRETARIO"
	PapelGestor        = "GESTOR"
	PapelAnalista      = "ANALISTA"
)

// Anonymous representa um usuário não autenticado
var Anonymous = &Usuario{}

type Usuario struct {
	ID              int64
	Nome            string
	CPF             string
	Email           string
	EmailVerificado bool
	HashSenha       sql.Null[string]
	Papel           sql.Null[string]
	CriadoEm        time.Time
	AtualizadoEm    time.Time
}

// SaveUsuario adiciona o usuário ao banco de dados. Retorna [ErrUsuarioCPFTaken] e [ErrUsuarioEmailTaken]
// no caso de campos duplicados.
func (s *Store) SaveUsuario(ctx context.Context, usuario *Usuario) error {
	q := `
	INSERT INTO usuarios (nome, cpf, email, email_verificado, hash_senha, papel)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, criado_em, atualizado_em`

	args := []any{usuario.Nome, usuario.CPF, usuario.Email, usuario.EmailVerificado, usuario.HashSenha, usuario.Papel}

	err := s.db.QueryRow(ctx, q, args...).Scan(
		&usuario.ID,
		&usuario.CriadoEm,
		&usuario.AtualizadoEm,
	)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "usuarios_email_key"):
			return ErrUsuarioEmailTaken
		case strings.Contains(err.Error(), "usuarios_cpf_key"):
			return ErrUsuarioCPFTaken
		default:
			return err
		}
	}
	return nil
}

// GetUsuario retorna um usuário do banco de dados pelo ID.
// Retorna [ErrNotFound] se nenhum usuário for encontrado.
func (s *Store) GetUsuario(ctx context.Context, usuarioID int64) (*Usuario, error) {
	q := `
	SELECT 
		id, nome, cpf, email, email_verificado,
		hash_senha, papel, criado_em, atualizado_em
	FROM usuarios
	WHERE id = $1`

	var usuario Usuario
	err := s.db.QueryRow(ctx, q, usuarioID).Scan(
		&usuario.ID,
		&usuario.Nome,
		&usuario.CPF,
		&usuario.Email,
		&usuario.EmailVerificado,
		&usuario.HashSenha,
		&usuario.Papel,
		&usuario.CriadoEm,
		&usuario.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &usuario, nil
}

// GetUsuarioByCPF retorna um usuário do banco de dados pelo CPF.
// Retorna [ErrNotFound] se nenhum usuário for encontrado.
func (s *Store) GetUsuarioByCPF(ctx context.Context, cpf string) (*Usuario, error) {
	q := `
	SELECT 
		id, nome, cpf, email, email_verificado,
		hash_senha, papel, criado_em, atualizado_em
	FROM usuarios
	WHERE cpf = $1`

	var usuario Usuario
	err := s.db.QueryRow(ctx, q, cpf).Scan(
		&usuario.ID,
		&usuario.Nome,
		&usuario.CPF,
		&usuario.Email,
		&usuario.EmailVerificado,
		&usuario.HashSenha,
		&usuario.Papel,
		&usuario.CriadoEm,
		&usuario.AtualizadoEm,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &usuario, nil
}

type ListUsuariosParams struct {
	Papel           string
	EmailVerificado sql.Null[bool]
}

// ListUsuarios retorna a lista de usuários e o total de resultados com os filtros aplicados.
func (s *Store) ListUsuarios(ctx context.Context, params ListUsuariosParams) ([]*Usuario, int, error) {
	query := `
	SELECT 
		id, nome, cpf, email, email_verificado,
		hash_senha, papel, criado_em, atualizado_em,
		COUNT(*) OVER()
	FROM usuarios
	WHERE (papel = $1 OR $1 = '')
	AND (email_verificado = $2 OR $2 IS NULL)`

	rows, err := s.db.Query(ctx, query, params.Papel, params.EmailVerificado)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var totalCount int
	usuarios := make([]*Usuario, 0)
	for rows.Next() {
		var usuario Usuario
		err := rows.Scan(
			&usuario.ID,
			&usuario.Nome,
			&usuario.CPF,
			&usuario.Email,
			&usuario.EmailVerificado,
			&usuario.HashSenha,
			&usuario.Papel,
			&usuario.CriadoEm,
			&usuario.AtualizadoEm,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}
		usuarios = append(usuarios, &usuario)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return usuarios, totalCount, nil
}

// UpdateUsuario atualiza os dados do usuário no banco de dados.
func (s *Store) UpdateUsuario(ctx context.Context, usuario *Usuario) error {
	q := `
	UPDATE usuarios SET
		nome = $2,
		email_verificado = $3,
		hash_senha = $4,
		papel = $5,
		atualizado_em = CURRENT_TIMESTAMP
	WHERE id = $1
	RETURNING atualizado_em`
	args := []any{usuario.ID, usuario.Nome, usuario.EmailVerificado, usuario.HashSenha, usuario.Papel}

	err := s.db.QueryRow(ctx, q, args...).Scan(&usuario.AtualizadoEm)
	if err != nil {
		return err
	}
	return nil
}

// IsUsuariosEmpty reporta se a tabela 'usuarios' está vazia (não possui nenhum registro).
func (s *Store) IsUsuariosEmpty(ctx context.Context) (bool, error) {
	q := `SELECT NOT EXISTS (SELECT 1 FROM usuarios LIMIT 1)`

	var empty bool
	err := s.db.QueryRow(ctx, q).Scan(&empty)
	if err != nil {
		return false, err
	}

	return empty, nil
}

func (s *Store) DeleteUsuario(ctx context.Context, id int64) error {
	q := `DELETE FROM usuarios WHERE id = $1`
	_, err := s.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	return nil
}
