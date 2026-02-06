package tasks

import (
	"context"

	"github.com/automatiza-mg/fila/internal/mail"
	"github.com/riverqueue/river"
)

type SendEmailArgs struct {
	Email *mail.Email
}

func (args SendEmailArgs) Kind() string {
	return "send:email"
}

type SendEmailWorker struct {
	sender mail.Sender
	river.WorkerDefaults[SendEmailArgs]
}

func NewSendEmailWorker(sender mail.Sender) *SendEmailWorker {
	return &SendEmailWorker{
		sender: sender,
	}
}

func (w *SendEmailWorker) Work(ctx context.Context, job *river.Job[SendEmailArgs]) error {
	return w.sender.Send(ctx, job.Args.Email)
}
