package auth

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/mail"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrAlreadySetup é o erro retornado em tentativa de emails de verificação
	// para usuários já verificados.
	ErrAlreadySetup = errors.New("usuario is already setup")
	// ErrNoPassword é o erro retornado quando não é possível verificar a senha
	// de um usuário que não possui uma.
	ErrNoPassword = errors.New("usuario has no password")
)

type Service struct {
	pool   *pgxpool.Pool
	store  *database.Store
	sender mail.Sender

	providers []ActionProvider
}

func New(pool *pgxpool.Pool, sender mail.Sender, providers ...ActionProvider) *Service {
	return &Service{
		pool:   pool,
		store:  database.New(pool),
		sender: sender,

		providers: providers,
	}
}

// SendSetup envia um novo email de verificação para o usuário. Caso o usuário
// já esteja verificado, retorna [ErrAlreadySetup].
func (s *Service) SendSetup(ctx context.Context, usuario *Usuario, tokenFn func(token string) string) error {
	if usuario.EmailVerificado {
		return ErrAlreadySetup
	}

	token, err := s.createToken(ctx, s.store, usuario.ID, EscopoSetup, 72*time.Hour)
	if err != nil {
		return err
	}

	email, err := mail.NewSetupEmail(usuario.Email, mail.SetupEmailParams{
		SetupURL: tokenFn(token.Token),
	})

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err := s.sender.Send(ctx, email)
		if err != nil {
			log.Printf("Não foi possível enviar email: %v", err)
		}
	}()

	return nil
}
