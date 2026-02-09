package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/automatiza-mg/fila/internal/auth"
	"github.com/automatiza-mg/fila/internal/config"
	"github.com/automatiza-mg/fila/internal/logging"
	"github.com/automatiza-mg/fila/internal/postgres"
	"github.com/automatiza-mg/fila/internal/tasks"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

func main() {
	_ = godotenv.Load()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.NewFromEnv()
	if err != nil {
		return err
	}

	pool, err := postgres.New(context.Background(), &cfg.Postgres)
	if err != nil {
		return err
	}

	logger := logging.NewLogger(os.Stdout, false)

	queue, err := tasks.NewRiverClient(context.Background(), pool, nil)
	if err != nil {
		return err
	}

	a := auth.New(pool, logger, queue)

	cmd := &cli.Command{
		Name:  "fila",
		Usage: "Executa tarefas administrativas da Fila de Aposentadoria",
		Commands: []*cli.Command{
			{
				Name:  "create-admin",
				Usage: "Adiciona um novo administrador da aplicação",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "nome",
						Aliases:  []string{"n"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "cpf",
						Aliases:  []string{"c"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "email",
						Aliases:  []string{"e"},
						Required: true,
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {

					nome := c.String("nome")
					cpf := c.String("cpf")
					email := c.String("email")

					fmt.Printf("Digite a senha: ")
					senha, err := term.ReadPassword(syscall.Stdin)
					if err != nil {
						return err
					}
					fmt.Println()

					u, err := a.CreateAdmin(ctx, auth.CreateAdminParams{
						Nome:  nome,
						CPF:   cpf,
						Email: email,
						Senha: string(senha),
					})
					if err != nil {
						return err
					}

					fmt.Printf("Admin criado para %s\n", u.Nome)
					return nil
				},
			},
		},
	}

	return cmd.Run(context.Background(), os.Args)
}
