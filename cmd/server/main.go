package main

import (
	"context"
	"errors"
	"flag"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/automatiza-mg/fila/internal/config"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/logging"
	"github.com/automatiza-mg/fila/internal/mail"
	"github.com/automatiza-mg/fila/internal/postgres"
	"github.com/go-playground/form/v4"
	"github.com/jackc/pgx/v5/pgxpool"
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

type application struct {
	cfg    *config.Config
	logger *slog.Logger
	pool   *pgxpool.Pool
	store  *database.Store
	mail   mail.Sender

	decoder *form.Decoder
	views   fs.FS
	static  fs.FS
}

func run(ctx context.Context) error {
	dev := flag.Bool("dev", false, "Executa a aplicação em modo de desenvolvimento")
	flag.Parse()

	cfg, err := config.NewFromEnv()
	if err != nil {
		return err
	}

	logger := logging.NewLogger(os.Stdout, *dev)

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

	app := &application{
		cfg:    cfg,
		logger: logger,
		pool:   pool,
		store:  database.New(pool),
		mail:   sender,

		decoder: form.NewDecoder(),
		views:   os.DirFS("web/views"),
		static:  os.DirFS("web/static"),
	}

	srv := &http.Server{
		Addr:         ":4000",
		Handler:      app.routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  time.Minute,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	go func() {
		logger.Info("Iniciando servidor HTTP", slog.String("addr", srv.Addr))
		err := srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	exitCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	logger.Info("Encerrando servidor HTTP", slog.String("addr", srv.Addr))
	err = errors.Join(
		srv.Shutdown(exitCtx),
	)
	return err
}
