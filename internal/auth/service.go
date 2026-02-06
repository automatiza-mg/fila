package auth

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/mail"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrAlreadySetup é o erro retornado em tentativa de emails de verificação
	// para usuários já verificados.
	ErrAlreadySetup = errors.New("usuario is already setup")
	// ErrNoPassword é o erro retornado quando não é possível verificar a senha
	// de um usuário que não possui uma.
	ErrNoPassword = errors.New("usuario has no password")
	// ErrDuplicateProvider é o erro retornado quando há tentativa de registro
	// de providers com o mesmo label.
	ErrDuplicateProvider = errors.New("duplicate lifecycle provider")
	// ErrInvalidCredentials é o erro retornado quando há uma tentativa
	// de login com credenciais inválidas
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type LifecycleProvider interface {
	// Label retorna um ID do provider registrado.
	Label() string
	// GetActions retorna pendenciais específicas de um usuário.
	GetActions(ctx context.Context, u *Usuario) ([]PendingAction, error)
	// Cleanup executa ações de limpeza durante a exclusão ou alteração de papel
	// de um usuário.
	Cleanup(ctx context.Context, tx pgx.Tx, usuario *Usuario) error
}

type Service struct {
	pool   *pgxpool.Pool
	store  *database.Store
	logger *slog.Logger
	sender mail.Sender

	providers map[string]LifecycleProvider
}

func New(pool *pgxpool.Pool, logger *slog.Logger, sender mail.Sender) *Service {
	return &Service{
		pool:      pool,
		store:     database.New(pool),
		logger:    logger.With(slog.String("service", "auth")),
		sender:    sender,
		providers: make(map[string]LifecycleProvider),
	}
}

// RegisterProvider registra um novo [LifecycleProvider] no serviço.
// Tentativa de registro de providers com o mesmo Label serão ignoradas.
func (s *Service) RegisterProvider(p LifecycleProvider) error {
	label := p.Label()

	if _, ok := s.providers[label]; ok {
		return ErrDuplicateProvider
	}

	s.providers[label] = p
	s.logger.Debug("Provider registrado", slog.String("provider", label))
	return nil
}

// Authenticate retorna um usuário caso as credenciais informadas sejam válidas.
// Se o CPF ou Senha estiverem incorretos, retorn [ErrInvalidCredentials].
func (s *Service) Authenticate(ctx context.Context, cpf, senha string) (*Usuario, error) {
	record, err := s.store.GetUsuarioByCPF(ctx, cpf)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	ok, err := record.CheckSenha(senha)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidCredentials
	}

	u := mapUsuario(record)
	u.Pendencias = s.getPendingActions(ctx, u)

	return u, nil
}

// GetTokenOwner retorna o usuário dono de um token com o escopo especificado.
// Retorna [ErrInvalidToken] caso o token seja inválido ou tenha expirado.
func (s *Service) GetTokenOwner(ctx context.Context, token string, escopo Escopo) (*Usuario, error) {
	r, err := s.store.GetUsuarioForToken(ctx, token, escopo.String())
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			return nil, ErrInvalidToken
		default:
			return nil, err
		}
	}

	u := mapUsuario(r)
	u.Pendencias = s.getPendingActions(ctx, u)

	return u, nil
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
