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

	"github.com/automatiza-mg/fila/internal/auth"
	"github.com/automatiza-mg/fila/internal/blob"
	"github.com/automatiza-mg/fila/internal/cache"
	"github.com/automatiza-mg/fila/internal/config"
	"github.com/automatiza-mg/fila/internal/database"
	"github.com/automatiza-mg/fila/internal/datalake"
	"github.com/automatiza-mg/fila/internal/docintel"
	"github.com/automatiza-mg/fila/internal/fila"
	"github.com/automatiza-mg/fila/internal/infra"
	"github.com/automatiza-mg/fila/internal/logging"
	"github.com/automatiza-mg/fila/internal/mail"
	"github.com/automatiza-mg/fila/internal/postgres"
	"github.com/automatiza-mg/fila/internal/processos"
	"github.com/automatiza-mg/fila/internal/sei"
	"github.com/automatiza-mg/fila/internal/tasks"
	"github.com/go-playground/form/v4"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/riverqueue/river"
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
	dev      bool
	cfg      *config.Config
	logger   *slog.Logger
	rdb      *redis.Client
	pool     *pgxpool.Pool
	store    *database.Store
	mail     mail.Sender
	cache    cache.Cache
	storage  blob.Storage
	datalake *datalake.DataLake
	di       *docintel.AzureDocIntel
	sei      *sei.Client
	queue    *river.Client[pgx.Tx]

	auth      *auth.Service
	fila      *fila.Service
	processos *processos.Service

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

	rdb, err := infra.NewRedis(ctx, cfg.RedisURL)
	if err != nil {
		return err
	}
	defer rdb.Close()

	pool, err := postgres.New(ctx, &cfg.Postgres)
	if err != nil {
		return err
	}
	defer pool.Close()

	storage, err := blob.New(ctx, &cfg.Blob)
	if err != nil {
		return err
	}
	defer storage.Close()

	dl, err := datalake.New(ctx, &cfg.DataLake)
	if err != nil {
		return err
	}
	defer dl.Close()

	sender, err := mail.NewSMTPSender(&cfg.Mail)
	if err != nil {
		return err
	}
	defer sender.Close()

	workers := river.NewWorkers()
	river.AddWorker(workers, tasks.NewSendEmailWorker(sender))

	queue, err := tasks.NewRiverClient(ctx, pool, workers)
	if err != nil {
		return err
	}
	if err := queue.Start(ctx); err != nil {
		return err
	}

	sei := sei.NewClient(&cfg.SEI)

	cache := cache.NewRedisCache(rdb)

	di := docintel.NewAzureDocIntel(&cfg.DocIntel)

	auth := auth.New(pool, logger, queue)

	fila := fila.New(pool, auth, sei, cache)
	if err := auth.RegisterHook(fila); err != nil {
		return err
	}

	proc := processos.New(&processos.ServiceOpts{
		Pool:  pool,
		Sei:   sei,
		Cache: cache,
		OCR:   di,
	})

	app := &application{
		dev:      *dev,
		cfg:      cfg,
		logger:   logger,
		rdb:      rdb,
		pool:     pool,
		store:    database.New(pool),
		mail:     sender,
		cache:    cache,
		storage:  storage,
		datalake: dl,
		sei:      sei,
		di:       di,

		fila:      fila,
		auth:      auth,
		processos: proc,

		decoder: form.NewDecoder(),
		views:   os.DirFS("web/views"),
		static:  os.DirFS("web/static"),
	}

	srv := &http.Server{
		Addr:         ":4000",
		Handler:      app.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 25 * time.Second,
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

	return errors.Join(
		srv.Shutdown(exitCtx),
		queue.Stop(exitCtx),
	)
}
