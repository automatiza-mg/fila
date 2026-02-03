package database

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	// ErrInvalidToken é o erro retornado para tokens que são inválidos ou expiraram (não encontrado).
	ErrInvalidToken = errors.New("invalid or expired token")
)

type Escopo string

const (
	// EscopoSetup é o escopo usado para finalizar o cadastro de um usuário.
	EscopoSetup Escopo = "setup"
	// EscopoResetSenha é o escopo usado para redefinir a senha de um usuário.
	EscopoResetSenha Escopo = "reset-senha"
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UsuarioID int64     `json:"-"`
	Escopo    Escopo    `json:"-"`
	ExpiraEm  time.Time `json:"expira_em"`
}

func hashToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}

// CreateToken gera um novo token com os dados informados e salva no banco de dados.
func (s *Store) CreateToken(ctx context.Context, usuarioID int64, escopo Escopo, ttl time.Duration) (*Token, error) {
	b := make([]byte, 32)
	_, _ = rand.Read(b)

	plaintext := base64.URLEncoding.EncodeToString(b)

	token := &Token{
		Plaintext: plaintext,
		Hash:      hashToken(plaintext),
		UsuarioID: usuarioID,
		Escopo:    escopo,
		ExpiraEm:  time.Now().Add(ttl),
	}

	q := `INSERT INTO tokens (hash, usuario_id, escopo, expira_em) VALUES ($1, $2, $3, $4)`
	args := []any{token.Hash, token.UsuarioID, token.Escopo, token.ExpiraEm}

	_, err := s.db.Exec(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// GetUsuarioForToken retorna um usuário dono de um token válido. Caso o token não exista ou tenha expirado,
// retorna [ErrInvalidToken].
func (s *Store) GetUsuarioForToken(ctx context.Context, token string, escopo Escopo) (*Usuario, error) {
	q := `
	SELECT usuario_id
	FROM tokens
	WHERE hash = $1
	AND escopo = $2
	AND expira_em > CURRENT_TIMESTAMP`

	var usuarioID int64
	err := s.db.QueryRow(ctx, q, hashToken(token), escopo).Scan(&usuarioID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	usuario, err := s.GetUsuario(ctx, usuarioID)
	if err != nil {
		return nil, err
	}
	return usuario, nil
}

// DeleteTokensUsuario exclui todos os tokens de um usuário com determinado escopo do banco de dados.
func (s *Store) DeleteTokensUsuario(ctx context.Context, usuarioID int64, escopo Escopo) error {
	q := `DELETE FROM tokens WHERE usuario_id = $1 AND escopo = $2`
	_, err := s.db.Exec(ctx, q, usuarioID, escopo)
	if err != nil {
		return err
	}
	return nil
}
