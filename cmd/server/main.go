package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/automatiza-mg/fila/internal/config"
	"github.com/automatiza-mg/fila/internal/mail"
	"github.com/automatiza-mg/fila/internal/postgres"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.NewFromEnv()
	if err != nil {
		return err
	}

	pool, err := postgres.New(ctx, &cfg.Postgres)
	if err != nil {
		return err
	}
	defer pool.Close()

	sender, err := mail.NewSMTPSender(&cfg.Mail)
	if err != nil {
		return err
	}
	defer sender.Close()

	<-ctx.Done()

	return nil
}
