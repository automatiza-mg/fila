package mail

import "context"

type Email struct {
	To      []string
	Subject string
	Text    string
	HTML    string
}

// Sender é uma interface mínima para o envio de emails.
type Sender interface {
	Send(ctx context.Context, email Email) error
}
