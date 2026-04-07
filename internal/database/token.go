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

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UsuarioID int64     `json:"-"`
	Escopo    string    `json:"-"`
	ExpiraEm  time.Time `json:"expira_em"`
}

func hashToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}

// CreateToken gera um novo token com os dados informados e salva no banco de dados.
func (s *Store) CreateToken(ctx context.Context, usuarioID int64, escopo string, ttl time.Duration) (*Token, error) {
	b := make([]byte, 32)
	_, _ = rand.Read(b)

	plaintext := base64.RawURLEncoding.EncodeToString(b)

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

func (s *Store) SaveToken(ctx context.Context, token *Token) error {
	q := `INSERT INTO tokens (hash, usuario_id, escopo, expira_em) VALUES ($1, $2, $3, $4)`
	args := []any{token.Hash, token.UsuarioID, token.Escopo, token.ExpiraEm}
	_, err := s.db.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	return nil
}

// GetUsuarioIDForToken retorna o ID do usuário dono de um token válido.
// Retorna [ErrNotFound] caso o token seja inválido ou tenha expirado.
func (s *Store) GetUsuarioIDForToken(ctx context.Context, token string, escopo string) (int64, error) {
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
			return 0, ErrNotFound
		}
		return 0, err
	}

	return usuarioID, nil
}

// DeleteTokensUsuario exclui todos os tokens de um usuário com determinado escopo do banco de dados.
func (s *Store) DeleteTokensUsuario(ctx context.Context, usuarioID int64, escopo string) error {
	q := `DELETE FROM tokens WHERE usuario_id = $1 AND escopo = $2`
	_, err := s.db.Exec(ctx, q, usuarioID, escopo)
	if err != nil {
		return err
	}
	return nil
}

// DeleteToken remove um token pelo hash informado.
func (s *Store) DeleteToken(ctx context.Context, hash []byte) error {
	q := `DELETE FROM tokens WHERE hash = $1`
	_, err := s.db.Exec(ctx, q, hash)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteExpiredTokens(ctx context.Context) (int64, error) {
	q := `DELETE FROM tokens WHERE expira_em < CURRENT_TIMESTAMP`
	res, err := s.db.Exec(ctx, q)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}
