package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
)

const (
	// EscopoSetup é o escopo usado para finalizar o cadastro de um usuário.
	EscopoSetup Escopo = "setup"
	// EscopoResetSenha é o escopo usado para redefinir a senha de um usuário.
	EscopoResetSenha Escopo = "reset-senha"
	// EscopoAuth é o escopo usado para autenticar um usuário.
	EscopoAuth Escopo = "auth"

	// Tamanho em bytes do token gerado.
	tokenSize = 32
)

var (
	// ErrInvalidToken é o erro retornado para tokens que são inválidos ou expiraram (não encontrado).
	ErrInvalidToken = errors.New("invalid or expired token")
)

// Escopo representa a finalidade de um token, como reset de senha, conclusão do cadastro, etc.
type Escopo string

func (e Escopo) String() string {
	return string(e)
}

type Token struct {
	Token  string    `json:"token"`
	Expira time.Time `json:"expira"`
}

// Gera o hash do token usando sha256.
func hashToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}

func (s *Service) createToken(ctx context.Context, store *database.Store, usuarioID int64, escopo Escopo, ttl time.Duration) (*Token, error) {
	b := make([]byte, tokenSize)
	_, _ = rand.Read(b)

	plaintext := base64.RawURLEncoding.EncodeToString(b)
	hash := hashToken(plaintext)
	expira := time.Now().Add(ttl)

	err := store.SaveToken(ctx, &database.Token{
		Hash:      hash,
		UsuarioID: usuarioID,
		Escopo:    escopo.String(),
		ExpiraEm:  expira,
	})
	if err != nil {
		return nil, err
	}

	return &Token{
		Token:  plaintext,
		Expira: expira,
	}, nil
}

// CreateToken cria um novo token de acesso para o usuário e finalidade especificada.
func (s *Service) CreateToken(ctx context.Context, usuarioID int64, escopo Escopo, ttl time.Duration) (*Token, error) {
	return s.createToken(ctx, s.store, usuarioID, escopo, ttl)
}

// DeleteToken remove um único token do banco de dados.
func (s *Service) DeleteToken(ctx context.Context, token string) error {
	return s.store.DeleteToken(ctx, hashToken(token))
}
