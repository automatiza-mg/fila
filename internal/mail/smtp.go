package mail

import (
	"context"
	"errors"
	"time"

	"github.com/wneessen/go-mail"
)

const (
	devPort = 1025
)

var (
	ErrEmptyText = errors.New("email text is empty")
)

var _ Sender = (*SMTPSender)(nil)

type SMTPSender struct {
	fromAddress string
	client      *mail.Client
}

func NewSMTPSender(cfg *Config) (*SMTPSender, error) {
	var opts []mail.Option
	if cfg.Port == devPort {
		opts = append(opts, mail.WithPort(cfg.Port), mail.WithTLSPolicy(mail.NoTLS))
	} else {
		opts = append(opts, mail.WithTLSPortPolicy(mail.TLSMandatory))
	}

	client, err := mail.NewClient(cfg.Host, opts...)
	if err != nil {
		return nil, err
	}
	if cfg.User != "" && cfg.Password != "" {
		client.SetSMTPAuth(mail.SMTPAuthAutoDiscover)
		client.SetUsername(cfg.User)
		client.SetPassword(cfg.Password)
	}

	return &SMTPSender{
		fromAddress: cfg.FromAddress,
		client:      client,
	}, nil
}

// Send envia um Email usando um servidor SMTP. O campo Email.Text nunca deve estar em branco e
// retorna [ErrEmptyText] se estiver. O campo Email.HTML é sempre usado como alternativa.
func (s *SMTPSender) Send(ctx context.Context, email Email) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	msg := mail.NewMsg()
	msg.Subject(email.Subject)

	err := msg.From(s.fromAddress)
	if err != nil {
		return err
	}

	err = msg.To(email.To...)
	if err != nil {
		return err
	}

	if email.Text == "" {
		return ErrEmptyText
	}

	msg.SetBodyString(mail.TypeTextPlain, email.Text)
	if email.HTML != "" {
		msg.AddAlternativeString(mail.TypeTextHTML, email.HTML)
	}

	return s.client.DialAndSendWithContext(ctx, msg)
}

// Close encerra o client do servidor STMP.
func (s *SMTPSender) Close() error {
	return s.client.Close()
}
