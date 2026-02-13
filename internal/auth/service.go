package auth

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/mail"
	"github.com/automatiza-mg/fila/internal/tasks"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"golang.org/x/crypto/bcrypt"
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

type TaskInserter interface {
	InsertTx(ctx context.Context, tx pgx.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error)
}

type Service struct {
	pool   *pgxpool.Pool
	store  *database.Store
	logger *slog.Logger
	queue  TaskInserter

	hooks map[string]UsuarioHook
}

func New(pool *pgxpool.Pool, logger *slog.Logger, queue TaskInserter) *Service {
	return &Service{
		pool:   pool,
		store:  database.New(pool),
		logger: logger.With(slog.String("service", "auth")),
		queue:  queue,

		hooks: make(map[string]UsuarioHook),
	}
}

// RegisterProvider registra um novo [UsuarioHook] no serviço.
// Tentativa de registro de providers com o mesmo Label serão ignoradas.
func (s *Service) RegisterHook(h UsuarioHook) error {
	label := h.Label()

	if _, ok := s.hooks[label]; ok {
		return ErrDuplicateProvider
	}

	s.hooks[label] = h
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

	// Carrega os dados do usuário.
	u := MapUsuario(record)
	u.Pendencias = s.getPendingActions(ctx, u)

	// Verifica se o usuário possui uma senha.
	if !u.HasSenha() {
		return nil, ErrNoPassword
	}

	// Compara o hash com a senha.
	err = bcrypt.CompareHashAndPassword([]byte(u.hashSenha), []byte(senha))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	return u, nil
}

type SetupUsuarioParams struct {
	Token string
	Senha string
}

func (s *Service) SetupUsuario(ctx context.Context, params SetupUsuarioParams) error {
	r, err := s.store.GetUsuarioForToken(ctx, params.Token, EscopoSetup.String())
	if err != nil {
		switch {
		case errors.Is(err, database.ErrNotFound):
			return ErrInvalidToken
		default:
			return err
		}
	}

	if r.EmailVerificado {
		return ErrAlreadySetup
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(params.Senha), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	r.EmailVerificado = true
	r.HashSenha = sql.Null[string]{
		V:     string(hash),
		Valid: true,
	}
	err = store.UpdateUsuario(ctx, r)
	if err != nil {
		return err
	}

	err = store.DeleteTokensUsuario(ctx, r.ID, EscopoSetup.String())
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
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

	u := MapUsuario(r)
	u.Pendencias = s.getPendingActions(ctx, u)

	return u, nil
}

// SendSetup envia um novo email de verificação para o usuário. Caso o usuário
// já esteja verificado, retorna [ErrAlreadySetup].
func (s *Service) SendSetup(ctx context.Context, usuario *Usuario, tokenFn func(token string) string) error {
	if usuario.EmailVerificado {
		return ErrAlreadySetup
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	token, err := s.createToken(ctx, store, usuario.ID, EscopoSetup, 72*time.Hour)
	if err != nil {
		return err
	}

	email, err := mail.NewSetupEmail(usuario.Email, mail.SetupEmailParams{
		SetupURL: tokenFn(token.Token),
	})
	if err != nil {
		return err
	}

	_, err = s.queue.InsertTx(ctx, tx, tasks.SendEmailArgs{
		Email: email,
	}, nil)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// SendResetSenha envia um email de recuperação de senha para o usuário identificado pelo CPF.
// Retorna nil silenciosamente caso o usuário não exista ou não tenha email verificado,
// evitando enumeração de contas.
func (s *Service) SendResetSenha(ctx context.Context, cpf string, tokenFn func(token string) string) error {
	r, err := s.store.GetUsuarioByCPF(ctx, cpf)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil
		}
		return err
	}

	if !r.EmailVerificado {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	err = store.DeleteTokensUsuario(ctx, r.ID, EscopoResetSenha.String())
	if err != nil {
		return err
	}

	token, err := s.createToken(ctx, store, r.ID, EscopoResetSenha, 1*time.Hour)
	if err != nil {
		return err
	}

	email, err := mail.NewResetSenhaEmail(r.Email, mail.ResetSenhaEmailParams{
		ResetURL: tokenFn(token.Token),
	})
	if err != nil {
		return err
	}

	_, err = s.queue.InsertTx(ctx, tx, tasks.SendEmailArgs{
		Email: email,
	}, nil)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ResetSenhaParams são os parâmetros para redefinição de senha.
type ResetSenhaParams struct {
	Token string
	Senha string
}

// ResetSenha redefine a senha do usuário utilizando um token de recuperação válido.
// Retorna [ErrInvalidToken] se o token for inválido ou expirado.
func (s *Service) ResetSenha(ctx context.Context, params ResetSenhaParams) error {
	r, err := s.store.GetUsuarioForToken(ctx, params.Token, EscopoResetSenha.String())
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return ErrInvalidToken
		}
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(params.Senha), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	store := s.store.WithTx(tx)

	err = store.UpdateUsuarioSenha(ctx, r.ID, string(hash))
	if err != nil {
		return err
	}

	err = store.DeleteTokensUsuario(ctx, r.ID, EscopoResetSenha.String())
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
